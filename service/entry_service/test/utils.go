package entry_service_test

import (
	entity_public "armazenda/entity/public"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// CreateBasicTestEntry creates a minimal valid entry for testing
func CreateBasicTestEntry() entity_public.Entry {
	return entity_public.Entry{
		Field:   1,
		Crop:    1,
		Vehicle: 1,
		CargoWeight: entity_public.CargoWeight{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(100),
		},
		ArrivalDate: time.Now(),
		Farm:        1,
	}
}

func CreateTestEntry(entry entity_public.Entry) entity_public.Entry {
	defaultEntry := entity_public.Entry{
		Field:   1,
		Crop:    1,
		Vehicle: 1,
		CargoWeight: entity_public.CargoWeight{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(100),
		},
		ArrivalDate: time.Now(),
		Farm:        1,
	}

	defaultValue := reflect.ValueOf(defaultEntry)
	providedValue := reflect.ValueOf(entry)

	for i := 0; i < defaultValue.NumField(); i++ {
		providedField := providedValue.Field(i)
		if !providedField.IsZero() {
			defaultValue.Field(i).Set(providedField)
		}
	}

	return defaultValue.Interface().(entity_public.Entry)
}

// CreateTestDisplayEntryWithNetWeight creates a display entry with specified ID and NetWeight
func CreateTestDisplayEntryWithNetWeight(id uint32, netWeight decimal.Decimal) entity_public.DisplayEntry {
	return entity_public.DisplayEntry{
		Id:        id,
		NetWeight: netWeight,
	}
}

// CreateTestCrop creates a test crop
func CreateTestCrop(id uint8, productId uint8, farm uint32) entity_public.Crop {
	return entity_public.Crop{
		Id:      id,
		Name:    "Test Crop",
		Product: productId,
		Farm:    farm,
	}
}

// CreateTestProduct creates a test product
func CreateTestProduct(id uint8, name string) entity_public.Product {
	return entity_public.Product{
		Id:   id,
		Name: name,
	}
}

// CreateTestPersonConfig creates a test person config
func CreateTestPersonConfig() entity_public.PersonConfig {
	return entity_public.PersonConfig{
		HumidityDiscount:  decimal.NewFromFloat(1.15),
		EntrySoyDiscount:  decimal.NewFromFloat(3.5),
		EntryCornDiscount: decimal.NewFromFloat(5.5),
	}
}

func AssertNetWeightEquals(t *testing.T, expected decimal.Decimal, actual decimal.Decimal) {
	if !expected.Equal(actual) {
		t.Errorf("Expected NetWeight %s, got %s", expected.String(), actual.String())
	}
}
