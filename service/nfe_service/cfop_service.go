package nfe_service

import (
	"fmt"
	"strings"

	model_error "armazenda/model/error"
)

// Origem/Destino values of the CFOP catalog (third column of the reference
// table). They are derived from the first digit of the code.
const (
	CfopOriginSameState  = "Mesmo estado"
	CfopOriginOtherState = "Outro estado"
	CfopOriginForeign    = "Exterior"
)

// ValidSelectorCFOP reports whether code is a well-formed CFOP for the
// detached selector: exactly four ASCII digits whose first digit is a valid
// family (1-3 entries, 5-7 exits). The item parsers keep using the looser
// IsFourDigitCFOP so the form contract is unchanged; this stricter helper
// guards the register-modal route.
func ValidSelectorCFOP(code string) bool {
	if !IsFourDigitCFOP(code) {
		return false
	}

	switch code[0] {
	case '1', '2', '3', '5', '6', '7':
		return true
	default:
		return false
	}
}

// OriginDestinationForCFOP derives the Origem/Destino value from the first
// digit of a CFOP code: 1/5 same state, 2/6 other state, 3/7 foreign.
// It returns an empty string for codes outside the valid families.
func OriginDestinationForCFOP(code string) string {
	if code == "" {
		return ""
	}

	switch code[0] {
	case '1', '5':
		return CfopOriginSameState
	case '2', '6':
		return CfopOriginOtherState
	case '3', '7':
		return CfopOriginForeign
	default:
		return ""
	}
}

// distinctCFOPs returns the unique, non-empty codes preserving first-seen
// order. Counting and grouping are always per distinct code.
func distinctCFOPs(codes []string) []string {
	seen := make(map[string]struct{}, len(codes))
	distinct := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		distinct = append(distinct, code)
	}
	return distinct
}

// detachedItemCFOPs extracts the per-item codes of a detached emission.
func detachedItemCFOPs(items []DetachedItemInput) []string {
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.CFOP)
	}
	return codes
}

// farmCfopUseIncrementer is the slice of the cfop model consumed by the
// emission increment hook. It exists so tests can inject a mock.
type farmCfopUseIncrementer interface {
	IncrementFarmCfopUse(farmID uint32, codes []string) *model_error.ModelError
}

// countDistinctFarmCfopUse bumps the farm's use counters once per distinct
// code of one emission action. It is a local ranking update: failures are
// logged and never block or toast the emission, and the worker retry paths
// never call it (counting is bound to the persist step of BuildDetachedInvoice
// only, including its single SVC supersede branch).
func countDistinctFarmCfopUse(incrementer farmCfopUseIncrementer, farmID uint32, codes []string) {
	distinct := distinctCFOPs(codes)
	if len(distinct) == 0 {
		return
	}

	if err := incrementer.IncrementFarmCfopUse(farmID, distinct); err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("countDistinctFarmCfopUse error: %s", err.Error()))
	}
}
