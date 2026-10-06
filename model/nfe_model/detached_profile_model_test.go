package nfe_model_test

import (
	"encoding/json"
	"testing"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

func TestDetachedProfileItem_JSONRoundTrip(t *testing.T) {
	cest := "0100101"
	farmProductID := uint16(42)

	items := []nfe_model.DetachedProfileItem{
		{
			FarmProductID: &farmProductID,
			ProductName:   "Soja",
			NCM:           "12019000",
			CEST:          &cest,
			CFOP:          "5101",
			Unit:          "KG",
			UnitPrice:     decimal.RequireFromString("150.25"),
		},
		{
			ProductName: "Milho",
			NCM:         "10059000",
			CFOP:        "5102",
			Unit:        "SC",
			UnitPrice:   decimal.RequireFromString("120.75"),
		},
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded []nfe_model.DetachedProfileItem
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 items, got %d", len(decoded))
	}
	if decoded[0].CFOP != "5101" || decoded[1].CFOP != "5102" {
		t.Errorf("expected distinct CFOPs 5101/5102, got %q/%q", decoded[0].CFOP, decoded[1].CFOP)
	}
	if decoded[0].FarmProductID == nil || *decoded[0].FarmProductID != 42 {
		t.Errorf("expected farm product 42 to round-trip, got %v", decoded[0].FarmProductID)
	}
	if decoded[1].FarmProductID != nil {
		t.Errorf("expected nil farm product to round-trip, got %v", decoded[1].FarmProductID)
	}
	if !decoded[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
		t.Errorf("expected exact price 150.25, got %s", decoded[0].UnitPrice)
	}
	if !decoded[1].UnitPrice.Equal(decimal.RequireFromString("120.75")) {
		t.Errorf("expected exact price 120.75, got %s", decoded[1].UnitPrice)
	}
	if decoded[0].CEST == nil || *decoded[0].CEST != "0100101" {
		t.Errorf("expected CEST to round-trip, got %v", decoded[0].CEST)
	}
}

func TestDetachedProfile_NilRateIsDistinctFromExplicitZero(t *testing.T) {
	zero := decimal.Zero
	icms := decimal.RequireFromString("0.17")

	unset := nfe_model.DetachedProfile{Name: "Sem alíquotas"}
	explicit := nfe_model.DetachedProfile{
		Name: "Com alíquotas",
		TaxRates: &entity.TaxRates{
			ICMSRate: &icms,
			PISRate:  &zero,
		},
	}

	if unset.TaxRates != nil {
		t.Fatal("expected nil TaxRates for an unset profile")
	}
	if explicit.TaxRates == nil || explicit.TaxRates.ICMSRate == nil || explicit.TaxRates.PISRate == nil {
		t.Fatal("expected explicit rates to be set")
	}
	if !explicit.TaxRates.PISRate.IsZero() {
		t.Errorf("expected explicit zero PIS rate, got %s", explicit.TaxRates.PISRate)
	}
	if !explicit.TaxRates.ICMSRate.Equal(decimal.RequireFromString("0.17")) {
		t.Errorf("expected exact ICMS rate 0.17, got %s", explicit.TaxRates.ICMSRate)
	}
}

func TestFarmProductIDsFromProfileItems(t *testing.T) {
	a := uint16(1)
	b := uint16(2)

	items := []nfe_model.DetachedProfileItem{
		{FarmProductID: &a},
		{FarmProductID: nil},
		{FarmProductID: &a},
		{FarmProductID: &b},
	}

	ids := nfe_model.FarmProductIDsFromProfileItems(items)
	if len(ids) != 2 {
		t.Fatalf("expected 2 distinct ids, got %v", ids)
	}
	if ids[0] != 1 || ids[1] != 2 {
		t.Errorf("expected ids [1 2], got %v", ids)
	}
}

func TestDetachedInvoiceItems_PerItemCFOPJSON(t *testing.T) {
	items := []nfe_model.DetachedInvoiceItem{
		{ProductName: "Soja", NCM: "12019000", CFOP: "5101", Unit: "KG", Quantity: decimal.NewFromInt(10), UnitPrice: decimal.RequireFromString("150.25"), TotalValue: decimal.RequireFromString("1502.5")},
		{ProductName: "Milho", NCM: "10059000", CFOP: "5102", Unit: "KG", Quantity: decimal.NewFromInt(5), UnitPrice: decimal.RequireFromString("120.75"), TotalValue: decimal.RequireFromString("603.75")},
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded []nfe_model.DetachedInvoiceItem
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded[0].CFOP != "5101" || decoded[1].CFOP != "5102" {
		t.Errorf("expected distinct item CFOPs to round-trip, got %q/%q", decoded[0].CFOP, decoded[1].CFOP)
	}
	if !decoded[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
		t.Errorf("expected exact price 150.25, got %s", decoded[0].UnitPrice)
	}
}
