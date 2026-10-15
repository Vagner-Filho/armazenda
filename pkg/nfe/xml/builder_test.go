package xml_test

import (
	"strings"
	"testing"
	"time"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/xml"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

// findElement is a fatal-on-miss helper for assertions that navigate the
// built document by an etree path.
func findElement(t *testing.T, root *etree.Element, path string) *etree.Element {
	t.Helper()
	el := root.FindElement(path)
	if el == nil {
		t.Fatalf("element %q not found in built XML", path)
	}
	return el
}

// childTags returns the direct child tag names of el, in document order.
func childTags(el *etree.Element) []string {
	tags := make([]string, 0, len(el.ChildElements()))
	for _, c := range el.ChildElements() {
		tags = append(tags, c.Tag)
	}
	return tags
}

// assertChildOrder fails the test unless el's direct children match want.
func assertChildOrder(t *testing.T, el *etree.Element, what string, want []string) {
	t.Helper()
	got := childTags(el)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s child order = %v, want %v", what, got, want)
	}
}

func TestBuilder_FreeTextFieldsSanitized(t *testing.T) {
	builder := xml.NewBuilder()

	t.Run("infCpl_with_newlines_and_emoji", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.InformacoesAdicionais = "CND Fazenda:\nCert. Nº 222222, válido até 02/08/2026 🚀"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		xmlStr, _ := doc.WriteToString()

		if strings.Contains(xmlStr, "\n") && strings.Contains(xmlStr, "CND Fazenda:\n") {
			t.Error("infCpl must not contain newlines (SEFAZ rejects with cvc-type.3.1.3)")
		}
		if !strings.Contains(xmlStr, "<infCpl>CND Fazenda:; Cert. No 222222, valido ate 02/08/2026</infCpl>") {
			t.Errorf("unexpected infCpl content in XML: %s", xmlStr)
		}
		if strings.Contains(xmlStr, "🚀") {
			t.Error("characters above U+00FF must be dropped from infCpl")
		}
	})

	t.Run("recipient_name_with_newline", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Environment = 1 // production keeps the recipient name as-is
		input.Recipient.XNome = "Armazém\nTropical 🌾 Ltda"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		xmlStr, _ := doc.WriteToString()
		if !strings.Contains(xmlStr, "<xNome>Armazem; Tropical  Ltda</xNome>") {
			t.Errorf("unexpected xNome content in XML: %s", xmlStr)
		}
	})
}

func TestBuilder_ContingencyFields(t *testing.T) {
	builder := xml.NewBuilder()
	now := time.Now()

	t.Run("normal_emission_no_contingency_fields", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.TpEmis = defaults.EmissaoNormal
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		xmlStr, _ := doc.WriteToString()
		if strings.Contains(xmlStr, "<dhCont>") {
			t.Error("normal emission should NOT contain <dhCont>")
		}
		if strings.Contains(xmlStr, "<xJust>") {
			t.Error("normal emission should NOT contain <xJust>")
		}
	})

	t.Run("svc_contingency_has_dhCont_and_xJust", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.TpEmis = defaults.SVCRS
		input.DhCont = &now
		input.XJust = "Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		xmlStr, _ := doc.WriteToString()
		if !strings.Contains(xmlStr, "<dhCont>") {
			t.Error("SVC contingency should contain <dhCont>")
		}
		if !strings.Contains(xmlStr, "<xJust>") {
			t.Error("SVC contingency should contain <xJust>")
		}
		if !strings.Contains(xmlStr, "<tpEmis>7</tpEmis>") {
			t.Error("SVC contingency should have tpEmis=7")
		}
	})

	t.Run("access_key_contains_tpEmis", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.TpEmis = defaults.SVCAN
		input.CNF = "12345678"
		input.Numero = 99
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		xmlStr, _ := doc.WriteToString()
		// Find the infNFe Id attribute which contains the access key
		if !strings.Contains(xmlStr, `Id="NFe`) {
			t.Fatal("access key not found in XML")
		}
		// Extract the access key from the Id attribute (NFe + 44 chars)
		idx := strings.Index(xmlStr, `Id="NFe`)
		if idx < 0 {
			t.Fatal("Id attribute not found")
		}
		start := idx + len(`Id="NFe`)
		key := xmlStr[start : start+44]
		if key[34:35] != "6" {
			t.Errorf("expected tpEmis '6' at position 34 in access key, got '%s'", key[34:35])
		}
	})
}

func minimalInvoiceInput() entity.InvoiceInput {
	return entity.InvoiceInput{
		Serie:       1,
		Numero:      1,
		Environment: 1,
		CNF:         "12345678",
		NaturezaOp:  "Venda",
		TpEmis:      defaults.EmissaoNormal,
		Emitter: entity.EmitterData{
			Type:       1,
			Document:   "12345678000195",
			XNome:      "Test Emitter",
			Logradouro: "Rua Test",
			Numero:     "100",
			CodigoMun:  "5103403",
			Municipio:  "Cuiaba",
			UF:         "MT",
			CEP:        "78000000",
			IE:         "12345678901",
			CRT:        "3",
		},
		Recipient: entity.RecipientData{
			Type:       1,
			CNPJ:       "98765432000195",
			XNome:      "Test Recipient",
			Logradouro: "Rua Rec",
			Numero:     "200",
			CodigoMun:  "5103403",
			Municipio:  "Cuiaba",
			UF:         "MT",
			CEP:        "78000000",
			IndIEDest:  "9",
		},
		Items: []entity.ItemData{
			{
				Numero: 1,
				Produto: entity.ProdutoData{
					Codigo:   "1",
					CEAN:     "SEM GTIN",
					XProd:    "Milho",
					NCM:      "10059010",
					CFOP:     "5102",
					UCom:     "KG",
					QCom:     decimal.NewFromInt(100),
					VUnCom:   decimal.NewFromInt(10),
					VProd:    decimal.NewFromInt(1000),
					CEANTrib: "SEM GTIN",
					UTrib:    "KG",
					QTrib:    decimal.NewFromInt(100),
					VUnTrib:  decimal.NewFromInt(10),
					IndTot:   1,
				},
				Imposto: entity.ImpostoData{
					ICMS: entity.ICMSData{
						Origem: "0",
						CST:    "00",
						ModBC:  "3",
					},
					PIS: entity.PISData{
						CST:  "01",
						VBC:  decimal.NewFromInt(1000),
						PPIS: decimal.NewFromFloat(1.65),
						VPIS: decimal.NewFromFloat(16.50),
					},
					COFINS: entity.COFINSData{
						CST:     "01",
						VBC:     decimal.NewFromInt(1000),
						PCOFINS: decimal.NewFromFloat(7.6),
						VCOFINS: decimal.NewFromInt(76),
					},
					IBSCBS: entity.IBSCBSData{
						CST:        "000",
						CClassTrib: "000001",
						VBC:        decimal.NewFromInt(1000),
						VIBSUF:     decimal.NewFromFloat(1.00), // state share (UF)
						VIBSMun:    decimal.Zero,               // municipal share (zero in 2026 phase)
						VIBS:       decimal.NewFromFloat(1.00), // per-item total = UF + VIBSMun
						PCBS:       decimal.NewFromFloat(0.90),
						VCBS:       decimal.NewFromFloat(9.00),
					},
				},
			},
		},
		Transport: entity.TransportData{
			ModFrete: 1,
		},
		Payment: entity.PaymentData{
			IndPag: 1,
			Detalhes: []entity.PagamentoDetalhe{
				{IndPag: 1, TPag: "90", VPag: decimal.NewFromInt(0)},
			},
		},
		TotalValue: decimal.NewFromInt(1000),
	}
}

