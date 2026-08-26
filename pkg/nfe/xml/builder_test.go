package xml_test

import (
	"strings"
	"testing"
	"time"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/xml"

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
						PIBS:       decimal.NewFromFloat(0.10),
						VIBS:       decimal.NewFromFloat(1.00),
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
		"<gIBSUF>",
		"<pIBSUF>0.1000</pIBSUF>",
		"<vIBSUF>1.00</vIBSUF>",
		"<gIBSMun>",
		"<pIBSMun>0.0000</pIBSMun>",
		"<vIBSMun>0.00</vIBSMun>",
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

	// Per-NF-e totals: vBCIBSCBS + gIBS (with vIBSUF + vIBSMun + vIBS) + gCBS (with vCBS)
	for _, want := range []string{
		"<vBCIBSCBS>1000.00</vBCIBSCBS>",
		"<vIBSUF>1.00</vIBSUF>",
		"<vIBSMun>0.00</vIBSMun>",
		"<vIBS>1.00</vIBS>",
		"<vCBS>9.00</vCBS>",
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("expected %q in XML totals", want)
		}
	}
}
