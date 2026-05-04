package entry_service

import "github.com/shopspring/decimal"

var DEFAULT_HUMIDITY_THRESHOLD = decimal.NewFromInt(14)
var DAMAGE_THRESHOLD = decimal.NewFromInt(8)
var IMPURITY_THRESHOLD = decimal.NewFromInt(1)
var DECIMAL_ZERO = decimal.NewFromInt(0)
var DECIMAL_HUNDRED = decimal.NewFromInt(100)

func DiscountHumidity(humidity *decimal.Decimal, rawNetWeight decimal.Decimal, discountModifier *decimal.Decimal, humidityThreshold *decimal.Decimal) decimal.Decimal {
	if humidity == nil {
		return DECIMAL_ZERO
	}
	if discountModifier == nil {
		return DECIMAL_ZERO
	}

	// Use provided threshold or default
	threshold := DEFAULT_HUMIDITY_THRESHOLD
	if humidityThreshold != nil {
		threshold = *humidityThreshold
	}

	var exceedingHumidity = humidity.Sub(threshold)
	if exceedingHumidity.LessThanOrEqual(DECIMAL_ZERO) {
		return DECIMAL_ZERO
	}

	var discount = exceedingHumidity.Mul(*discountModifier)
	return rawNetWeight.Mul(discount).Div(DECIMAL_HUNDRED)
}

func DiscountImpurity(impurity *decimal.Decimal, rawNetWeight decimal.Decimal) decimal.Decimal {
	if impurity == nil {
		return DECIMAL_ZERO
	}

	var exceedingImpurity = impurity.Sub(IMPURITY_THRESHOLD)
	if exceedingImpurity.LessThanOrEqual(DECIMAL_ZERO) {
		return DECIMAL_ZERO
	}

	return rawNetWeight.Mul(exceedingImpurity).Div(DECIMAL_HUNDRED)
}

func DiscountDamage(damage *decimal.Decimal, rawNetWeight decimal.Decimal) decimal.Decimal {
	if damage == nil {
		return DECIMAL_ZERO
	}

	var exceedingDamage = damage.Sub(DAMAGE_THRESHOLD)
	if exceedingDamage.LessThanOrEqual(DECIMAL_ZERO) {
		return DECIMAL_ZERO
	}

	return rawNetWeight.Mul(exceedingDamage).Div(DECIMAL_HUNDRED)
}
