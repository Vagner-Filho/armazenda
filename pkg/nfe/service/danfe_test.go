package service_test

import (
	"testing"

	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/service"

	"github.com/shopspring/decimal"
)

func TestDANFEGenerator_Generate(t *testing.T) {
	g := service.NewDANFEGenerator()

	data := entity.DANFEData{
		AccessKey:      "51250312345678000190550010000001231234567890",
		EmitterName:    "Fazenda Exemplo LTDA",
		EmitterCNPJ:    "12345678000190",
		EmitterAddress: "Rodovia BR-163, 1000, Zona Rural",
		EmitterCity:    "Sorriso",
		EmitterUF:      "MT",
		DestName:       "Cliente Exemplo S/A",
		DestCNPJ:       "98765432000190",
		DestAddress:    "Av. Principal, 500, Centro",
		DestCity:       "Cuiabá",
		DestUF:         "MT",
		NaturezaOp:     "Venda de Mercadoria",
		Numero:         123,
		Serie:          1,
		EmissionDate:   "15/03/2025 10:30:00",
		Products: []entity.DANFEProduct{
			{
				Code:      "SOJA",
				Desc:      "Soja em Graos",
				NCM:       "12010010",
				CFOP:      "5102",
				Unit:      "KG",
				Quantity:  decimal.NewFromFloat(50000),
				UnitPrice: decimal.NewFromFloat(150),
				Total:     decimal.NewFromFloat(7500000),
			},
		},
		TotalValue:   decimal.NewFromFloat(7500000),
		ICMSValue:    decimal.NewFromFloat(1275000),
		Protocol:     "351250123456789",
		ProtocolDate: "15/03/2025 10:31:00",
	}

	pdfBytes, err := g.Generate(data)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}

	// Very basic sanity check: PDF files start with "%PDF-"
	if len(pdfBytes) < 5 || string(pdfBytes[:5]) != "%PDF-" {
		t.Fatalf("output does not look like a PDF, got: %s", string(pdfBytes[:min(20, len(pdfBytes))]))
	}
}

func TestDANFEGenerator_Generate_NoProtocol(t *testing.T) {
	g := service.NewDANFEGenerator()

	data := entity.DANFEData{
		AccessKey:      "51250312345678000190550010000001231234567890",
		EmitterName:    "Fazenda Exemplo LTDA",
		EmitterCNPJ:    "12345678000190",
		EmitterAddress: "Rodovia BR-163, 1000",
		EmitterCity:    "Sorriso",
		EmitterUF:      "MT",
		DestName:       "Cliente Exemplo S/A",
		DestCNPJ:       "98765432000190",
		DestAddress:    "Av. Principal, 500",
		DestCity:       "Cuiabá",
		DestUF:         "MT",
		NaturezaOp:     "Venda de Mercadoria",
		Numero:         123,
		Serie:          1,
		EmissionDate:   "15/03/2025 10:30:00",
		Products: []entity.DANFEProduct{
			{
				Code:      "SOJA",
				Desc:      "Soja em Graos",
				NCM:       "12010010",
				CFOP:      "5102",
				Unit:      "KG",
				Quantity:  decimal.NewFromFloat(50000),
				UnitPrice: decimal.NewFromFloat(150),
				Total:     decimal.NewFromFloat(7500000),
			},
		},
		TotalValue: decimal.NewFromFloat(7500000),
		ICMSValue:  decimal.NewFromFloat(1275000),
	}

	pdfBytes, err := g.Generate(data)
	if err != nil {
		t.Fatalf("unexpected error generating PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
