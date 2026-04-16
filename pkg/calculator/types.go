package calculator

import (
	entity_public "armazenda/entity/public"
	"github.com/shopspring/decimal"
)

// Thresholds for quality discounts
var (
	DefaultHumidityThreshold = decimal.NewFromInt(14) // Default when no progression specified
	DamageThreshold          = decimal.NewFromInt(8)
	ImpurityThreshold        = decimal.NewFromInt(1)
	Base100                  = decimal.NewFromInt(100)
	DecimalZero              = decimal.NewFromInt(0)
)

// EntryCalculationInput holds all data needed for entry calculations
type EntryCalculationInput struct {
	GrossWeight        decimal.Decimal
	Tare               decimal.Decimal
	Humidity           *decimal.Decimal
	Damage             *decimal.Decimal
	Impurity           *decimal.Decimal
	HumidityModifier   *decimal.Decimal // Caller must fetch this from DB
	StorageTaxModifier *decimal.Decimal // Caller must fetch from PersonConfig if Origin is set
	HumidityThreshold  *decimal.Decimal // First tier threshold from progression (defaults to 14 if nil)
}

// EntryCalculationResult holds the result of entry calculations
type EntryCalculationResult struct {
	RawNetWeight             decimal.Decimal
	NetWeight                decimal.Decimal
	DiscountedHumidity       decimal.Decimal
	DiscountedDamage         decimal.Decimal
	DiscountedImpurity       decimal.Decimal
	StorageTax               decimal.Decimal
	StorageTaxModifier       decimal.Decimal
	HumidityDiscountModifier decimal.Decimal
	IsValid                  bool
	ErrorMessage             string
}

// DiscountResult holds individual discount calculations
type DiscountResult struct {
	HumidityDiscount decimal.Decimal
	DamageDiscount   decimal.Decimal
	ImpurityDiscount decimal.Decimal
	TotalDiscount    decimal.Decimal
}

// DepartureCalculationInput holds data needed for departure calculations
type DepartureCalculationInput struct {
	GrossWeight decimal.Decimal
	Tare        decimal.Decimal
	Humidity    *decimal.Decimal
	Damage      *decimal.Decimal
	Impurity    *decimal.Decimal
}

// DepartureCalculationResult holds the result of departure calculations
type DepartureCalculationResult struct {
	NetWeight    decimal.Decimal
	RawNetWeight decimal.Decimal
	IsValid      bool
	ErrorMessage string
}

// Helper function to create test display entry with specific net weight
func CreateTestDisplayEntryWithNetWeight(id uint32, netWeight decimal.Decimal) entity_public.DisplayEntry {
	return entity_public.DisplayEntry{
		Id:        id,
		NetWeight: netWeight,
	}
}
