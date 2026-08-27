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
		if !strings.Contains(xmlStr, "<infCpl>CND Fazenda:; Cert. Nº 222222, válido até 02/08/2026</infCpl>") {
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
		if !strings.Contains(xmlStr, "<xNome>Armazém; Tropical  Ltda</xNome>") {
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
						VIBSMun:    decimal.Zero,                  // municipal share (zero in 2026 phase)
						VIBS:       decimal.NewFromFloat(1.00),   // per-item total = UF + VIBSMun
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