func TestBuilder_MandatoryEnderEmitFields(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	// Ensure bairro and CEP are set (caller is responsible)
	input.Emitter.Bairro = "Centro"

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// xBairro must always be present (C09, 1-1 mandatory)
	if !strings.Contains(xmlStr, "<xBairro>Centro</xBairro>") {
		t.Error("emit/enderEmit/xBairro must always be present (mandatory)")
	}
	// CEP must always be present (C13, 1-1 mandatory)
	if !strings.Contains(xmlStr, "<CEP>78000000</CEP>") {
		t.Error("emit/enderEmit/CEP must always be present (mandatory)")
	}
}

func TestBuilder_MandatoryEnderDestFields(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	input.Recipient.Bairro = "Centro"

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// xBairro must always be present (E09, 1-1 mandatory)
	if !strings.Contains(xmlStr, "<xBairro>Centro</xBairro>") {
		t.Error("dest/enderDest/xBairro must always be present (mandatory)")
	}
}

func TestBuilder_TransportadoraEndereco(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	input.Transport.Transportadora = &entity.TransportadoraData{
		Type:      1,
		CNPJ:      "11223344000155",
		XNome:     "Transp Test",
		IE:        "123456789",
		Endereco:  "Rua dos Transportes, 100",
		UF:        "MT",
		Municipio: "Sorriso",
	}

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	if !strings.Contains(xmlStr, "<xEnder>Rua dos Transportes, 100</xEnder>") {
		t.Error("xEnder should contain the transportadora's address, not the municipality")
	}
	if !strings.Contains(xmlStr, "<xMun>Sorriso</xMun>") {
		t.Error("xMun should contain the municipality name")
	}
}

func TestBuilder_MultipleVolumesSeparateElements(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	input.Transport.Volumes = []entity.VolumeData{
		{QVol: 1, Esp: "Saco", PesoL: decimal.NewFromInt(50), PesoB: decimal.NewFromInt(51)},
		{QVol: 2, Esp: "Big Bag", PesoL: decimal.NewFromInt(100), PesoB: decimal.NewFromInt(102)},
	}

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// Count <vol> elements — should be 2
	count := strings.Count(xmlStr, "<vol>")
	if count != 2 {
		t.Errorf("expected 2 <vol> elements, got %d", count)
	}
	// Verify each vol has its own qVol
	if !strings.Contains(xmlStr, "<qVol>1</qVol>") {
		t.Error("first vol should have qVol=1")
	}
	if !strings.Contains(xmlStr, "<qVol>2</qVol>") {
		t.Error("second vol should have qVol=2")
	}
}

func TestBuilder_MultiplePaymentsSeparateDetPag(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	input.Payment.Detalhes = []entity.PagamentoDetalhe{
		{IndPag: 0, TPag: "01", VPag: decimal.NewFromInt(500)},
		{IndPag: 1, TPag: "01", VPag: decimal.NewFromInt(500)},
	}

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// Count <detPag> elements — should be 2
	count := strings.Count(xmlStr, "<detPag>")
	if count != 2 {
		t.Errorf("expected 2 <detPag> elements, got %d", count)
	}
	detPags := doc.Root().FindElements("./infNFe/pag/detPag")
	if got := rwElText(detPags[0], "indPag"); got != "" {
		t.Errorf("first detPag indPag = %q, want omitted for detail value 0", got)
	}
	if got := rwElText(detPags[1], "indPag"); got != "1" {
		t.Errorf("second detPag indPag = %q, want 1 from that detail", got)
	}
}

// TestBuilder_IBSCBS verifies the tax-reform <IBSCBS> and <IBSCBSTot> groups
// per NT 2025.002-RTC: per-item <IBSCBS> appears after <COFINS> inside
// <imposto>, and <IBSCBSTot> appears as a sibling of <ICMSTot> inside <total>.
func TestBuilder_IBSCBS(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()
	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// Per-item <IBSCBS> presence + ordering: must be after <COFINS>...</COFINS>
	if !strings.Contains(xmlStr, "<COFINS>") {
		t.Fatal("expected <COFINS> in XML (prerequisite for ordering check)")
	}
	if !strings.Contains(xmlStr, "<IBSCBS>") {
		t.Fatal("expected <IBSCBS> group in XML")
	}
	cofinsIdx := strings.Index(xmlStr, "</COFINS>")
	ibscbsIdx := strings.Index(xmlStr, "<IBSCBS>")
	if ibscbsIdx <= cofinsIdx {
		t.Errorf("expected <IBSCBS> to appear after </COFINS>, got IBSCBS at %d after COFINS close at %d", ibscbsIdx, cofinsIdx)
	}

	// Per-item required children of <IBSCBS>
	for _, want := range []string{
		"<CST>000</CST>",
		"<cClassTrib>000001</cClassTrib>",
		"<gIBSCBS>",
		// vBC is mandatory inside gIBSCBS (UB16, occurrence 1-1) — the very
		// first child of gIBSCBS.
		"<vBC>1000.00</vBC>",
		// <gIBS> wrapper is mandatory inside <gIBSCBS> before <gCBS>.
		// SEFAZ rejects 215 if the wrapper is missing.
		"<gIBS>",
		"<gIBSUF>",
		"<pIBSUF>0.1000</pIBSUF>",
		"<vIBSUF>1.00</vIBSUF>",
		"<gIBSMun>",
		"<pIBSMun>0.0000</pIBSMun>",
		"<vIBSMun>0.00</vIBSMun>",
		// Per-item <vIBS> (total) is mandatory inside <gIBS> and must appear
		// BEFORE <gCBS> as a sibling inside <gIBSCBS>.
		"<vIBS>1.00</vIBS>",
		// vCredPres + vCredPresCondSus are TOTALS-ONLY — they must NOT appear
		// as direct children of per-item <gIBS>. Emitting them here would
		// produce SEFAZ 215 with cvc-complex-type.2.4.d. Per-item credit-
		// presumption fields belong inside the optional <gIBSCredPres> /
		// <gCBSCredPres> subgroups, which we don't emit in the 2026 phase.
		"<gCBS>",
		"<pCBS>0.9000</pCBS>",
		"<vCBS>9.00</vCBS>",
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("expected %q in XML", want)
		}
	}

	// <IBSCBSTot> presence + ordering: must be after </ICMSTot>
	if !strings.Contains(xmlStr, "<IBSCBSTot>") {
		t.Fatal("expected <IBSCBSTot> group in XML")
	}
	icmsTotIdx := strings.Index(xmlStr, "</ICMSTot>")
	ibscbsTotIdx := strings.Index(xmlStr, "<IBSCBSTot>")
	if ibscbsTotIdx <= icmsTotIdx {
		t.Errorf("expected <IBSCBSTot> to appear after </ICMSTot>, got IBSCBSTot at %d after ICMSTot close at %d", ibscbsTotIdx, icmsTotIdx)
	}

	// Per-NF-e totals — every mandatory child of <IBSCBSTot>/<gIBS>/<gCBSUF>/
	// <gIBSMun>/<gCBS> per NT 2025.002-RTC v1.36 §6.7.4. SEFAZ rejects with
	// status 215 when any of these are missing.
	for _, want := range []string{
		"<vBCIBSCBS>1000.00</vBCIBSCBS>",
		// gIBSUF (totals): vDif + vDevTrib + vIBSUF, in that order
		"<vDif>0.00</vDif>",
		"<vDevTrib>0.00</vDevTrib>",
		"<vIBSUF>1.00</vIBSUF>",
		// gIBSMun (totals): vDif + vDevTrib + vIBSMun, in that order
		"<vIBSMun>0.00</vIBSMun>",
		// gIBS (totals): vIBS, vCredPres, vCredPresCondSus
		"<vIBS>1.00</vIBS>",
		// vCredPres + vCredPresCondSus DO appear in totals <gIBS> as direct
		// children (W48, W49). Distinct from the per-item block, where they
		// are forbidden.
		"<vCredPres>0.00</vCredPres>",
		"<vCredPresCondSus>0.00</vCredPresCondSus>",
		// gCBS (totals): vDif + vDevTrib + vCBS + vCredPres + vCredPresCondSus
		"<vCBS>9.00</vCBS>",
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("expected %q in XML totals", want)
		}
	}
}

