package entry_service

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"github.com/shopspring/decimal"
)

// EntryModelInterface defines entry-related operations
type EntryModelInterface interface {
	AddEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError)
	AddEntryTax(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error
}

// PersonModelInterface defines person-related operations
type PersonModelInterface interface {
	GetHumidityDiscount(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, error)
	GetPersonConfig(person uint32) (entity_public.PersonConfig, error)
}

// ProductModelInterface defines product-related operations
type ProductModelInterface interface {
	GetProductById(id uint8) (entity_public.Product, error)
}

// CropModelInterface defines crop-related operations
type CropModelInterface interface {
	GetCropById(id uint8) (entity_public.Crop, error)
}

// HumidityProgressionModelInterface defines humidity progression operations
type HumidityProgressionModelInterface interface {
	GetFirstTierThreshold(progressionId *uint32) (decimal.Decimal, error)
}

type FarmConfigModelInterface interface {
	GetFarmConfig(farm uint32) (*entity_public.Farm, error)
}
