package nfe_service

import (
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// TestBuildDetachedItems_InfAdProd verifies the per-item additional
// information flows from the emission input into both the XML entity
// (ItemData.InfAdProd, emitted as the last <det> child) and the JSONB row,
// with a blank input staying blank (the builder then omits the element).
func TestBuildDetachedItems_InfAdProd(t *testing.T) {
	svc := NewNFeService()

	cfg := &entity_public.FarmConfig{
		TaxRegime:       3, // normal (non-Simples) regime
		DefaultModFrete: 9,
		ICMSRate:        decimal.RequireFromString("0.17"),
		PISRate:         decimal.RequireFromString("0.01"),
		COFINSRate:      decimal.RequireFromString("0.07"),
		IBSRate:         decimal.RequireFromString("0.001"),
		CBSRate:         decimal.RequireFromString("0.009"),
	}

	inputs := []DetachedItemInput{
		{
			ProductName: "Soja",
			CFOP:        "5101",
			Quantity:    decimal.RequireFromString("10"),
			UnitPrice:   decimal.RequireFromString("150"),
			Unit:        "KG",
			InfAdProd:   "SERIE ABC123",
		},
		{
			ProductName: "Milho",
			CFOP:        "5102",
			Quantity:    decimal.RequireFromString("5"),
			UnitPrice:   decimal.RequireFromString("120"),
			Unit:        "KG",
		},
	}

	items, itemsForDB, _, _, _ := svc.buildDetachedItems(
		inputs, cfg, resolvedSnapshot(t, entity.TaxRates{}, nil, cfg), nil)

	if len(items) != 2 || len(itemsForDB) != 2 {
		t.Fatalf("expected 2 items, got %d/%d", len(items), len(itemsForDB))
	}
	if items[0].InfAdProd != "SERIE ABC123" {
		t.Errorf("ItemData.InfAdProd = %q, want %q", items[0].InfAdProd, "SERIE ABC123")
	}
	if items[1].InfAdProd != "" {
		t.Errorf("blank input must map to an empty InfAdProd, got %q", items[1].InfAdProd)
	}

	// The persisted row (items_json) carries the same values.
	if itemsForDB[0].InfAdProd != "SERIE ABC123" {
		t.Errorf("items_json InfAdProd = %q, want %q", itemsForDB[0].InfAdProd, "SERIE ABC123")
	}
	if itemsForDB[1].InfAdProd != "" {
		t.Errorf("blank input must persist an empty InfAdProd, got %q", itemsForDB[1].InfAdProd)
	}

	// The preview/print DANFE mapping carries the note per item (the DANFE
	// generator prints it under the item description).
	danfeData := svc.buildDetachedDANFEData(entity.InvoiceInput{Items: items})
	if len(danfeData.Products) != 2 {
		t.Fatalf("expected 2 DANFE products, got %d", len(danfeData.Products))
	}
	if danfeData.Products[0].InfAdProd != "SERIE ABC123" {
		t.Errorf("DANFE product InfAdProd = %q, want %q", danfeData.Products[0].InfAdProd, "SERIE ABC123")
	}
	if danfeData.Products[1].InfAdProd != "" {
		t.Errorf("blank input must map to an empty DANFE InfAdProd, got %q", danfeData.Products[1].InfAdProd)
	}
}