// TestBuilder_IBSCBS_ElementOrder walks both the per-item <gIBSCBS> wrapper
// and the totals <IBSCBSTot> group and asserts the canonical element order
// per NT 2025.002-RTC v1.36 §6.7.4. This is the regression guard for the
// two SEFAZ 215 patterns that have already cost emissions:
//
//  1. Per-item <gIBSCBS> missing the <gIBS> wrapper — parser sees <gCBS>
//     where it expected <vIBS>.
//  2. Totals <IBSCBSTot>/<gIBS>/<gCBS> children missing or out of order
//     — parser sees <vCBS> where it expected <vDif>.
func TestBuilder_IBSCBS_ElementOrder(t *testing.T) {
	builder := xml.NewBuilder()
	doc, err := builder.Build(minimalInvoiceInput())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	childNames := func(parent *etree.Element) []string {
		names := make([]string, 0, len(parent.ChildElements()))
		for _, c := range parent.ChildElements() {
			names = append(names, c.Tag)
		}
		return names
	}

	root := doc.Root()

	// --- Per-item side -----------------------------------------------------
	// <gIBSCBS> is the only <gIBSCBS> in the fixture (the totals block uses
	// <IBSCBSTot> as its container, which has no direct child <gIBSCBS>).
	var perItemGIBSCBS *etree.Element
	for _, c := range root.FindElements("./infNFe/det/imposto/IBSCBS/gIBSCBS") {
		perItemGIBSCBS = c
		break
	}
	if perItemGIBSCBS == nil {
		t.Fatal("per-item <gIBSCBS> not found in built XML")
	}
	if got, want := strings.Join(childNames(perItemGIBSCBS), ","), defaults.PerItemGIBSCBSOrder; got != want {
		t.Errorf("per-item <gIBSCBS> child order = %q, want %q", got, want)
	}

	// Per-item <gIBS> wrapper does NOT exist (NT 2025.002-RTC v1.51 §6.7.4
	// UB15 is a FLAT sequence). gIBSUF/gIBSMun/gCBS are direct siblings of
	// vIBS inside <gIBSCBS>. This is the canonical structure that fixes the
	// 4th SEFAZ 215 error ("Invalid content starting with element 'gIBS'.
	// One of 'gIBSUF' is expected").
	//
	// Per-item <gIBSUF> children (direct child of gIBSCBS)
	var perItemGIBSUF *etree.Element
	for _, c := range perItemGIBSCBS.FindElements("./gIBSUF") {
		perItemGIBSUF = c
		break
	}
	if perItemGIBSUF == nil {
		t.Fatal("per-item <gIBSUF> not found as direct child of <gIBSCBS>")
	}
	if got, want := strings.Join(childNames(perItemGIBSUF), ","), defaults.PerItemGIBSUFOrder; got != want {
		t.Errorf("per-item <gIBSUF> child order = %q, want %q", got, want)
	}

	// Per-item <gIBSMun> children (direct child of gIBSCBS)
	var perItemGIBSMun *etree.Element
	for _, c := range perItemGIBSCBS.FindElements("./gIBSMun") {
		perItemGIBSMun = c
		break
	}
	if perItemGIBSMun == nil {
		t.Fatal("per-item <gIBSMun> not found as direct child of <gIBSCBS>")
	}
	if got, want := strings.Join(childNames(perItemGIBSMun), ","), defaults.PerItemGIBSMunOrder; got != want {
		t.Errorf("per-item <gIBSMun> child order = %q, want %q", got, want)
	}

	// Per-item <gCBS> children (direct sibling of <gIBSUF>/<gIBSMun>/<vIBS>)
	var perItemGCBS *etree.Element
	for _, c := range perItemGIBSCBS.FindElements("./gCBS") {
		perItemGCBS = c
		break
	}
	if perItemGCBS == nil {
		t.Fatal("per-item <gCBS> not found as direct child of <gIBSCBS>")
	}
	if got, want := strings.Join(childNames(perItemGCBS), ","), defaults.PerItemGCBSOrder; got != want {
		t.Errorf("per-item <gCBS> child order = %q, want %q", got, want)
	}

	// --- Totals side -------------------------------------------------------
	var ibscbsTot *etree.Element
	for _, child := range root.FindElements("//IBSCBSTot") {
		ibscbsTot = child
		break
	}
	if ibscbsTot == nil {
		t.Fatal("<IBSCBSTot> not found in built XML")
	}

	// Totals gIBSUF
	var tGIBSUF *etree.Element
	for _, c := range ibscbsTot.FindElements("//gIBSUF") {
		// skip per-item one (already covered above)
		if c.Parent().Parent().Parent() != nil && c.Parent().Parent().Parent().Tag == "IBSCBSTot" {
			tGIBSUF = c
			break
		}
	}
	if tGIBSUF == nil {
		// fallback: find any gIBSUF under IBSCBSTot
		for _, c := range ibscbsTot.FindElements(".//gIBSUF") {
			tGIBSUF = c
			break
		}
	}
	if tGIBSUF == nil {
		t.Fatal("totals <gIBSUF> not found inside <IBSCBSTot>")
	}
	if got, want := strings.Join(childNames(tGIBSUF), ","), defaults.IBSCBSTotGIBSUFOrder; got != want {
		t.Errorf("totals gIBSUF child order = %q, want %q", got, want)
	}

	// Totals gIBSMun
	var tGIBSMun *etree.Element
	for _, c := range ibscbsTot.FindElements(".//gIBSMun") {
		tGIBSMun = c
		break
	}
	if tGIBSMun == nil {
		t.Fatal("totals <gIBSMun> not found inside <IBSCBSTot>")
	}
	if got, want := strings.Join(childNames(tGIBSMun), ","), defaults.IBSCBSTotGIBSMunOrder; got != want {
		t.Errorf("totals gIBSMun child order = %q, want %q", got, want)
	}

	// Totals gCBS
	var tGCBS *etree.Element
	for _, c := range ibscbsTot.FindElements("./gCBS") {
		tGCBS = c
		break
	}
	if tGCBS == nil {
		t.Fatal("totals <gCBS> not found inside <IBSCBSTot>")
	}
	if got, want := strings.Join(childNames(tGCBS), ","), defaults.IBSCBSTotGCBSOrder; got != want {
		t.Errorf("totals gCBS child order = %q, want %q", got, want)
	}

	// Totals gIBS siblings
	var tGIBS *etree.Element
	for _, c := range ibscbsTot.FindElements("./gIBS") {
		tGIBS = c
		break
	}
	if tGIBS == nil {
		t.Fatal("totals <gIBS> not found inside <IBSCBSTot>")
	}
	if got, want := strings.Join(childNames(tGIBS), ","), defaults.IBSCBSTotGIBSOrder; got != want {
		t.Errorf("totals gIBS child order = %q, want %q", got, want)
	}

	// IBSCBSTot root
	if got, want := strings.Join(childNames(ibscbsTot), ","), defaults.IBSCBSTotRootOrder; got != want {
		t.Errorf("IBSCBSTot child order = %q, want %q", got, want)
	}
}

