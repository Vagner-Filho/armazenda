//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"armazenda/pkg/calculator"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("WASM Calculator loaded")

	// Export functions to JavaScript
	js.Global().Set("calculateEntry", js.FuncOf(calculateEntry))
	js.Global().Set("calculateDeparture", js.FuncOf(calculateDeparture))
	js.Global().Set("calculateDiscounts", js.FuncOf(calculateDiscounts))
	js.Global().Set("calculateStorageTax", js.FuncOf(calculateStorageTax))

	// Keep the program running
	select {}
}

// calculateEntry calculates net weight and all discounts for an entry
// Input: JSON string of EntryCalculationInput
// Output: JSON string of EntryCalculationResult
func calculateEntry(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResponse("missing input")
	}

	inputJSON := args[0].String()
	var input calculator.EntryCalculationInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return createErrorResponse("invalid input: " + err.Error())
	}

	result := calculator.CalculateEntry(input)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return createErrorResponse("failed to marshal result: " + err.Error())
	}

	return string(resultJSON)
}

// calculateDeparture calculates net weight for a departure
// Input: JSON string of DepartureCalculationInput
// Output: JSON string of DepartureCalculationResult
func calculateDeparture(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResponse("missing input")
	}

	inputJSON := args[0].String()
	var input calculator.DepartureCalculationInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return createErrorResponse("invalid input: " + err.Error())
	}

	result := calculator.CalculateDeparture(input)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return createErrorResponse("failed to marshal result: " + err.Error())
	}

	return string(resultJSON)
}

// calculateDiscounts calculates individual discount components
// Input: JSON string of DiscountCalculationInput
// Output: JSON string of DiscountResult
func calculateDiscounts(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResponse("missing input")
	}

	inputJSON := args[0].String()
	var input DiscountCalculationInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return createErrorResponse("invalid input: " + err.Error())
	}

	result := calculator.CalculateDiscounts(
		input.Humidity,
		input.Damage,
		input.Impurity,
		input.GrossWeight,
		input.Tare,
		input.HumidityModifier,
		input.HumidityThreshold,
	)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return createErrorResponse("failed to marshal result: " + err.Error())
	}

	return string(resultJSON)
}

// calculateStorageTax calculates storage tax
// Input: JSON string of StorageTaxInput
// Output: JSON string of StorageTaxResult
func calculateStorageTax(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResponse("missing input")
	}

	inputJSON := args[0].String()
	var input StorageTaxInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return createErrorResponse("invalid input: " + err.Error())
	}

	tax := calculator.CalculateStorageTax(input.NetWeight, input.StorageTaxModifier)

	result := StorageTaxResult{
		StorageTax: tax,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return createErrorResponse("failed to marshal result: " + err.Error())
	}

	return string(resultJSON)
}

func createErrorResponse(message string) string {
	err := map[string]interface{}{
		"error": message,
	}
	errJSON, _ := json.Marshal(err)
	return string(errJSON)
}

// Input types for WASM communication

type DiscountCalculationInput struct {
	Humidity          *decimal.Decimal `json:"humidity,omitempty"`
	Damage            *decimal.Decimal `json:"damage,omitempty"`
	Impurity          *decimal.Decimal `json:"impurity,omitempty"`
	GrossWeight       decimal.Decimal  `json:"grossWeight"`
	Tare              decimal.Decimal  `json:"tare"`
	HumidityModifier  *decimal.Decimal `json:"humidityModifier,omitempty"`
	HumidityThreshold *decimal.Decimal `json:"humidityThreshold,omitempty"`
}

type StorageTaxInput struct {
	NetWeight          decimal.Decimal `json:"netWeight"`
	StorageTaxModifier decimal.Decimal `json:"storageTaxModifier"`
}

type StorageTaxResult struct {
	StorageTax decimal.Decimal `json:"storageTax"`
}
