package entry_service_test

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/service/entry_service"
	"testing"

	"github.com/shopspring/decimal"
)

var errorCases []addEntryTestCase = []addEntryTestCase{
	{
		name: "Error - Negative net weight (tare > gross weight)",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.CargoWeight.Tare = decimal.NewFromInt(2000)
			return entry
		}(),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return entity_public.DisplayEntry{}, &model_error.ModelError{
					Message:     "Connection failed",
					IsServerErr: true,
				}
			}
		},
		expectedToastType:    entity_public.ErrorToast,
		expectedToastMessage: "O peso líquido não pode ser menor do que zero",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if em.AddEntryCalled {
				t.Error("Expected AddEntry NOT to be called for negative net weight")
			}
		},
	},
	{
		name: "Error - Humidity discount calculation fails",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			humidity := decimal.NewFromInt(16)
			entry.Humidity = &humidity
			return entry
		}(),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return decimal.Zero, &model_error.ModelError{Message: "Database error", IsServerErr: true}
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return entity_public.DisplayEntry{}, &model_error.ModelError{
					Message:     "Connection failed",
					IsServerErr: true,
				}
			}
		},
		expectedToastType:    entity_public.ErrorToast,
		expectedToastMessage: "Falha ao calcular desconto de humidade",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !pm.GetHumidityDiscountCalled {
				t.Error("Expected GetHumidityDiscount to be called")
			}
			if em.AddEntryCalled {
				t.Error("Expected AddEntry NOT to be called after humidity discount error")
			}
		},
	},
	{
		name:  "Error - AddEntry returns server error",
		entry: CreateBasicTestEntry(),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return entity_public.DisplayEntry{}, &model_error.ModelError{
					Message:     "Connection failed",
					IsServerErr: true,
				}
			}
		},
		expectedToastType:    entity_public.ErrorToast,
		expectedToastMessage: "Houve um erro interno ao adicionar a entrada",
	},
	{
		name:  "Error - AddEntry returns user error",
		entry: CreateBasicTestEntry(),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return entity_public.DisplayEntry{}, &model_error.ModelError{
					Message:     "Field not found",
					IsServerErr: false,
				}
			}
		},
		expectedToastType:    entity_public.WarningToast,
		expectedToastMessage: "Field not found",
	},
}