func TestBuilder_TransportVehicleAndVolumes(t *testing.T) {
	builder := xml.NewBuilder()
	input := minimalInvoiceInput()

	input.Transport.Transportadora = &entity.TransportadoraData{
		Type:      2,
		CPF:       "52998224725",
		XNome:     "Carreteiro Fulano",
		Endereco:  "Rodovia BR-163, Km 12",
		Municipio: "Sorriso",
		UF:        "MT",
	}
	input.Transport.Veiculo = &entity.VeiculoData{
		Placa: "ABC1D23",
		UF:    "MT",
		RNTC:  "12345678901234567890",
	}
	input.Transport.Volumes = []entity.VolumeData{
		{QVol: 20, Esp: "Big Bag", Marca: "BigFarm", PesoL: decimal.RequireFromString("15000.500"), PesoB: decimal.RequireFromString("15020.250")},
	}

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	xmlStr, _ := doc.WriteToString()

	// Transportadora (CPF branch must emit <CPF>, not <CNPJ>)
	if !strings.Contains(xmlStr, "<CPF>52998224725</CPF>") {
		t.Error("transporta/CPF missing for a Type-2 transportadora")
	}
	if !strings.Contains(xmlStr, "<xNome>Carreteiro Fulano</xNome>") {
		t.Error("transporta/xNome missing")
	}
	// Vehicle group, including RNTC (ANTT)
	if !strings.Contains(xmlStr, "<veicTransp><placa>ABC1D23</placa><UF>MT</UF><RNTC>12345678901234567890</RNTC></veicTransp>") {
		t.Error("veicTransp group (placa/UF/RNTC) missing or out of order")
	}
	// Volume block with marca and v3 weights
	if !strings.Contains(xmlStr, "<vol><qVol>20</qVol><esp>Big Bag</esp><marca>BigFarm</marca><pesoL>15000.500</pesoL><pesoB>15020.250</pesoB></vol>") {
		t.Error("vol block (qVol/esp/marca/pesoL/pesoB) missing or out of order")
	}
}

// ---------------------------------------------------------------------------
// Real-world corpus gap fixes (docs/sdd/nfe_builder_replay_gaps/)

// formDecimalBookend returns a representative ICMSData for the table tests:
// scenario values mirror the corpus shapes (all amounts zero in the corpus,
// non-zero here so formatting is asserted too).
func icmsVariant(cst, modBC string) entity.ICMSData {
	return entity.ICMSData{
		Origem:  "0",
		CST:     cst,
		ModBC:   modBC,
		VBC:     decimal.NewFromFloat(100.00),
		PICMS:   decimal.NewFromFloat(18.00),
		VICMS:   decimal.NewFromFloat(18.00),
		PRedBC:  decimal.NewFromFloat(40.00),
		ModBCST: "4",
		VBCST:   decimal.NewFromFloat(120.00),
		PICMSST: decimal.NewFromFloat(18.00),
		VICMSST: decimal.NewFromFloat(21.60),
	}
}

// TestBuilder_ICMS verifies the regime-normal ICMS group switch over the
// corpus-verified CST variants (docs/sdd/nfe_builder_replay_gaps AC1): tag
// choice AND exact child sequence per CST.
func TestBuilder_ICMS(t *testing.T) {
	builder := xml.NewBuilder()

	cases := []struct {
		name     string
		icms     entity.ICMSData
		wantTag  string
		wantKids []string
	}{
		{
			// Compat guard: the CST-00 shape is byte-identical to the
			// pre-change builder (ICMS00 with modBC, vBC, pICMS, vICMS).
			name:     "CST00_ICMS00_unchanged",
			icms:     entity.ICMSData{Origem: "0", CST: "00", ModBC: "3", VBC: decimal.NewFromInt(100), PICMS: decimal.NewFromFloat(18), VICMS: decimal.NewFromInt(18)},
			wantTag:  "ICMS00",
			wantKids: []string{"orig", "CST", "modBC", "vBC", "pICMS", "vICMS"},
		},
		{
			name:     "CST20_ICMS20",
			icms:     icmsVariant("20", "3"),
			wantTag:  "ICMS20",
			wantKids: []string{"orig", "CST", "modBC", "pRedBC", "vBC", "pICMS", "vICMS"},
		},
		{
			name:     "CST40_ICMS40",
			icms:     icmsVariant("40", ""),
			wantTag:  "ICMS40",
			wantKids: []string{"orig", "CST"},
		},
		{
			name:     "CST41_ICMS40",
			icms:     icmsVariant("41", ""),
			wantTag:  "ICMS40",
			wantKids: []string{"orig", "CST"},
		},
		{
			name:     "CST50_ICMS40",
			icms:     icmsVariant("50", ""),
			wantTag:  "ICMS40",
			wantKids: []string{"orig", "CST"},
		},
		{
			name:     "CST51_ICMS51_simplified",
			icms:     icmsVariant("51", "3"),
			wantTag:  "ICMS51",
			wantKids: []string{"orig", "CST", "modBC", "pRedBC", "vBC", "pICMS", "vICMS"},
		},
		{
			name:     "CST90_ICMS90",
			icms:     icmsVariant("90", "3"),
			wantTag:  "ICMS90",
			wantKids: []string{"orig", "CST", "modBC", "vBC", "pRedBC", "pICMS", "vICMS", "modBCST", "vBCST", "pICMSST", "vICMSST"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			input := minimalInvoiceInput()
			input.Items[0].Imposto.ICMS = tc.icms
			doc, err := builder.Build(input)
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}
			icmsWrap := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS")
			assertChildOrder(t, icmsWrap, "ICMS children", []string{tc.wantTag})

			group := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/"+tc.wantTag)
			assertChildOrder(t, group, tc.wantTag+" children", tc.wantKids)

			// Value formatting check: ICMS40 carries no amounts; the others
			// keep the existing 2dp/4dp precision.
			if tag := group.SelectElement("pRedBC"); tag != nil && tag.Text() != "40.0000" {
				t.Errorf("pRedBC = %q, want 40.0000", tag.Text())
			}
			if tag := group.SelectElement("vBCST"); tag != nil && tag.Text() != "120.00" {
				t.Errorf("vBCST = %q, want 120.00", tag.Text())
			}
		})
	}
}

