package service_test

import (
	"testing"

	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/service"

	"github.com/shopspring/decimal"
)

func TestDANFEGenerator_Generate(t *testing.T) {
	g := service.NewDANFEGenerator()

	data := fullDANFEData()

	pdfBytes, err := g.Generate(data)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}

	if len(pdfBytes) < 5 || string(pdfBytes[:5]) != "%PDF-" {
		t.Fatalf("output does not look like a PDF, got: %s", string(pdfBytes[:min(20, len(pdfBytes))]))
	}
}

func TestDANFEGenerator_GeneratePreview(t *testing.T) {
	g := service.NewDANFEGenerator()

	data := fullDANFEData()
	pdfBytes, err := g.GeneratePreview(data)
	if err != nil {
		t.Fatalf("unexpected error generating preview PDF: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty preview PDF bytes")
	}
	if len(pdfBytes) < 5 || string(pdfBytes[:5]) != "%PDF-" {
		t.Fatalf("preview output does not look like a PDF")
	}
}

func TestDANFEGenerator_Generate_NoProtocol(t *testing.T) {
	g := service.NewDANFEGenerator()

	data := fullDANFEData()
	data.Protocol = ""
	data.ProtocolDate = ""
	data.CStat = ""
	data.XMotivo = ""

	pdfBytes, err := g.Generate(data)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func fullDANFEData() entity.DANFEData {
	return entity.DANFEData{
		AccessKey:           "51250312345678000190550010000001231234567890",
		EmitterName:         "Fazenda Exemplo LTDA",
		EmitterCNPJ:         "12345678000190",
		EmitterIE:           "123456789",
		EmitterCRT:          "3",
		EmitterAddress:      "Rodovia BR-163",
		EmitterNumber:       "1000",
		EmitterComplement:   "Km 500",
		EmitterNeighborhood: "Zona Rural",
		EmitterCEP:          "78890000",
		EmitterCity:         "Sorriso",
		EmitterUF:           "MT",
		EmitterPhone:        "6635441234",

		DestName:         "Cliente Exemplo S/A",
		DestCNPJ:         "98765432000190",
		DestIE:           "987654321",
		DestIndIEDest:    "1",
		DestAddress:      "Av. Principal",
		DestNumber:       "500",
		DestComplement:   "Sala 101",
		DestNeighborhood: "Centro",
		DestCEP:          "78000000",
		DestCity:         "Cuiabá",
		DestUF:           "MT",
		DestPhone:        "6533224455",

		NaturezaOp:   "Venda de Mercadoria",
		Numero:       123,
		Serie:        1,
		EmissionDate: "15/03/2025 10:30:00",

		Products: []entity.DANFEProduct{
			{
				Code:      "SOJA",
				Desc:      "Soja em Graos",
				NCM:       "12010010",
				CST:       "00",
				CFOP:      "5102",
				Unit:      "KG",
				Quantity:  decimal.NewFromFloat(50000),
				UnitPrice: decimal.NewFromFloat(150),
				Total:     decimal.NewFromFloat(7500000),
				VBC:       decimal.NewFromFloat(7500000),
				PICMS:     decimal.NewFromFloat(17),
				VICMS:     decimal.NewFromFloat(1275000),
				PIPI:      decimal.NewFromFloat(0),
				VIPI:      decimal.NewFromFloat(0),
				PPIS:      decimal.NewFromFloat(0.65),
				VPIS:      decimal.NewFromFloat(48750),
				PCOFINS:   decimal.NewFromFloat(3),
				VCOFINS:   decimal.NewFromFloat(225000),
				VFrete:    decimal.NewFromFloat(0),
				VSeg:      decimal.NewFromFloat(0),
				VDesc:     decimal.NewFromFloat(0),
				VOutro:    decimal.NewFromFloat(0),
			},
		},

		TotalValue: decimal.NewFromFloat(7500000),
		VBC:        decimal.NewFromFloat(7500000),
		VICMS:      decimal.NewFromFloat(1275000),
		VICMSDeson: decimal.NewFromFloat(0),
		VBCST:      decimal.NewFromFloat(0),
		VST:        decimal.NewFromFloat(0),
		VII:        decimal.NewFromFloat(0),
		VIPI:       decimal.NewFromFloat(0),
		VPIS:       decimal.NewFromFloat(48750),
		VCOFINS:    decimal.NewFromFloat(225000),
		VFrete:     decimal.NewFromFloat(0),
		VSeg:       decimal.NewFromFloat(0),
		VDesc:      decimal.NewFromFloat(0),
		VOutro:     decimal.NewFromFloat(0),
		VTotTrib:   decimal.NewFromFloat(0),

		ModFrete:      "9",
		TranspName:    "Transportadora Exemplo LTDA",
		TranspCNPJ:    "11223344000155",
		TranspIE:      "112233445",
		TranspAddress: "Rua dos Transportes, 100",
		TranspCity:    "Sorriso",
		TranspUF:      "MT",
		QVol:          "1",
		Esp:           "Granel",
		Marca:         "",
		NVol:          "",
		PesoL:         decimal.NewFromFloat(50000),
		PesoB:         decimal.NewFromFloat(50020),

		InfCpl:     "Entrega conforme contrato nº 12345. Peso aferido na balança.",
		InfAdFisco: "",

		Protocol:     "351250123456789",
		ProtocolDate: "15/03/2025 10:31:00",
		CStat:        "100",
		XMotivo:      "Autorizado o uso da NF-e",
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
