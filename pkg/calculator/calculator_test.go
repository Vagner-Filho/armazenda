package calculator_test

import (
	"armazenda/pkg/calculator"
	"testing"

	"github.com/shopspring/decimal"
)

// Test case structure for calculator tests
type calculatorTestCase struct {
	name              string
	input             calculator.EntryCalculationInput
	expectedNetWeight decimal.Decimal
	expectedValid     bool
	expectedErrorMsg  string
}

// Helper function to create decimal pointer
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// Error Cases
var errorCases []calculatorTestCase = []calculatorTestCase{
	{
		name: "Error - Negative net weight (tare > gross weight)",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(2000),
		},
		expectedValid:    false,
		expectedErrorMsg: "O peso líquido não pode ser menor do que zero",
	},
}

// Success Cases - adapted from service_test.go
var successCases []calculatorTestCase = []calculatorTestCase{
	{
		name: "Success - Basic entry (no origin, no quality discounts, no service tax)",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(100),
		},
		expectedNetWeight: decimal.NewFromFloat(900.0),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with exceeding humidity limit by 2% and default discount for farm",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
		},
		expectedNetWeight: decimal.NewFromFloat(24.425),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with exceeding humidity limit by 2% and default discount for person",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.7)),
		},
		expectedNetWeight: decimal.NewFromFloat(24.150),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with damage exceeding limit by 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromFloat(50.000),
			Tare:        decimal.NewFromFloat(25.000),
			Damage:      decimalPtr(decimal.NewFromInt(10)),
		},
		expectedNetWeight: decimal.NewFromFloat(24.500),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with impurity exceeding limit by 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromFloat(50.000),
			Tare:        decimal.NewFromFloat(25.000),
			Impurity:    decimalPtr(decimal.NewFromInt(3)),
		},
		expectedNetWeight: decimal.NewFromFloat(24.500),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with damage > 8% threshold",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(100),
			Damage:      decimalPtr(decimal.NewFromInt(10)),
		},
		expectedNetWeight: decimal.NewFromFloat(882.0),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with corn discount only",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromFloat(50.000),
			Tare:               decimal.NewFromFloat(25.000),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(5.5)), // Corn discount
		},
		expectedNetWeight: decimal.NewFromFloat(23.625),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + damage exceeding 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			Damage:           decimalPtr(decimal.NewFromInt(10)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
		},
		expectedNetWeight: decimal.NewFromFloat(23.925),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + impurity exceeding 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			Impurity:         decimalPtr(decimal.NewFromInt(3)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
		},
		expectedNetWeight: decimal.NewFromFloat(23.925),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + soy storage tax",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromFloat(50.000),
			Tare:               decimal.NewFromFloat(25.000),
			Humidity:           decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier:   decimalPtr(decimal.NewFromFloat(1.15)),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(3.5)), // Soy discount
		},
		expectedNetWeight: decimal.NewFromFloat(23.570125),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for farm + corn storage tax",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromFloat(50.000),
			Tare:               decimal.NewFromFloat(25.000),
			Humidity:           decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier:   decimalPtr(decimal.NewFromFloat(1.15)),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(5.5)), // Corn discount
		},
		expectedNetWeight: decimal.NewFromFloat(23.081625),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + damage exceeding 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			Damage:           decimalPtr(decimal.NewFromInt(10)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.7)),
		},
		expectedNetWeight: decimal.NewFromFloat(23.650),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + impurity exceeding 2%",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromFloat(50.000),
			Tare:             decimal.NewFromFloat(25.000),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			Impurity:         decimalPtr(decimal.NewFromInt(3)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.7)),
		},
		expectedNetWeight: decimal.NewFromFloat(23.650),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + soy storage tax",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromFloat(50.000),
			Tare:               decimal.NewFromFloat(25.000),
			Humidity:           decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier:   decimalPtr(decimal.NewFromFloat(1.7)),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(3.5)), // Soy discount
		},
		expectedNetWeight: decimal.NewFromFloat(23.304750),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity exceending 2% + hum modifier for person + corn storage tax",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromFloat(50.000),
			Tare:               decimal.NewFromFloat(25.000),
			Humidity:           decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier:   decimalPtr(decimal.NewFromFloat(1.7)),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(5.5)), // Corn discount
		},
		expectedNetWeight: decimal.NewFromFloat(22.821750),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with impurity > 1% threshold",
		input: calculator.EntryCalculationInput{
			GrossWeight: decimal.NewFromInt(1000),
			Tare:        decimal.NewFromInt(100),
			Impurity:    decimalPtr(decimal.NewFromInt(2)),
		},
		expectedNetWeight: decimal.NewFromFloat(891.0),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with humidity > 14% threshold",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromInt(1000),
			Tare:             decimal.NewFromInt(100),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
		},
		expectedNetWeight: decimal.NewFromFloat(879.3),
		expectedValid:     true,
	},
	{
		name: "Success - Entry with origin and soy storage tax",
		input: calculator.EntryCalculationInput{
			GrossWeight:        decimal.NewFromInt(1000),
			Tare:               decimal.NewFromInt(100),
			StorageTaxModifier: decimalPtr(decimal.NewFromFloat(3.5)), // Soy discount
		},
		expectedNetWeight: decimal.NewFromFloat(868.5),
		expectedValid:     true,
	},
	{
		name: "Success - All quality discounts combined",
		input: calculator.EntryCalculationInput{
			GrossWeight:      decimal.NewFromInt(1000),
			Tare:             decimal.NewFromInt(100),
			Humidity:         decimalPtr(decimal.NewFromInt(16)),
			Damage:           decimalPtr(decimal.NewFromInt(10)),
			Impurity:         decimalPtr(decimal.NewFromInt(2)),
			HumidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
		},
		expectedNetWeight: decimal.NewFromFloat(852.3),
		expectedValid:     true,
	},
}