// TestBuilder_ICMS_OmittedOptionalTerms walks the corpus-omission rules:
// modBC/modBCST are 0-1 inside ICMS20/51/90 and only emitted when set; the
// structured ICMS51 deferral terms (vICMSOp/pDif/vICMSDif) and the ICMS20
// desoneração pair are never emitted.
func TestBuilder_ICMS_OmittedOptionalTerms(t *testing.T) {
	builder := xml.NewBuilder()

	t.Run("CST51_without_modBC", func(t *testing.T) {
		input := minimalInvoiceInput()
		icms := icmsVariant("51", "")
		icms.ModBC = ""
		input.Items[0].Imposto.ICMS = icms
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		group := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/ICMS51")
		assertChildOrder(t, group, "ICMS51 children", []string{"orig", "CST", "pRedBC", "vBC", "pICMS", "vICMS"})
	})

	t.Run("CST51_no_deferral_terms", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Items[0].Imposto.ICMS = icmsVariant("51", "3")
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		group := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/ICMS51")
		for _, banned := range []string{"vICMSOp", "pDif", "vICMSDif", "vICMSDeson", "motivoDesoneracao"} {
			if group.SelectElement(banned) != nil {
				t.Errorf("ICMS51 must not emit <%s> (corpus never does)", banned)
			}
		}
	})

	t.Run("CST20_no_desoneracao_terms", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Items[0].Imposto.ICMS = icmsVariant("20", "3")
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		group := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/ICMS20")
		for _, banned := range []string{"vICMSDeson", "motivoDesoneracao"} {
			if group.SelectElement(banned) != nil {
				t.Errorf("ICMS20 must not emit <%s> (corpus never does)", banned)
			}
		}
	})

	t.Run("CST90_modBCST_positional_when_unset", func(t *testing.T) {
		// SEFAZ-verified (215, cvc-complex-type.2.4.a): the ICMS-ST block
		// inside ICMS90 must OPEN with <modBCST> — a document that starts
		// the block at <vBCST> is rejected ("One of 'vBCFCP, pFCPDif,
		// modBCST, vICMSDeson, vICMSSTDeson' is expected"). Empty ModBCST
		// defaults to "0" (corpus value), keeping the zero-valued ST chunk
		// structurally valid.
		input := minimalInvoiceInput()
		icms := icmsVariant("90", "3")
		icms.ModBCST = ""
		input.Items[0].Imposto.ICMS = icms
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		group := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/ICMS90")
		assertChildOrder(t, group, "ICMS90 children", []string{"orig", "CST", "modBC", "vBC", "pRedBC", "pICMS", "vICMS", "modBCST", "vBCST", "pICMSST", "vICMSST"})
		if got, want := rwElText(group, "modBCST"), "0"; got != want {
			t.Errorf("modBCST = %q, want the positional default %q", got, want)
		}
	})
}

// TestBuilder_PIS_COFINS_Variants verifies the PIS/COFINS CST switch (AC2):
// CST 04–09 → PISNT/COFINSNT with a single <CST> child; CST 01 keeps the
// PISAliq/COFINSAliq shape byte-identical to the pre-change builder.
func TestBuilder_PIS_COFINS_Variants(t *testing.T) {
	builder := xml.NewBuilder()

	for _, cst := range []string{"04", "05", "06", "07", "08", "09"} {
		cst := cst
		t.Run("NT_"+cst, func(t *testing.T) {
			input := minimalInvoiceInput()
			input.Items[0].Imposto.PIS = entity.PISData{CST: cst}
			input.Items[0].Imposto.COFINS = entity.COFINSData{CST: cst}
			doc, err := builder.Build(input)
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}
			pis := findElement(t, doc.Root(), "./infNFe/det/imposto/PIS")
			assertChildOrder(t, pis, "PIS children", []string{"PISNT"})
			assertChildOrder(t, pis.SelectElement("PISNT"), "PISNT children", []string{"CST"})
			if got := pis.SelectElement("PISNT").SelectElement("CST").Text(); got != cst {
				t.Errorf("PISNT/CST = %q, want %q", got, cst)
			}
			cofins := findElement(t, doc.Root(), "./infNFe/det/imposto/COFINS")
			assertChildOrder(t, cofins, "COFINS children", []string{"COFINSNT"})
			assertChildOrder(t, cofins.SelectElement("COFINSNT"), "COFINSNT children", []string{"CST"})
			if got := cofins.SelectElement("COFINSNT").SelectElement("CST").Text(); got != cst {
				t.Errorf("COFINSNT/CST = %q, want %q", got, cst)
			}
		})
	}

	t.Run("default_CST01_keeps_Aliq", func(t *testing.T) {
		input := minimalInvoiceInput() // CST 01 with amounts (compat guard)
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		pis := findElement(t, doc.Root(), "./infNFe/det/imposto/PIS")
		assertChildOrder(t, pis, "PIS children", []string{"PISAliq"})
		assertChildOrder(t, pis.SelectElement("PISAliq"), "PISAliq children", []string{"CST", "vBC", "pPIS", "vPIS"})
		cofins := findElement(t, doc.Root(), "./infNFe/det/imposto/COFINS")
		assertChildOrder(t, cofins, "COFINS children", []string{"COFINSAliq"})
		assertChildOrder(t, cofins.SelectElement("COFINSAliq"), "COFINSAliq children", []string{"CST", "vBC", "pCOFINS", "vCOFINS"})
	})
}

// g1StyleInput mirrors the corpus G1 grain-sale shape: one item, no
// discounts, IBS/CBS amounts present, vNF from input.TotalValue.
func g1StyleInput() entity.InvoiceInput {
	input := minimalInvoiceInput()
	input.Items[0].Imposto.ICMS = entity.ICMSData{
		Origem: "0",
		CST:    "51",
		ModBC:  "3",
	}
	input.Items[0].Imposto.PIS = entity.PISData{CST: "08"}
	input.Items[0].Imposto.COFINS = entity.COFINSData{CST: "08"}
	input.Items[0].Produto.VProd = decimal.RequireFromString("61385.23")
	input.Items[0].Imposto.IBSCBS = entity.IBSCBSData{
		CST:        "000",
		CClassTrib: "000001",
		VBC:        decimal.RequireFromString("61385.23"),
		VIBSUF:     decimal.RequireFromString("61.39"),
		VIBS:       decimal.RequireFromString("61.39"),
		PCBS:       decimal.NewFromFloat(0.9),
		VCBS:       decimal.RequireFromString("552.47"),
	}
	input.TotalValue = decimal.RequireFromString("61385.23")
	return input
}

