package entry_service

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	entry_testdata "armazenda/service/entry_service/testdata"
	"testing"

	"github.com/shopspring/decimal"
)

type addEntryTestCase struct {
	name                 string
	entry                entity_public.Entry
	setupMocks           func(*MockEntryModel, *MockPersonModel, *MockProductModel, *MockCropModel)
	expectedToastType    entity_public.ToastType
	expectedToastMessage string
	expectedNetWeight    decimal.Decimal
	validateResult       func(*testing.T, addEntryTestCase, entity_public.DisplayEntry, *MockEntryModel, *MockPersonModel, *MockProductModel, *MockCropModel)
}

func assertNetWeightEquals(t *testing.T, expected decimal.Decimal, actual decimal.Decimal) {
	if !expected.Equal(actual) {
		t.Errorf("Expected NetWeight %s, got %s", expected.String(), actual.String())
	}
}

func TestAddEntry(t *testing.T) {
	testCases := []addEntryTestCase{
		{
			name:              "Success - Basic entry (no origin, no quality discounts)",
			entry:             entry_testdata.CreateBasicTestEntry(),
			expectedNetWeight: decimal.NewFromFloat(900.0),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(123, ge.NetWeight), (*model_error.ModelError)(nil)
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if !em.AddEntryCalled {
					t.Error("Expected AddEntry to be called")
				}
				if pm.GetHumidityDiscountCalled {
					t.Error("Expected GetHumidityDiscount NOT to be called")
				}
				if em.AddEntryTaxCalled {
					t.Error("Expected AddEntryTax NOT to be called")
				}
				if result.Id != 123 {
					t.Errorf("Expected ID 123, got %d", result.Id)
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
		{
			name: "Success - Entry with damage > 8% threshold",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				damage := decimal.NewFromInt(10)
				entry.Damage = &damage
				return entry
			}(),
			expectedNetWeight: decimal.NewFromFloat(882.0),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(456, ge.NetWeight), (*model_error.ModelError)(nil)
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if !em.AddEntryCalled {
					t.Error("Expected AddEntry to be called")
				}
				if result.Id != 456 {
					t.Errorf("Expected ID 456, got %d", result.Id)
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
		{
			name: "Success - Entry with impurity > 1% threshold",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				impurity := decimal.NewFromInt(2)
				entry.Impurity = &impurity
				return entry
			}(),
			expectedNetWeight: decimal.NewFromFloat(891.0),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(789, ge.NetWeight), (*model_error.ModelError)(nil)
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if !em.AddEntryCalled {
					t.Error("Expected AddEntry to be called")
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
		{
			name: "Success - Entry with humidity > 14% threshold",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				humidity := decimal.NewFromInt(16)
				entry.Humidity = &humidity
				return entry
			}(),
			expectedNetWeight: decimal.NewFromFloat(879.3),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				discountModifier := decimal.NewFromFloat(1.15)
				pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32) (decimal.Decimal, *model_error.ModelError) {
					return discountModifier, (*model_error.ModelError)(nil)
				}
				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(321, ge.NetWeight), (*model_error.ModelError)(nil)
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if !pm.GetHumidityDiscountCalled {
					t.Error("Expected GetHumidityDiscount to be called")
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
		{
			name: "Success - Entry with origin and storage tax",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				origin := uint32(123)
				entry.Origin = &origin
				return entry
			}(),
			expectedNetWeight: decimal.NewFromFloat(868.5),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				crop := entry_testdata.CreateTestCrop(1, 2, 1)
				cm.GetCropByIdFunc = func(id uint8) (entity_public.Crop, error) {
					return crop, nil
				}

				product := entry_testdata.CreateTestProduct(2, "Soy")
				prodM.GetProductByIdFunc = func(id uint8) (entity_public.Product, error) {
					return product, nil
				}

				personConfig := entry_testdata.CreateTestPersonConfig()
				pm.GetPersonConfigFunc = func(person uint32) (entity_public.PersonConfig, *model_error.ModelError) {
					return personConfig, (*model_error.ModelError)(nil)
				}

				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(555, ge.NetWeight), (*model_error.ModelError)(nil)
				}

				em.AddEntryTaxFunc = func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
					return nil
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
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
				if em.AddEntryTaxArgs.Id != result.Id {
					t.Errorf("Expected tax ID %d, got %d", result.Id, em.AddEntryTaxArgs.Id)
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
		{
			name: "Error - Negative net weight (tare > gross weight)",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				entry.CargoWeight.Tare = decimal.NewFromInt(2000)
				return entry
			}(),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
			},
			expectedToastType:    entity_public.WarningToast,
			expectedToastMessage: "O peso líquido não pode ser menor do que zero",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if em.AddEntryCalled {
					t.Error("Expected AddEntry NOT to be called for negative net weight")
				}
			},
		},
		{
			name: "Error - Humidity discount calculation fails",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				humidity := decimal.NewFromInt(16)
				entry.Humidity = &humidity
				return entry
			}(),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32) (decimal.Decimal, *model_error.ModelError) {
					return decimal.Zero, &model_error.ModelError{Message: "Database error", IsServerErr: true}
				}
			},
			expectedToastType:    entity_public.WarningToast,
			expectedToastMessage: "Falha ao calcular desconto de humidade",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
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
			entry: entry_testdata.CreateBasicTestEntry(),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
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
			entry: entry_testdata.CreateBasicTestEntry(),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
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
		{
			name: "Success - All quality discounts combined",
			entry: func() entity_public.Entry {
				entry := entry_testdata.CreateBasicTestEntry()
				damage := decimal.NewFromInt(10)
				impurity := decimal.NewFromInt(2)
				humidity := decimal.NewFromInt(16)
				entry.Damage = &damage
				entry.Impurity = &impurity
				entry.Humidity = &humidity
				return entry
			}(),
			expectedNetWeight: decimal.NewFromFloat(852.3),
			setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				discountModifier := decimal.NewFromFloat(1.15)
				pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32) (decimal.Decimal, *model_error.ModelError) {
					return discountModifier, (*model_error.ModelError)(nil)
				}
				em.AddEntryFunc = func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
					return entry_testdata.CreateTestDisplayEntryWithNetWeight(999, ge.NetWeight), (*model_error.ModelError)(nil)
				}
			},
			expectedToastType:    entity_public.SuccessToast,
			expectedToastMessage: "Entrada adicionada",
			validateResult: func(t *testing.T, tc addEntryTestCase, result entity_public.DisplayEntry, em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
				if !pm.GetHumidityDiscountCalled {
					t.Error("Expected GetHumidityDiscount to be called")
				}
				assertNetWeightEquals(t, tc.expectedNetWeight, result.NetWeight)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEM := &MockEntryModel{}
			mockPM := &MockPersonModel{}
			mockProdM := &MockProductModel{}
			mockCM := &MockCropModel{}

			tc.setupMocks(mockEM, mockPM, mockProdM, mockCM)

			result, toast := AddEntry(tc.entry, mockEM, mockPM, mockProdM, mockCM)

			if toast.Type != tc.expectedToastType {
				t.Errorf("Expected toast type %v, got %v", tc.expectedToastType, toast.Type)
			}

			if toast.Message != tc.expectedToastMessage {
				t.Errorf("Expected toast message '%s', got '%s'", tc.expectedToastMessage, toast.Message)
			}

			if tc.validateResult != nil {
				tc.validateResult(t, tc, result, mockEM, mockPM, mockProdM, mockCM)
			}
		})
	}
}
