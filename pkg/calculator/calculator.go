package calculator

import (
	"github.com/shopspring/decimal"
)

// CalculateEntry performs all calculations for an entry
// Pure function - no DB dependencies. Callers must fetch all required data beforehand.
func CalculateEntry(input EntryCalculationInput) EntryCalculationResult {
	result := EntryCalculationResult{}

	// Calculate raw net weight
	rawNetWeight := input.GrossWeight.Sub(input.Tare)
	result.RawNetWeight = rawNetWeight

	// Validate net weight
	if rawNetWeight.LessThan(decimal.Zero) {
		result.IsValid = false
		result.ErrorMessage = "O peso líquido não pode ser menor do que zero"
		return result
	}

	// Calculate quality discounts
	discounts := CalculateDiscounts(
		input.Humidity,
		input.Damage,
		input.Impurity,
		input.GrossWeight,
		input.Tare,
		input.HumidityModifier,
		input.HumidityThreshold,
	)

	result.DiscountedHumidity = discounts.HumidityDiscount
	result.DiscountedDamage = discounts.DamageDiscount
	result.DiscountedImpurity = discounts.ImpurityDiscount

	if input.HumidityModifier != nil {
		result.HumidityDiscountModifier = *input.HumidityModifier
	}

	totalDiscount := discounts.TotalDiscount

	// Calculate net weight after quality discounts
	netWeightAfterQuality := rawNetWeight.Sub(totalDiscount)

	// Calculate storage tax if modifier is provided (indicates Origin is set)
	if input.StorageTaxModifier != nil {
		result.StorageTaxModifier = *input.StorageTaxModifier
		storageTax := netWeightAfterQuality.Mul(*input.StorageTaxModifier).Div(Base100)
		result.StorageTax = storageTax
		result.NetWeight = netWeightAfterQuality.Sub(storageTax)
	} else {
		result.NetWeight = netWeightAfterQuality
	}

	result.IsValid = true
	return result
}

// CalculateDiscounts calculates all quality-based discounts
func CalculateDiscounts(
	humidity *decimal.Decimal,
	damage *decimal.Decimal,
	impurity *decimal.Decimal,
	grossWeight decimal.Decimal,
	tare decimal.Decimal,
	humidityModifier *decimal.Decimal,
	humidityThreshold *decimal.Decimal,
) DiscountResult {
	result := DiscountResult{}

	rawNetWeight := grossWeight.Sub(tare)

	// Use provided threshold or default
	threshold := DefaultHumidityThreshold
	if humidityThreshold != nil {
		threshold = *humidityThreshold
	}

	// Calculate humidity discount
	if humidity != nil {
		exceedingHumidity := humidity.Sub(threshold)
		if exceedingHumidity.GreaterThan(decimal.Zero) {
			var discount decimal.Decimal
			if humidityModifier != nil {
				discount = exceedingHumidity.Mul(*humidityModifier)
			} else {
				discount = exceedingHumidity // Default modifier of 1
			}
			result.HumidityDiscount = rawNetWeight.Mul(discount).Div(Base100)
		}
	}

	// Calculate damage discount
	if damage != nil {
		exceedingDamage := damage.Sub(DamageThreshold)
		if exceedingDamage.GreaterThan(decimal.Zero) {
			result.DamageDiscount = rawNetWeight.Mul(exceedingDamage).Div(Base100)
		}
	}

	// Calculate impurity discount
	if impurity != nil {
		exceedingImpurity := impurity.Sub(ImpurityThreshold)
		if exceedingImpurity.GreaterThan(decimal.Zero) {
			result.ImpurityDiscount = rawNetWeight.Mul(exceedingImpurity).Div(Base100)
		}
	}

	// Calculate total
	result.TotalDiscount = result.HumidityDiscount.Add(result.DamageDiscount).Add(result.ImpurityDiscount)

	return result
}

// CalculateDeparture performs calculations for a departure
func CalculateDeparture(input DepartureCalculationInput) DepartureCalculationResult {
	result := DepartureCalculationResult{}

	// Calculate raw net weight
	rawNetWeight := input.GrossWeight.Sub(input.Tare)
	result.RawNetWeight = rawNetWeight

	// Validate net weight
	if rawNetWeight.LessThan(decimal.Zero) {
		result.IsValid = false
		result.ErrorMessage = "O peso líquido não pode ser menor do que zero"
		return result
	}

	// Calculate quality discounts (same logic as entry)
	discounts := CalculateDiscounts(
		nil,
		nil,
		nil,
		input.GrossWeight,
		input.Tare,
		nil, // No humidity modifier for departures
		nil, // No custom threshold for departures (uses default)
	)

	totalDiscount := discounts.TotalDiscount

	// Calculate final net weight
	result.NetWeight = rawNetWeight.Sub(totalDiscount)
	result.IsValid = true

	return result
}

// CalculateStorageTax calculates storage tax given net weight and modifier
func CalculateStorageTax(netWeight decimal.Decimal, storageTaxModifier decimal.Decimal) decimal.Decimal {
	return netWeight.Mul(storageTaxModifier).Div(Base100)
}