// TestBuilder_VItem_VNFTot_NumericFormula verifies the full NT 2025.002
// vItem formula and the vNFTot = Σ vItem total (AC3), including the
// deliberate vItem ≠ vNF divergence.
func TestBuilder_VItem_VNFTot_NumericFormula(t *testing.T) {
	builder := xml.NewBuilder()

	t.Run("G1_style_no_discount", func(t *testing.T) {
		// vItem = 61385.23 − 0 + 61.39 + 0 + 552.47 = 61999.09
		input := g1StyleInput()
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		vItem := rwElText(findElement(t, doc.Root(), "./infNFe/det"), "vItem")
		if got, want := vItem, "61999.09"; got != want {
			t.Errorf("det/vItem = %q, want %q", got, want)
		}
		total := findElement(t, doc.Root(), "./infNFe/total")
		assertChildOrder(t, total, "total children", []string{"ICMSTot", "IBSCBSTot", "vNFTot"})
		vNFTot := rwElText(total, "vNFTot")
		if got, want := vNFTot, "61999.09"; got != want {
			t.Errorf("total/vNFTot = %q, want %q", got, want)
		}
		// vNF remains sourced from input.TotalValue — deliberately ≠ vNFTot.
		if got, want := rwElText(total.SelectElement("ICMSTot"), "vNF"), "61385.23"; got != want {
			t.Errorf("total/vNF = %q, want %q (vNF from input, vNFTot from formula)", got, want)
		}
		// No discounts: totals keep 0.00 and prod has no vDesc.
		discount := findElement(t, doc.Root(), "./infNFe/det/prod").SelectElement("vDesc")
		if discount != nil {
			t.Errorf("prod/vDesc must be omitted when VDesc is zero")
		}
		if got, want := rwElText(total.SelectElement("ICMSTot"), "vDesc"), "0.00"; got != want {
			t.Errorf("ICMSTot/vDesc = %q, want %q", got, want)
		}
	})

	t.Run("G3_style_with_discount", func(t *testing.T) {
		input := g1StyleInput()
		prod := &input.Items[0].Produto
		prod.VProd = decimal.RequireFromString("1277.55")
		prod.QCom = decimal.NewFromInt(3)
		prod.VUnCom = decimal.RequireFromString("425.85")
		prod.VUnTrib = decimal.RequireFromString("425.85")
		prod.VDesc = decimal.RequireFromString("319.39")
		ibscbs := &input.Items[0].Imposto.IBSCBS
		ibscbs.VBC = decimal.RequireFromString("1277.55")
		ibscbs.VIBSUF = decimal.RequireFromString("1.28")
		ibscbs.VIBS = decimal.RequireFromString("1.28")
		ibscbs.VCBS = decimal.RequireFromString("11.50")
		// vNF stays the invoice total: 1277.55 − 319.39 = 958.16
		input.TotalValue = decimal.RequireFromString("958.16")

		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		prodEl := findElement(t, doc.Root(), "./infNFe/det/prod")
		assertChildOrder(t, prodEl, "prod children", []string{"cProd", "cEAN", "xProd", "NCM", "CFOP", "uCom", "qCom", "vUnCom", "vProd", "cEANTrib", "uTrib", "qTrib", "vUnTrib", "vDesc", "indTot"})
		if got, want := rwElText(prodEl, "vDesc"), "319.39"; got != want {
			t.Errorf("prod/vDesc = %q, want %q", got, want)
		}
		// vItem = 1277.55 − 319.39 + 1.28 + 0 + 11.50 = 970.94
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/det"), "vItem"), "970.94"; got != want {
			t.Errorf("det/vItem = %q, want %q", got, want)
		}
		icmsTot := findElement(t, doc.Root(), "./infNFe/total/ICMSTot")
		if got, want := rwElText(icmsTot, "vDesc"), "319.39"; got != want {
			t.Errorf("ICMSTot/vDesc = %q, want %q", got, want)
		}
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/total"), "vNFTot"), "970.94"; got != want {
			t.Errorf("total/vNFTot = %q, want %q", got, want)
		}
	})

	t.Run("det_child_order_with_and_without_infAdProd", func(t *testing.T) {
		input := g1StyleInput()
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		det := findElement(t, doc.Root(), "./infNFe/det")
		assertChildOrder(t, det, "det children", []string{"prod", "imposto", "vItem"})

		// With infAdProd, vItem stays last.
		input2 := g1StyleInput()
		input2.Items[0].InfAdProd = "Referente ao romaneio 123"
		doc2, err := builder.Build(input2)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		det2 := findElement(t, doc2.Root(), "./infNFe/det")
		assertChildOrder(t, det2, "det children", []string{"prod", "imposto", "infAdProd", "vItem"})
		if got, want := rwElText(det2, "infAdProd"), "Referente ao romaneio 123"; got != want {
			t.Errorf("det/infAdProd = %q, want %q", got, want)
		}

		// Free-text sanitization (schema TString): accented characters fold
		// to their ASCII base, line breaks become "; " and surrounding
		// whitespace is trimmed — the same contract SEFAZ-MT requires of
		// every área-de-dados text.
		input3 := g1StyleInput()
		input3.Items[0].InfAdProd = "  Série nº 123\nObservação  "
		doc3, err := builder.Build(input3)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if got, want := rwElText(findElement(t, doc3.Root(), "./infNFe/det"), "infAdProd"), "Serie no 123; Observacao"; got != want {
			t.Errorf("det/infAdProd sanitized = %q, want %q", got, want)
		}
	})

	t.Run("vNFTot_sums_emitted_per_item_values", func(t *testing.T) {
		// Two items whose raw vItem lands on a half-cent; vNFTot must sum
		// the emitted (2dp) values, not the raw ones.
		input := g1StyleInput()
		input.Items = []entity.ItemData{
			{
				Numero: 1,
				Produto: entity.ProdutoData{
					Codigo: "1", CEAN: "SEM GTIN", XProd: "A", NCM: "12019000", CFOP: "5101",
					UCom: "KG", QCom: decimal.NewFromInt(1), VUnCom: decimal.RequireFromString("100.005"),
					VProd: decimal.RequireFromString("100.005"), CEANTrib: "SEM GTIN", UTrib: "KG",
					QTrib: decimal.NewFromInt(1), VUnTrib: decimal.RequireFromString("100.005"), IndTot: 1,
				},
				Imposto: entity.ImpostoData{
					ICMS: entity.ICMSData{Origem: "0", CST: "51", ModBC: "3"},
					PIS:  entity.PISData{CST: "08"}, COFINS: entity.COFINSData{CST: "08"},
				},
			},
			{
				Numero: 2,
				Produto: entity.ProdutoData{
					Codigo: "2", CEAN: "SEM GTIN", XProd: "B", NCM: "12019000", CFOP: "5101",
					UCom: "KG", QCom: decimal.NewFromInt(1), VUnCom: decimal.RequireFromString("200.005"),
					VProd: decimal.RequireFromString("200.005"), CEANTrib: "SEM GTIN", UTrib: "KG",
					QTrib: decimal.NewFromInt(1), VUnTrib: decimal.RequireFromString("200.005"), IndTot: 1,
				},
				Imposto: entity.ImpostoData{
					ICMS: entity.ICMSData{Origem: "0", CST: "51", ModBC: "3"},
					PIS:  entity.PISData{CST: "08"}, COFINS: entity.COFINSData{CST: "08"},
				},
			},
		}
		input.TotalValue = decimal.RequireFromString("300.01")
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		dets := doc.Root().FindElements("./infNFe/det")
		if len(dets) != 2 {
			t.Fatalf("expected 2 det elements, got %d", len(dets))
		}
		if got, want := rwElText(dets[0], "vItem"), "100.00"; got != want {
			t.Errorf("det[1]/vItem = %q, want %q (2dp truncation)", got, want)
		}
		if got, want := rwElText(dets[1], "vItem"), "200.00"; got != want {
			t.Errorf("det[2]/vItem = %q, want %q (2dp truncation)", got, want)
		}
		// Σ emitted values = 300.00 (raw sum would be 300.01).
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/total"), "vNFTot"), "300.00"; got != want {
			t.Errorf("total/vNFTot = %q, want %q (sum of emitted per-item values)", got, want)
		}
	})
}