func TestCalculateEntry(t *testing.T) {
	testCases := append(successCases, errorCases...)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculator.CalculateEntry(tc.input)

			if result.IsValid != tc.expectedValid {
				t.Errorf("Expected IsValid=%v, got %v", tc.expectedValid, result.IsValid)
			}

			if !tc.expectedValid {
				if result.ErrorMessage != tc.expectedErrorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.expectedErrorMsg, result.ErrorMessage)
				}
				return
			}

			if !result.NetWeight.Equal(tc.expectedNetWeight) {
				t.Errorf("Expected NetWeight %s, got %s", tc.expectedNetWeight.String(), result.NetWeight.String())
			}
		})
	}
}

func TestCalculateDiscounts(t *testing.T) {
	tests := []struct {
		name             string
		humidity         *decimal.Decimal
		damage           *decimal.Decimal
		impurity         *decimal.Decimal
		grossWeight      decimal.Decimal
		tare             decimal.Decimal
		humidityModifier *decimal.Decimal
		expectedHumidity decimal.Decimal
		expectedDamage   decimal.Decimal
		expectedImpurity decimal.Decimal
		expectedTotal    decimal.Decimal
	}{
		{
			name:             "No discounts - all within thresholds",
			humidity:         nil,
			damage:           nil,
			impurity:         nil,
			grossWeight:      decimal.NewFromInt(1000),
			tare:             decimal.NewFromInt(100),
			humidityModifier: nil,
			expectedHumidity: decimal.Zero,
			expectedDamage:   decimal.Zero,
			expectedImpurity: decimal.Zero,
			expectedTotal:    decimal.Zero,
		},
		{
			name:             "Humidity exceeding by 2% with modifier 1.15",
			humidity:         decimalPtr(decimal.NewFromInt(16)),
			damage:           nil,
			impurity:         nil,
			grossWeight:      decimal.NewFromFloat(50),
			tare:             decimal.NewFromFloat(25),
			humidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
			expectedHumidity: decimal.NewFromFloat(0.575),
			expectedDamage:   decimal.Zero,
			expectedImpurity: decimal.Zero,
			expectedTotal:    decimal.NewFromFloat(0.575),
		},
		{
			name:             "Damage exceeding by 2%",
			humidity:         nil,
			damage:           decimalPtr(decimal.NewFromInt(10)),
			impurity:         nil,
			grossWeight:      decimal.NewFromFloat(50),
			tare:             decimal.NewFromFloat(25),
			humidityModifier: nil,
			expectedHumidity: decimal.Zero,
			expectedDamage:   decimal.NewFromFloat(0.5),
			expectedImpurity: decimal.Zero,
			expectedTotal:    decimal.NewFromFloat(0.5),
		},
		{
			name:             "Impurity exceeding by 2%",
			humidity:         nil,
			damage:           nil,
			impurity:         decimalPtr(decimal.NewFromInt(3)),
			grossWeight:      decimal.NewFromFloat(50),
			tare:             decimal.NewFromFloat(25),
			humidityModifier: nil,
			expectedHumidity: decimal.Zero,
			expectedDamage:   decimal.Zero,
			expectedImpurity: decimal.NewFromFloat(0.5),
			expectedTotal:    decimal.NewFromFloat(0.5),
		},
		{
			name:             "All discounts combined",
			humidity:         decimalPtr(decimal.NewFromInt(16)),
			damage:           decimalPtr(decimal.NewFromInt(10)),
			impurity:         decimalPtr(decimal.NewFromInt(2)),
			grossWeight:      decimal.NewFromInt(1000),
			tare:             decimal.NewFromInt(100),
			humidityModifier: decimalPtr(decimal.NewFromFloat(1.15)),
			expectedHumidity: decimal.NewFromFloat(20.7),
			expectedDamage:   decimal.NewFromInt(18),
			expectedImpurity: decimal.NewFromInt(9),
			expectedTotal:    decimal.NewFromFloat(47.7),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.CalculateDiscounts(tt.humidity, tt.damage, tt.impurity, tt.grossWeight, tt.tare, tt.humidityModifier)

			if !result.HumidityDiscount.Equal(tt.expectedHumidity) {
				t.Errorf("HumidityDiscount = %v, want %v", result.HumidityDiscount, tt.expectedHumidity)
			}
			if !result.DamageDiscount.Equal(tt.expectedDamage) {
				t.Errorf("DamageDiscount = %v, want %v", result.DamageDiscount, tt.expectedDamage)
			}
			if !result.ImpurityDiscount.Equal(tt.expectedImpurity) {
				t.Errorf("ImpurityDiscount = %v, want %v", result.ImpurityDiscount, tt.expectedImpurity)
			}
			if !result.TotalDiscount.Equal(tt.expectedTotal) {
				t.Errorf("TotalDiscount = %v, want %v", result.TotalDiscount, tt.expectedTotal)
			}
		})
	}
}

