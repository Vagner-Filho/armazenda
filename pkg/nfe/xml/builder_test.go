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
			CNPJ:       "12345678000195",
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
					Codigo:  "1",
					CEAN:    "SEM GTIN",
					XProd:   "Milho",
					NCM:     "10059010",
					CFOP:    "5102",
					UCom:    "KG",
					QCom:    decimal.NewFromInt(100),
					VUnCom:  decimal.NewFromInt(10),
					VProd:   decimal.NewFromInt(1000),
					CEANTrib: "SEM GTIN",
					UTrib:   "KG",
					QTrib:   decimal.NewFromInt(100),
					VUnTrib: decimal.NewFromInt(10),
					IndTot:  1,
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