// TestBuilder_SmallRealworldFields verifies presence/position/omission rules
// for the small corpus fields (AC4): dhSaiEnt, infRespTec, cPais/xPais,
// dest/email, detPag/xPag, indPag=0 omission, empty veicTransp/UF.
func TestBuilder_SmallRealworldFields(t *testing.T) {
	builder := xml.NewBuilder()
	dh, err := time.Parse(time.RFC3339, "2026-07-16T08:06:19-04:00")
	if err != nil {
		t.Fatalf("fixture time: %v", err)
	}

	t.Run("dhSaiEnt_after_dhEmi_when_set", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.DhSaiEnt = &dh
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		ide := findElement(t, doc.Root(), "./infNFe/ide")
		assertChildOrder(t, ide, "ide children", []string{"cUF", "cNF", "natOp", "mod", "serie", "nNF", "dhEmi", "dhSaiEnt", "tpNF", "idDest", "cMunFG", "tpImp", "tpEmis", "cDV", "tpAmb", "finNFe", "indFinal", "indPres", "indIntermed", "procEmi", "verProc"})
		if got, want := rwElText(ide, "dhSaiEnt"), "2026-07-16T08:06:19-04:00"; got != want {
			t.Errorf("dhSaiEnt = %q, want %q (RFC3339)", got, want)
		}
	})

	t.Run("dhSaiEnt_defaults_to_dhEmi_when_nil", func(t *testing.T) {
		// Corpus rule: dhSaiEnt == dhEmi (SEFAZ rejects dhSaiEnt < dhEmi;
		// the default shares the same clock as dhEmi, so the texts match).
		doc, err := builder.Build(minimalInvoiceInput())
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		ide := findElement(t, doc.Root(), "./infNFe/ide")
		dhEmi, dhSaiEnt := rwElText(ide, "dhEmi"), rwElText(ide, "dhSaiEnt")
		if dhSaiEnt == "" {
			t.Fatal("dhSaiEnt must be emitted with dhEmi's instant when DhSaiEnt is nil")
		}
		if dhSaiEnt != dhEmi {
			t.Errorf("default dhSaiEnt = %q, want the dhEmi instant %q", dhSaiEnt, dhEmi)
		}
		assertChildOrder(t, ide, "ide children", []string{"cUF", "cNF", "natOp", "mod", "serie", "nNF", "dhEmi", "dhSaiEnt", "tpNF", "idDest", "cMunFG", "tpImp", "tpEmis", "cDV", "tpAmb", "finNFe", "indFinal", "indPres", "indIntermed", "procEmi", "verProc"})
	})

	t.Run("infRespTec_last_child_of_infNFe", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.RespTec = &entity.RespTecData{
			CNPJ:     "45624641000155",
			XContato: "Responsável Técnico",
			Email:    "resp.tecnico@exemplo.com.br",
			Fone:     "65705220883",
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		infNFe := findElement(t, doc.Root(), "./infNFe")
		kids := childTags(infNFe)
		if len(kids) == 0 || kids[len(kids)-1] != "infRespTec" {
			t.Errorf("infRespTec must be the last child of infNFe, got %v", kids)
		}
		respTec := infNFe.SelectElement("infRespTec")
		assertChildOrder(t, respTec, "infRespTec children", []string{"CNPJ", "xContato", "email", "fone"})
		// Free-text fields are sanitized ("Responsável Técnico" → ASCII).
		if got, want := rwElText(respTec, "xContato"), "Responsavel Tecnico"; got != want {
			t.Errorf("xContato = %q, want %q (sanitized)", got, want)
		}
	})

	t.Run("infRespTec_omitted_when_nil", func(t *testing.T) {
		doc, err := builder.Build(minimalInvoiceInput())
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if doc.Root().FindElement("./infNFe/infRespTec") != nil {
			t.Error("infRespTec must be omitted when RespTec is nil")
		}
	})

	t.Run("infRespTec_full_provider_shape", func(t *testing.T) {
		// System provider shape (docs/sdd/resptec_system_defaults): all
		// four E28a fields are mandatory (1-1) — the block must emit
		// CNPJ → xContato → email → fone in the layout order. An empty
		// <email> would get SEFAZ 215 ("One of 'email' is expected").
		input := minimalInvoiceInput()
		input.RespTec = &entity.RespTecData{
			CNPJ:     "47962538000140",
			XContato: "VAGNER SIMONE MARTINS FILHO CONSULTORIA EM TECNOLOGIA DA INFORMACAO LTDA",
			Email:    "vagner.filhopro@gmail.com",
			Fone:     "66996615930",
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		respTec := findElement(t, doc.Root(), "./infNFe/infRespTec")
		assertChildOrder(t, respTec, "infRespTec children", []string{"CNPJ", "xContato", "email", "fone"})
		if got, want := rwElText(respTec, "CNPJ"), "47962538000140"; got != want {
			t.Errorf("infRespTec/CNPJ = %q, want %q", got, want)
		}
		if got, want := rwElText(respTec, "email"), "vagner.filhopro@gmail.com"; got != want {
			t.Errorf("infRespTec/email = %q, want %q", got, want)
		}
		if got, want := rwElText(respTec, "fone"), "66996615930"; got != want {
			t.Errorf("infRespTec/fone = %q, want %q", got, want)
		}
	})

	t.Run("cPais_xPais_after_CEP_in_both_enders", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Emitter.Bairro = "Centro"
		input.Recipient.Bairro = "Centro"
		input.Emitter.Fone = "65337345815"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		enderEmit := findElement(t, doc.Root(), "./infNFe/emit/enderEmit")
		assertChildOrder(t, enderEmit, "enderEmit children", []string{"xLgr", "nro", "xBairro", "cMun", "xMun", "UF", "CEP", "cPais", "xPais", "fone"})
		enderDest := findElement(t, doc.Root(), "./infNFe/dest/enderDest")
		assertChildOrder(t, enderDest, "enderDest children", []string{"xLgr", "nro", "xBairro", "cMun", "xMun", "UF", "CEP", "cPais", "xPais"})
		if got, want := rwElText(enderEmit, "cPais"), "1058"; got != want {
			t.Errorf("cPais = %q, want %q", got, want)
		}
		if got, want := rwElText(enderDest, "xPais"), "Brasil"; got != want {
			t.Errorf("xPais = %q, want %q", got, want)
		}
	})

	t.Run("dest_email_after_IE", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Recipient.IE = "517680283" // corpus destination carries IE
		input.Recipient.Email = "dest151069@exemplo.com.br"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		dest := findElement(t, doc.Root(), "./infNFe/dest")
		assertChildOrder(t, dest, "dest children", []string{"CNPJ", "xNome", "enderDest", "indIEDest", "IE", "email"})
		if got, want := rwElText(dest, "email"), "dest151069@exemplo.com.br"; got != want {
			t.Errorf("dest/email = %q, want %q", got, want)
		}
	})

	t.Run("dest_email_omitted_when_empty", func(t *testing.T) {
		doc, err := builder.Build(minimalInvoiceInput())
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if doc.Root().FindElement("./infNFe/dest/email") != nil {
			t.Error("dest/email must be omitted when Email is empty")
		}
	})

	t.Run("xPag_between_tPag_and_vPag", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Payment.Detalhes = []entity.PagamentoDetalhe{{IndPag: 1, TPag: "99", XPag: "Outros", VPag: decimal.NewFromInt(1000)}}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		detPag := findElement(t, doc.Root(), "./infNFe/pag/detPag")
		assertChildOrder(t, detPag, "detPag children", []string{"indPag", "tPag", "xPag", "vPag"})
		if got, want := rwElText(detPag, "xPag"), "Outros"; got != want {
			t.Errorf("xPag = %q, want %q", got, want)
		}
	})

	t.Run("xPag_default_Outros_for_tPag99", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Payment.Detalhes = []entity.PagamentoDetalhe{{IndPag: 1, TPag: "99", VPag: decimal.NewFromInt(1000)}}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		detPag := findElement(t, doc.Root(), "./infNFe/pag/detPag")
		assertChildOrder(t, detPag, "detPag children", []string{"indPag", "tPag", "xPag", "vPag"})
		if got, want := rwElText(detPag, "xPag"), "Outros"; got != want {
			t.Errorf("default xPag for tPag 99 = %q, want %q", got, want)
		}
	})

	t.Run("xPag_explicit_wins_and_other_tPag_omit", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Payment.IndPag = 0
		// tPag 99 with an explicit label: kept verbatim (not the default).
		input.Payment.Detalhes = []entity.PagamentoDetalhe{{TPag: "99", XPag: "Boleto bancário", VPag: decimal.NewFromInt(1000)}}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/pag/detPag"), "xPag"), "Boleto bancario"; got != want {
			t.Errorf("explicit xPag = %q, want %q (sanitized)", got, want)
		}

		// tPag 01 and 90 never carry the default.
		input.Payment.Detalhes = []entity.PagamentoDetalhe{{TPag: "01", VPag: decimal.NewFromInt(1000)}}
		doc2, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if doc2.Root().FindElement("./infNFe/pag/detPag/xPag") != nil {
			t.Error("tPag 01 must not carry a default xPag")
		}
		input.Payment.Detalhes = []entity.PagamentoDetalhe{{TPag: "90", VPag: decimal.NewFromInt(0)}}
		doc3, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if doc3.Root().FindElement("./infNFe/pag/detPag/xPag") != nil {
			t.Error("tPag 90 must not carry a default xPag (sem pagamento)")
		}
	})

	t.Run("indPag_omitted_when_zero", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Payment.IndPag = 1 // container value must not leak into detPag
		input.Payment.Detalhes[0].IndPag = 0
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		detPag := findElement(t, doc.Root(), "./infNFe/pag/detPag")
		assertChildOrder(t, detPag, "detPag children", []string{"tPag", "vPag"})
	})

	t.Run("indPag_is_per_detPag", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Payment.IndPag = 0
		input.Payment.Detalhes = []entity.PagamentoDetalhe{
			{IndPag: 1, TPag: "01", VPag: decimal.NewFromInt(500)},
			{IndPag: 2, TPag: "99", XPag: "Outros", VPag: decimal.NewFromInt(500)},
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		detPags := doc.Root().FindElements("./infNFe/pag/detPag")
		if len(detPags) != 2 {
			t.Fatalf("expected 2 detPag elements, got %d", len(detPags))
		}
		for i, want := range []string{"1", "2"} {
			if got := rwElText(detPags[i], "indPag"); got != want {
				t.Errorf("detPag[%d]/indPag = %q, want %q", i, got, want)
			}
		}
	})

	t.Run("cobrancia_can_be_omitted_or_keeps_legacy_default", func(t *testing.T) {
		input := minimalInvoiceInput()
		if input.Cobranca != nil {
			t.Fatal("nil Cobranca should preserve the legacy generated fatura")
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build with legacy default failed: %v", err)
		}
		if doc.Root().FindElement("./infNFe/cobr/fat") == nil {
			t.Fatal("legacy nil setting must continue to emit cobr/fat for a non-zero total")
		}

		input.Cobranca = &entity.CobrancaData{Enabled: false}
		doc, err = builder.Build(input)
		if err != nil {
			t.Fatalf("Build with cobranca disabled failed: %v", err)
		}
		if doc.Root().FindElement("./infNFe/cobr") != nil {
			t.Error("cobr must be omitted when explicitly disabled")
		}
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/total/ICMSTot"), "vNF"), "1000.00"; got != want {
			t.Errorf("disabling cobr changed vNF = %q, want %q", got, want)
		}
	})

	t.Run("veicTransp_UF_omitted_when_empty_and_kept_when_set", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Transport.Veiculo = &entity.VeiculoData{Placa: "GDB6G25"} // no UF
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		veic := findElement(t, doc.Root(), "./infNFe/transp/veicTransp")
		assertChildOrder(t, veic, "veicTransp children", []string{"placa"})

		input.Transport.Veiculo = &entity.VeiculoData{Placa: "GDB6G25", UF: "MT"}
		doc2, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		assertChildOrder(t, findElement(t, doc2.Root(), "./infNFe/transp/veicTransp"), "veicTransp children", []string{"placa", "UF"})
	})
}

