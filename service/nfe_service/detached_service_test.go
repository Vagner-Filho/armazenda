package nfe_service_test

import (
	"testing"

	"armazenda/pkg/nfe/entity"
	"armazenda/service/nfe_service"

	"github.com/shopspring/decimal"
)

func TestDetachedRecipient_WithPersonID(t *testing.T) {
	personID := uint32(42)
	recipient := nfe_service.DetachedRecipient{
		PersonID: &personID,
	}

	if recipient.PersonID == nil || *recipient.PersonID != 42 {
		t.Errorf("expected PersonID 42, got %v", recipient.PersonID)
	}
}

func TestDetachedRecipient_InlineFields(t *testing.T) {
	ie := "123456789"
	street := "Rua Teste"
	number := "123"
	neighborhood := "Centro"
	city := "Cuiabá"
	state := "MT"
	cep := "78000000"
	phone := "65999999999"
	email := "test@example.com"

	recipient := nfe_service.DetachedRecipient{
		Name:         "Empresa Teste LTDA",
		Document:     "12345678000195",
		IE:           &ie,
		Street:       &street,
		Number:       &number,
		Neighborhood: &neighborhood,
		City:         &city,
		State:        &state,
		CEP:          &cep,
		PhoneNumber:  &phone,
		Email:        &email,
	}

	if recipient.Name != "Empresa Teste LTDA" {
		t.Errorf("expected Name 'Empresa Teste LTDA', got '%s'", recipient.Name)
	}
	if recipient.Document != "12345678000195" {
		t.Errorf("expected Document '12345678000195', got '%s'", recipient.Document)
	}
	if recipient.IE == nil || *recipient.IE != "123456789" {
		t.Errorf("expected IE '123456789', got %v", recipient.IE)
	}
	if recipient.City == nil || *recipient.City != "Cuiabá" {
		t.Errorf("expected City 'Cuiabá', got %v", recipient.City)
	}
	if recipient.State == nil || *recipient.State != "MT" {
		t.Errorf("expected State 'MT', got %v", recipient.State)
	}
}

func TestDetachedItemInput_WithFarmProductID(t *testing.T) {
	farmProductID := uint16(10)
	cest := "0100101"

	item := nfe_service.DetachedItemInput{
		FarmProductID: &farmProductID,
		CEST:          &cest,
		Quantity:      decimal.NewFromFloat(10000.0),
		GrossWeight:   decimal.NewFromFloat(10200.0),
		UnitPrice:     decimal.NewFromFloat(150.0),
		Unit:          "KG",
	}

	if item.FarmProductID == nil || *item.FarmProductID != 10 {
		t.Errorf("expected FarmProductID 10, got %v", item.FarmProductID)
	}
	if !item.Quantity.Equal(decimal.NewFromFloat(10000.0)) {
		t.Errorf("expected Quantity 10000.0, got %s", item.Quantity)
	}
	if !item.UnitPrice.Equal(decimal.NewFromFloat(150.0)) {
		t.Errorf("expected UnitPrice 150.0, got %s", item.UnitPrice)
	}
}

func TestDetachedItemInput_ManualProduct(t *testing.T) {
	cest := "0100101"
	item := nfe_service.DetachedItemInput{
		ProductName: "Soja em Grão",
		NCM:         "12019000",
		CEST:        &cest,
		Quantity:    decimal.NewFromFloat(25000.0),
		GrossWeight: decimal.NewFromFloat(25500.0),
		UnitPrice:   decimal.NewFromFloat(145.50),
		Unit:        "KG",
	}

	if item.ProductName != "Soja em Grão" {
		t.Errorf("expected ProductName 'Soja em Grão', got '%s'", item.ProductName)
	}
	if item.NCM != "12019000" {
		t.Errorf("expected NCM '12019000', got '%s'", item.NCM)
	}
	if item.FarmProductID != nil {
		t.Errorf("expected FarmProductID to be nil, got %v", item.FarmProductID)
	}
}

