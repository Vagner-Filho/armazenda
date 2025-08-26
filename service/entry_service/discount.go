package entry_service

import "github.com/shopspring/decimal"

var HUMIDITY_THRESHOLD = decimal.NewFromInt(14)
var DAMAGE_THRESHOLD = decimal.NewFromInt(8)
var IMPURITY_THRESHOLD = decimal.NewFromInt(1)
var DECIMAL_ZERO = decimal.NewFromInt(0)
var DECIMAL_HUNDRED = decimal.NewFromInt(100)

func DiscountHumidity(humidity *decimal.Decimal, rawNetWeight decimal.Decimal, discountModifier *decimal.Decimal) decimal.Decimal {
	if humidity == nil {
		return rawNetWeight
	}
	if discountModifier == nil {
		return rawNetWeight
	}

	var exceedingHumidity = humidity.Sub(HUMIDITY_THRESHOLD)
	if exceedingHumidity.LessThanOrEqual(DECIMAL_ZERO) {
		return rawNetWeight
	}

	var discount = exceedingHumidity.Mul(*discountModifier)
	return rawNetWeight.Mul(discount).Div(DECIMAL_HUNDRED)
}

func DiscountImpurity(impurity *decimal.Decimal, rawNetWeight decimal.Decimal) decimal.Decimal {
	if impurity == nil {
		return rawNetWeight
	}

	var exceedingImpurity = impurity.Sub(IMPURITY_THRESHOLD)
	if exceedingImpurity.LessThanOrEqual(DECIMAL_ZERO) {
		return rawNetWeight
	}

	return rawNetWeight.Mul(exceedingImpurity).Div(DECIMAL_HUNDRED)
}

func DiscountDamage(damage *decimal.Decimal, rawNetWeight decimal.Decimal) decimal.Decimal {
	if damage == nil {
		return rawNetWeight
	}

	var exceedingDamage = damage.Sub(IMPURITY_THRESHOLD)
	if exceedingDamage.LessThanOrEqual(DECIMAL_ZERO) {
		return rawNetWeight
	}

	return rawNetWeight.Mul(exceedingDamage).Div(DECIMAL_HUNDRED)
}