// TestBuilder_Devolution_Extras verifies the devolução / delivery
// capabilities (AC5): finNFe honored, NFref validation + block placement,
// <entrega> flat-group sibling of dest.
func TestBuilder_Devolution_Extras(t *testing.T) {
	builder := xml.NewBuilder()

	t.Run("finNFe_honored_and_defaulted", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.FinNFe = "4"
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if got, want := rwElText(findElement(t, doc.Root(), "./infNFe/ide"), "finNFe"), "4"; got != want {
			t.Errorf("finNFe = %q, want %q", got, want)
		}
		// Empty falls back to "1".
		doc2, err := builder.Build(minimalInvoiceInput())
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if got, want := rwElText(findElement(t, doc2.Root(), "./infNFe/ide"), "finNFe"), "1"; got != want {
			t.Errorf("default finNFe = %q, want %q", got, want)
		}
	})

	t.Run("NFref_after_verProc_only_valid_keys", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.NFRefs = []string{
			"51260437218116000101550010001202441561693345", // 44 digits — kept
			"5126043721811600010155001000120244156169334",  // 43 digits — skipped
			"51260437218116000101550010001202441561693X5",  // 44 chars, letter — skipped
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		ide := findElement(t, doc.Root(), "./infNFe/ide")
		assertChildOrder(t, ide, "ide children", []string{"cUF", "cNF", "natOp", "mod", "serie", "nNF", "dhEmi", "dhSaiEnt", "tpNF", "idDest", "cMunFG", "tpImp", "tpEmis", "cDV", "tpAmb", "finNFe", "indFinal", "indPres", "indIntermed", "procEmi", "verProc", "NFref"})
		refs := ide.SelectElements("NFref")
		if len(refs) != 1 {
			t.Fatalf("expected 1 valid NFref, got %d", len(refs))
		}
		if got := rwElText(refs[0], "refNFe"); got != "51260437218116000101550010001202441561693345" {
			t.Errorf("refNFe = %q", got)
		}
	})

	t.Run("NFref_multiple_valid_keys_kept_in_order", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.NFRefs = []string{
			"51260437218116000101550010001202441561693345",
			"51260437218116000101550010001202441561693346",
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		refs := doc.Root().FindElements("./infNFe/ide/NFref")
		if len(refs) != 2 {
			t.Fatalf("expected 2 NFref elements, got %d", len(refs))
		}
		for i, want := range []string{"...3345", "...3346"} {
			got := rwElText(refs[i], "refNFe")
			if !strings.HasSuffix(got, strings.TrimPrefix(want, "...")) {
				t.Errorf("NFref[%d]/refNFe = %q, want suffix %s", i, got, want)
			}
		}
	})

	t.Run("entrega_flat_group_sibling_of_dest", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Entrega = &entity.EntregaData{
			Type:       1,
			CNPJ:       "76061198000185",
			XNome:      "Fazenda Exemplo Dois",
			Logradouro: "FAZENDA MODELO DOIS, GLEBA LOTE 292",
			Numero:     "S/N",
			Bairro:     "PERIMETRO URBANO",
			CodigoMun:  "5106190",
			Municipio:  "NOVA SANTA HELENA",
			UF:         "MT",
			CEP:        "78548000",
			IE:         "31841627867",
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		infNFe := findElement(t, doc.Root(), "./infNFe")
		kids := childTags(infNFe)
		want := []string{"ide", "emit", "dest", "entrega", "det", "total", "transp", "cobr", "pag"}
		if strings.Join(kids, ",") != strings.Join(want, ",") {
			t.Errorf("infNFe children = %v, want %v", kids, want)
		}
		e := infNFe.SelectElement("entrega")
		assertChildOrder(t, e, "entrega children", []string{"CNPJ", "xNome", "xLgr", "nro", "xBairro", "cMun", "xMun", "UF", "CEP", "IE"})
		if got, want := rwElText(e, "xNome"), "Fazenda Exemplo Dois"; got != want {
			t.Errorf("entrega/xNome = %q, want %q", got, want)
		}
	})

	t.Run("entrega_CPF_type_and_optional_fields", func(t *testing.T) {
		input := minimalInvoiceInput()
		input.Entrega = &entity.EntregaData{
			Type:       2,
			CPF:        "52998224725",
			Logradouro: "FAZENDA EXEMPLO",
			Numero:     "1",
			Bairro:     "CENTER",
			CodigoMun:  "5106190",
			Municipio:  "NOVA SANTA HELENA",
			UF:         "MT",
		}
		doc, err := builder.Build(input)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		e := findElement(t, doc.Root(), "./infNFe/entrega")
		assertChildOrder(t, e, "entrega children", []string{"CPF", "xLgr", "nro", "xBairro", "cMun", "xMun", "UF"})
	})
}