func TestCalculateDeparture(t *testing.T) {
	tests := []struct {
		name              string
		input             calculator.DepartureCalculationInput
		expectedNetWeight decimal.Decimal
		expectedValid     bool
		expectedErrorMsg  string
	}{
		{
			name: "Valid departure - no discounts",
			input: calculator.DepartureCalculationInput{
				GrossWeight: decimal.NewFromInt(1000),
				Tare:        decimal.NewFromInt(100),
			},
			expectedNetWeight: decimal.NewFromInt(900),
			expectedValid:     true,
		},
		{
			name: "Departure with damage exceeding threshold",
			input: calculator.DepartureCalculationInput{
				GrossWeight: decimal.NewFromFloat(50),
				Tare:        decimal.NewFromFloat(25),
				Damage:      decimalPtr(decimal.NewFromInt(10)),
			},
			expectedNetWeight: decimal.NewFromFloat(24.5),
			expectedValid:     true,
		},
		{
			name: "Invalid - negative net weight",
			input: calculator.DepartureCalculationInput{
				GrossWeight: decimal.NewFromInt(100),
				Tare:        decimal.NewFromInt(200),
			},
			expectedValid:    false,
			expectedErrorMsg: "O peso líquido não pode ser menor do que zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.CalculateDeparture(tt.input)

			if result.IsValid != tt.expectedValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.expectedValid)
			}

			if !tt.expectedValid {
				if result.ErrorMessage != tt.expectedErrorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedErrorMsg, result.ErrorMessage)
				}
				return
			}

			if !result.NetWeight.Equal(tt.expectedNetWeight) {
				t.Errorf("NetWeight = %v, want %v", result.NetWeight, tt.expectedNetWeight)
			}
		})
	}
}

func TestIndividualDiscountFunctions(t *testing.T) {
	t.Run("DiscountHumidity", func(t *testing.T) {
		humidity := decimal.NewFromInt(16)
		rawNetWeight := decimal.NewFromInt(900)
		modifier := decimal.NewFromFloat(1.15)

		result := calculator.DiscountHumidity(&humidity, rawNetWeight, &modifier)
		expected := decimal.NewFromFloat(20.7)

		if !result.Equal(expected) {
			t.Errorf("DiscountHumidity = %v, want %v", result, expected)
		}
	})

	t.Run("DiscountDamage", func(t *testing.T) {
		damage := decimal.NewFromInt(10)
		rawNetWeight := decimal.NewFromInt(900)

		result := calculator.DiscountDamage(&damage, rawNetWeight)
		expected := decimal.NewFromInt(18)

		if !result.Equal(expected) {
			t.Errorf("DiscountDamage = %v, want %v", result, expected)
		}
	})

	t.Run("DiscountImpurity", func(t *testing.T) {
		impurity := decimal.NewFromInt(2)
		rawNetWeight := decimal.NewFromInt(900)

		result := calculator.DiscountImpurity(&impurity, rawNetWeight)
		expected := decimal.NewFromInt(9)

		if !result.Equal(expected) {
			t.Errorf("DiscountImpurity = %v, want %v", result, expected)
		}
	})
}
