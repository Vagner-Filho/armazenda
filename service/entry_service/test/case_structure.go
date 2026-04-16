package entry_service_test

import (
	entity_public "armazenda/entity/public"
	"testing"

	"github.com/shopspring/decimal"
)

type addEntryTestCase struct {
	name                 string
	entry                entity_public.Entry
	setupMocks           func(*MockEntryModel, *MockPersonModel, *MockProductModel, *MockCropModel, *MockHumidityProgressionModel)
	expectedToastType    entity_public.ToastType
	expectedToastMessage string
	expectedNetWeight    decimal.Decimal
	validateResult       func(*testing.T, addEntryTestCase, entity_public.DisplayEntry, *MockEntryModel, *MockPersonModel, *MockProductModel, *MockCropModel, *MockHumidityProgressionModel)
}
