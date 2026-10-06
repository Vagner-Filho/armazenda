package nfe_model_test

import (
	"testing"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

func TestDetachedInvoiceItem_JSONSerialization(t *testing.T) {
	cest := "0100101"
	item := nfe_model.DetachedInvoiceItem{
		ProductName: "Soja",
		NCM:         "12019000",
		CEST:        &cest,
		CFOP:        "5101",
		Unit:        "KG",
		Quantity:    decimal.NewFromFloat(10000.5),
		GrossWeight: decimal.NewFromFloat(10200.75),
		UnitPrice:   decimal.NewFromFloat(150.25),
		TotalValue:  decimal.NewFromFloat(1502507.5),
	}

	// Test that fields are set correctly
	if item.ProductName != "Soja" {
		t.Errorf("expected ProductName 'Soja', got '%s'", item.ProductName)
	}
	if item.NCM != "12019000" {
		t.Errorf("expected NCM '12019000', got '%s'", item.NCM)
	}
	if item.CEST == nil || *item.CEST != "0100101" {
		t.Errorf("expected CEST '0100101', got '%v'", item.CEST)
	}
	if !item.Quantity.Equal(decimal.NewFromFloat(10000.5)) {
		t.Errorf("expected Quantity 10000.5, got %s", item.Quantity)
	}
	if !item.TotalValue.Equal(decimal.NewFromFloat(1502507.5)) {
		t.Errorf("expected TotalValue 1502507.5, got %s", item.TotalValue)
	}
}

func TestDetachedInvoiceItem_WithFarmProductID(t *testing.T) {
	farmProductID := uint16(42)
	item := nfe_model.DetachedInvoiceItem{
		FarmProductID: &farmProductID,
		ProductName:   "Milho",
		NCM:           "10059000",
		CFOP:          "5101",
		Unit:          "KG",
		Quantity:      decimal.NewFromFloat(5000.0),
		GrossWeight:   decimal.NewFromFloat(5100.0),
		UnitPrice:     decimal.NewFromFloat(120.0),
		TotalValue:    decimal.NewFromFloat(600000.0),
	}

	if item.FarmProductID == nil || *item.FarmProductID != 42 {
		t.Errorf("expected FarmProductID 42, got %v", item.FarmProductID)
	}
}

func TestDetachedInvoice_StructFields(t *testing.T) {
	naturezaOp := "Venda de produção"
	modFrete := 1
	icmsCST := "00"
	pisCST := "01"
	cofinsCST := "01"
	ibsCST := "000"
	cbsCST := "000"
	cClassTrib := "000001"
	infCpl := "Informações complementares de teste"

	inv := nfe_model.DetachedInvoice{
		ID:          1,
		FarmID:      10,
		RecipientID: 5,
		AccessKey:   "12345678901234567890123456789012345678901234",
		Serie:       1,
		Number:      42,
		Status:      "draft",
		NaturezaOp:  &naturezaOp,
		ModFrete:    &modFrete,
		TotalValue:  decimal.NewFromFloat(100000.0),
		IBSValue:    decimal.NewFromFloat(100.0),
		CBSValue:    decimal.NewFromFloat(900.0),
		ICMSCST:     &icmsCST,
		PISCST:      &pisCST,
		COFINSCST:   &cofinsCST,
		IBSCST:      &ibsCST,
		CBSCST:      &cbsCST,
		CClassTrib:  &cClassTrib,
		InfCpl:      &infCpl,
		TpEmis:      1,
		RetryCount:  0,
	}

	// Test basic fields
	if inv.ID != 1 {
		t.Errorf("expected ID 1, got %d", inv.ID)
	}
	if inv.FarmID != 10 {
		t.Errorf("expected FarmID 10, got %d", inv.FarmID)
	}
	if inv.RecipientID != 5 {
		t.Errorf("expected RecipientID 5, got %d", inv.RecipientID)
	}
	if inv.Status != "draft" {
		t.Errorf("expected Status 'draft', got '%s'", inv.Status)
	}
	if inv.Serie != 1 {
		t.Errorf("expected Serie 1, got %d", inv.Serie)
	}
	if inv.Number != 42 {
		t.Errorf("expected Number 42, got %d", inv.Number)
	}

	// Test optional fields
	if inv.NaturezaOp == nil || *inv.NaturezaOp != "Venda de produção" {
		t.Errorf("expected NaturezaOp 'Venda de produção', got %v", inv.NaturezaOp)
	}
	if inv.ModFrete == nil || *inv.ModFrete != 1 {
		t.Errorf("expected ModFrete 1, got %v", inv.ModFrete)
	}

	// Test tax CSTs
	if inv.ICMSCST == nil || *inv.ICMSCST != "00" {
		t.Errorf("expected ICMSCST '00', got %v", inv.ICMSCST)
	}
	if inv.IBSCST == nil || *inv.IBSCST != "000" {
		t.Errorf("expected IBSCST '000', got %v", inv.IBSCST)
	}

	// Test totals
	if !inv.TotalValue.Equal(decimal.NewFromFloat(100000.0)) {
		t.Errorf("expected TotalValue 100000.0, got %s", inv.TotalValue)
	}
	if !inv.IBSValue.Equal(decimal.NewFromFloat(100.0)) {
		t.Errorf("expected IBSValue 100.0, got %s", inv.IBSValue)
	}
	if !inv.CBSValue.Equal(decimal.NewFromFloat(900.0)) {
		t.Errorf("expected CBSValue 900.0, got %s", inv.CBSValue)
	}
}

func TestDetachedInvoice_WithTaxRates(t *testing.T) {
	icmsRate := decimal.NewFromFloat(0.17)
	pisRate := decimal.NewFromFloat(0.0165)
	cofinsRate := decimal.NewFromFloat(0.076)
	ibsRate := decimal.NewFromFloat(0.001)
	cbsRate := decimal.NewFromFloat(0.009)

	inv := nfe_model.DetachedInvoice{
		ID:        1,
		FarmID:    10,
		AccessKey: "12345678901234567890123456789012345678901234",
		Status:    "authorized",
		TaxRates: &entity.TaxRates{
			ICMSRate:   &icmsRate,
			PISRate:    &pisRate,
			COFINSRate: &cofinsRate,
			IBSRate:    &ibsRate,
			CBSRate:    &cbsRate,
		},
	}

	if inv.TaxRates == nil {
		t.Fatal("expected TaxRates to be set")
	}
	if inv.TaxRates.ICMSRate == nil || !inv.TaxRates.ICMSRate.Equal(decimal.NewFromFloat(0.17)) {
		t.Errorf("expected ICMSRate 0.17, got %v", inv.TaxRates.ICMSRate)
	}
	if inv.TaxRates.CBSRate == nil || !inv.TaxRates.CBSRate.Equal(decimal.NewFromFloat(0.009)) {
		t.Errorf("expected CBSRate 0.009, got %v", inv.TaxRates.CBSRate)
	}
}

func TestDetachedInvoice_WithItems(t *testing.T) {
	items := []nfe_model.DetachedInvoiceItem{
		{
			ProductName: "Soja",
			NCM:         "12019000",
			CFOP:        "5101",
			Unit:        "KG",
			Quantity:    decimal.NewFromFloat(10000.0),
			GrossWeight: decimal.NewFromFloat(10200.0),
			UnitPrice:   decimal.NewFromFloat(150.0),
			TotalValue:  decimal.NewFromFloat(1500000.0),
		},
		{
			ProductName: "Milho",
			NCM:         "10059000",
			CFOP:        "5101",
			Unit:        "KG",
			Quantity:    decimal.NewFromFloat(5000.0),
			GrossWeight: decimal.NewFromFloat(5100.0),
			UnitPrice:   decimal.NewFromFloat(120.0),
			TotalValue:  decimal.NewFromFloat(600000.0),
		},
	}

	inv := nfe_model.DetachedInvoice{
		ID:        1,
		FarmID:    10,
		AccessKey: "12345678901234567890123456789012345678901234",
		Status:    "draft",
		Items:     items,
	}

	if len(inv.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(inv.Items))
	}
	if inv.Items[0].ProductName != "Soja" {
		t.Errorf("expected first item 'Soja', got '%s'", inv.Items[0].ProductName)
	}
	if inv.Items[1].ProductName != "Milho" {
		t.Errorf("expected second item 'Milho', got '%s'", inv.Items[1].ProductName)
	}
	if !inv.Items[0].TotalValue.Equal(decimal.NewFromFloat(1500000.0)) {
		t.Errorf("expected first item total 1500000.0, got %s", inv.Items[0].TotalValue)
	}
}