var successCases []addEntryTestCase = []addEntryTestCase{
	{
		name:              "Success - Basic entry (no origin, no quality discounts, no service tax)",
		entry:             CreateBasicTestEntry(),
		expectedNetWeight: decimal.NewFromFloat(900.0),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}
			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax NOT to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with exceeding humidity limit by 2% and default discount for farm",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(24.425),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if !pm.GetHumidityDiscountCalled {
				t.Error("Expected GetHumidityDiscount to be called")
			}
			if em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax NOT to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with exceeding humidity limit by 2% and default discount for person",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(24.150),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if !pm.GetHumidityDiscountCalled {
				t.Error("Expected GetHumidityDiscount to be called")
			}
			if em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax NOT to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with damage exceeding limit by 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			dam := decimal.NewFromInt(10)
			entry.Damage = &dam
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(24.500),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax NOT to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with impurity exceeding limit by 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			imp := decimal.NewFromInt(3)
			entry.Impurity = &imp
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(24.500),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}
			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax NOT to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with damage > 8% threshold",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			damage := decimal.NewFromInt(10)
			entry.Damage = &damage
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(882.0),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(456, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with corn discount only",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			var origin uint32 = 1
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.625),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			crop := CreateTestCrop(1, 1, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(1, "Corn")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			if !em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + damage exceeding 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			dam := decimal.NewFromInt(10)
			entry.Humidity = &hum
			entry.Damage = &dam
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.925),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + impurity exceeding 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			imp := decimal.NewFromInt(3)
			entry.Humidity = &hum
			entry.Impurity = &imp
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.925),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + soy storage tax",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			var origin uint32 = 1
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.570125),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			crop := CreateTestCrop(1, 2, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(2, "Soy")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + corn storage tax",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			var origin uint32 = 1
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.081625),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			crop := CreateTestCrop(1, 1, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(1, "Corn")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + damage exceeding 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			dam := decimal.NewFromInt(10)
			entry.Humidity = &hum
			entry.Damage = &dam
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.650),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + impurity exceeding 2%",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			imp := decimal.NewFromInt(3)
			entry.Humidity = &hum
			entry.Impurity = &imp
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.650),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + soy storage tax",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			var origin uint32 = 1
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(23.304750),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			crop := CreateTestCrop(1, 2, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(2, "Soy")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + corn storage tax",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			entry.GrossWeight = decimal.NewFromFloat(50.000)
			entry.Tare = decimal.NewFromFloat(25.000)
			hum := decimal.NewFromInt(16)
			entry.Humidity = &hum
			var origin uint32 = 1
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(22.821750),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			crop := CreateTestCrop(1, 1, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(1, "Corn")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), nil
			}

			discountModifier := decimal.NewFromFloat(1.7)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with impurity > 1% threshold",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			impurity := decimal.NewFromInt(2)
			entry.Impurity = &impurity
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(891.0),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(789, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !em.AddEntryCalled {
				t.Error("Expected AddEntry to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with humidity > 14% threshold",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			humidity := decimal.NewFromInt(16)
			entry.Humidity = &humidity
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(879.3),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(321, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !pm.GetHumidityDiscountCalled {
				t.Error("Expected GetHumidityDiscount to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - Entry with origin and soy storage tax",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			origin := uint32(123)
			entry.Origin = &origin
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(868.5),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			crop := CreateTestCrop(1, 2, 1)
			cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
				return crop, nil
			}

			product := CreateTestProduct(2, "Soy")
			prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
				return product, nil
			}

			personConfig := CreateTestPersonConfig()
			pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, error) {
				return personConfig, nil
			}

			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(555, ge.NetWeight), nil
			}

			em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
				return nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !cm.GetCropByIdCalled {
				t.Error("Expected GetCropById to be called")
			}
			if !prodM.GetProductByIdCalled {
				t.Error("Expected GetProductById to be called")
			}
			if !pm.GetPersonConfigCalled {
				t.Error("Expected GetPersonConfig to be called")
			}
			if !em.AddEntryTaxCalled {
				t.Error("Expected AddEntryTax to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
	{
		name: "Success - All quality discounts combined",
		entry: func() entity_public.Entry {
			entry := CreateBasicTestEntry()
			damage := decimal.NewFromInt(10)
			impurity := decimal.NewFromInt(2)
			humidity := decimal.NewFromInt(16)
			entry.Damage = &damage
			entry.Impurity = &impurity
			entry.Humidity = &humidity
			return entry
		}(),
		expectedNetWeight: decimal.NewFromFloat(852.3),
		setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel, fcm *MockFarmConfigModel) {
			discountModifier := decimal.NewFromFloat(1.15)
			pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
				return discountModifier, nil
			}
			em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
				return CreateTestDisplayEntryWithNetWeight(999, ge.NetWeight), nil
			}
		},
		expectedToastType:    entity_public.SuccessToast,
		expectedToastMessage: "Entrada adicionada",
		validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel, hpm *MockHumidityProgressionModel) {
			if !pm.GetHumidityDiscountCalled {
				t.Error("Expected GetHumidityDiscount to be called")
			}
			AssertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
		},
	},
}

func TestAddEntry(t *testing.T) {
	testCases := append(successCases, errorCases...)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEM := &MockEntryModel{}
			mockPM := &MockPersonModel{}
			mockProdM := &MockProductModel{}
			mockCM := &MockCropModel{}
			mockHPM := &MockHumidityProgressionModel{
				GetFirstTierThresholdFunc: func(progressionId *uint32) (decimal.Decimal, error) {
					return decimal.NewFromInt(14), nil // Default threshold
				},
			}
			mockFCM := &MockFarmConfigModel{
				GetFarmConfigFunc: func(farm uint32) (*entity_public.Farm, error) {
					return &entity_public.Farm{}, nil
				},
			}

			tc.setupMocks(mockEM, mockPM, mockProdM, mockCM, mockHPM, mockFCM)

			result, toast := entry_service.AddEntry(tc.entry, mockEM, mockPM, mockProdM, mockCM, mockHPM, mockFCM)

			if toast.Type != tc.expectedToastType {
				t.Errorf("Expected toast type %v, got %v", tc.expectedToastType, toast.Type)
			}

			if toast.Message != tc.expectedToastMessage {
				t.Errorf("Expected toast message '%s', got '%s'", tc.expectedToastMessage, toast.Message)
			}

			if tc.validateResult != nil {
				tc.validateResult(t, tc, result, mockEM, mockPM, mockProdM, mockCM, mockHPM)
			}
		})
	}
}