func TestDetachedInvoiceInput_SingleItem(t *testing.T) {
	personID := uint32(5)
	vehicleID := uint16(3)

	input := nfe_service.DetachedInvoiceInput{
		FarmID: 10,
		Recipient: nfe_service.DetachedRecipient{
			PersonID: &personID,
		},
		VehicleID: &vehicleID,
		Items: []nfe_service.DetachedItemInput{
			{
				ProductName: "Soja",
				NCM:         "12019000",
				CFOP:        "5101",
				Quantity:    decimal.NewFromFloat(10000.0),
				GrossWeight: decimal.NewFromFloat(10200.0),
				UnitPrice:   decimal.NewFromFloat(150.0),
				Unit:        "KG",
			},
		},
	}

	if input.FarmID != 10 {
		t.Errorf("expected FarmID 10, got %d", input.FarmID)
	}
	if input.Recipient.PersonID == nil || *input.Recipient.PersonID != 5 {
		t.Errorf("expected Recipient.PersonID 5, got %v", input.Recipient.PersonID)
	}
	if input.VehicleID == nil || *input.VehicleID != 3 {
		t.Errorf("expected VehicleID 3, got %v", input.VehicleID)
	}
	if len(input.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(input.Items))
	}
}

func TestDetachedInvoiceInput_MultipleItems(t *testing.T) {
	input := nfe_service.DetachedInvoiceInput{
		FarmID: 10,
		Recipient: nfe_service.DetachedRecipient{
			Name:     "Cliente Teste",
			Document: "12345678901",
		},
		Items: []nfe_service.DetachedItemInput{
			{
				ProductName: "Soja",
				NCM:         "12019000",
				CFOP:        "5101",
				Quantity:    decimal.NewFromFloat(10000.0),
				GrossWeight: decimal.NewFromFloat(10200.0),
				UnitPrice:   decimal.NewFromFloat(150.0),
				Unit:        "KG",
			},
			{
				ProductName: "Milho",
				NCM:         "10059000",
				CFOP:        "5102",
				Quantity:    decimal.NewFromFloat(5000.0),
				GrossWeight: decimal.NewFromFloat(5100.0),
				UnitPrice:   decimal.NewFromFloat(120.0),
				Unit:        "KG",
			},
		},
	}

	if len(input.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(input.Items))
	}
	if input.Items[0].ProductName != "Soja" {
		t.Errorf("expected first item 'Soja', got '%s'", input.Items[0].ProductName)
	}
	if input.Items[1].ProductName != "Milho" {
		t.Errorf("expected second item 'Milho', got '%s'", input.Items[1].ProductName)
	}
}

func TestDetachedInvoiceInput_WithTaxRates(t *testing.T) {
	icmsRate := decimal.NewFromFloat(0.17)
	pisRate := decimal.NewFromFloat(0.0165)
	cofinsRate := decimal.NewFromFloat(0.076)
	ibsRate := decimal.NewFromFloat(0.001)
	cbsRate := decimal.NewFromFloat(0.009)

	input := nfe_service.DetachedInvoiceInput{
		FarmID: 10,
		Recipient: nfe_service.DetachedRecipient{
			Name:     "Cliente Teste",
			Document: "12345678901",
		},
		TaxRates: entity.TaxRates{
			ICMSRate:   &icmsRate,
			PISRate:    &pisRate,
			COFINSRate: &cofinsRate,
			IBSRate:    &ibsRate,
			CBSRate:    &cbsRate,
		},
		Items: []nfe_service.DetachedItemInput{
			{
				ProductName: "Soja",
				NCM:         "12019000",
				CFOP:        "5101",
				Quantity:    decimal.NewFromFloat(10000.0),
				GrossWeight: decimal.NewFromFloat(10200.0),
				UnitPrice:   decimal.NewFromFloat(150.0),
				Unit:        "KG",
			},
		},
	}

	if input.TaxRates.ICMSRate == nil || !input.TaxRates.ICMSRate.Equal(decimal.NewFromFloat(0.17)) {
		t.Errorf("expected ICMSRate 0.17, got %v", input.TaxRates.ICMSRate)
	}
	if input.TaxRates.CBSRate == nil || !input.TaxRates.CBSRate.Equal(decimal.NewFromFloat(0.009)) {
		t.Errorf("expected CBSRate 0.009, got %v", input.TaxRates.CBSRate)
	}
}
