package entry_service_test

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"

	"github.com/shopspring/decimal"
)

// MockEntryModel mocks EntryModelInterface for testing
type MockEntryModel struct {
	AddEntryFunc      func(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError)
	AddEntryTaxFunc   func(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error
	AddEntryCalled    bool
	AddEntryTaxCalled bool
	AddEntryTaxArgs   struct {
		Id         uint32
		Tax        decimal.Decimal
		AppliedTax decimal.Decimal
	}
}

func (m *MockEntryModel) AddEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
	m.AddEntryCalled = true
	return m.AddEntryFunc(ge)
}

func (m *MockEntryModel) AddEntryTax(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
	m.AddEntryTaxCalled = true
	m.AddEntryTaxArgs.Id = id
	m.AddEntryTaxArgs.Tax = tax
	m.AddEntryTaxArgs.AppliedTax = appliedTax
	return m.AddEntryTaxFunc(id, tax, appliedTax)
}

// MockPersonModel mocks PersonModelInterface for testing
type MockPersonModel struct {
	GetHumidityDiscountFunc   func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error)
	GetPersonConfigFunc       func(person uint32) (entity_public.PersonConfig, error)
	GetHumidityDiscountCalled bool
	GetPersonConfigCalled     bool
	GetHumidityDiscountArgs   struct {
		Person   *uint32
		Farm     uint32
		Humidity decimal.Decimal
	}
	GetPersonConfigArgs struct {
		Person uint32
	}
}

func (m *MockPersonModel) GetHumidityDiscount(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error) {
	m.GetHumidityDiscountCalled = true
	m.GetHumidityDiscountArgs.Person = person
	m.GetHumidityDiscountArgs.Farm = farm
	m.GetHumidityDiscountArgs.Humidity = humidity
	return m.GetHumidityDiscountFunc(person, farm, humidity)
}

func (m *MockPersonModel) GetPersonConfig(person uint32) (entity_public.PersonConfig, error) {
	m.GetPersonConfigCalled = true
	m.GetPersonConfigArgs.Person = person
	return m.GetPersonConfigFunc(person)
}

// MockProductModel mocks ProductModelInterface for testing
type MockProductModel struct {
	GetProductByIdFunc   func(id uint8) (entity_public.Product, error)
	GetProductByIdCalled bool
	GetProductByIdArgs   uint8
}

func (m *MockProductModel) GetProductById(id uint8) (entity_public.Product, error) {
	m.GetProductByIdCalled = true
	m.GetProductByIdArgs = id
	return m.GetProductByIdFunc(id)
}

// MockCropModel mocks CropModelInterface for testing
type MockCropModel struct {
	GetCropByIdFunc   func(id uint8) (entity_public.Crop, error)
	GetCropByIdCalled bool
	GetCropByIdArgs   uint8
}

func (m *MockCropModel) GetCropById(id uint8) (entity_public.Crop, error) {
	m.GetCropByIdCalled = true
	m.GetCropByIdArgs = id
	return m.GetCropByIdFunc(id)
}

// MockHumidityProgressionModel mocks HumidityProgressionModelInterface for testing
type MockHumidityProgressionModel struct {
	GetFirstTierThresholdFunc   func(progressionId *uint32) (decimal.Decimal, error)
	GetFirstTierThresholdCalled bool
	GetFirstTierThresholdArgs   *uint32
}

func (m *MockHumidityProgressionModel) GetFirstTierThreshold(progressionId *uint32) (decimal.Decimal, error) {
	m.GetFirstTierThresholdCalled = true
	m.GetFirstTierThresholdArgs = progressionId
	return m.GetFirstTierThresholdFunc(progressionId)
}

type MockFarmConfigModel struct {
	GetFarmConfigFunc   func(farm uint32) (*entity_public.Farm, error)
	GetFarmConfigCalled bool
	GetFarmConfigArgs   uint32
}

func (m *MockFarmConfigModel) GetFarmConfig(farm uint32) (*entity_public.Farm, error) {
	m.GetFarmConfigCalled = true
	m.GetFarmConfigArgs = farm
	return m.GetFarmConfigFunc(farm)
}
