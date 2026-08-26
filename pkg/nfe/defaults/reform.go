package defaults

import (
	"time"

	"github.com/shopspring/decimal"
)

// Indirect tax reform constants (Reforma Tributária, EC 132/2023).
//
// NT 2025.002-RTC / MOC 7.0 introduces two new VAT-style taxes on the NF-e:
//
//   - IBS — Imposto sobre Bens e Serviços, state + municipal (replaces ICMS + ISS)
//   - CBS — Contribuição sobre Bens e Serviços, federal          (replaces PIS + COFINS)
//
// The XML groups (<IBSCBS> per item, <IBSCBSTot> per NF-e) become mandatory in
// the NF-e layout starting August 2026. Rates ramp up gradually until 2033.

// IBSCBS CST codes per NT 2025.002-RTC. "000" is the "tributado integralmente"
// default; the other codes denote specific exemptions, suspensions, or zero-rate
// regimes (see NT 2025.002-RTC for the full taxonomy).
const (
	IBSCBSCSTTributadaIntegral     = "000"
	IBSCBSCSTTributadaParcial     = "010"
	IBSCBSCSTIsenta                = "200"
	IBSCBSCSTNaoTributada          = "400"
	IBSCBSCSTComSuspensao          = "510"
	IBSCBSCSTTributavelZero        = "600"
	IBSCBSCSTIsencaoCondicional    = "620"
	IBSCBSCSTSemIncidencia         = "800"
	IBSCBSCSTImunidade             = "810"
	IBSCBSCSTNaoTributavelFora     = "900"

	IBSCBSTributados = "000,010,200,400,510,600,620,800,810,900"
)

// 2026 symbolic rates per Ato Conjunto RFB/CGIBS. Stored as decimal *rates*,
// not percentages — the XML expects pIBS / pCBS as percentage values
// (e.g. 0.9000 = 0.9 %), so we multiply by 100 at render time.
var (
	CBSRate2026 = decimal.RequireFromString("0.009") // 0.9 % → XML pCBS = 0.90
	IBSRate2026 = decimal.RequireFromString("0.001") // 0.1 % → XML pIBS = 0.10

	CBSRate2026Pct = decimal.NewFromFloat(0.9) // display only
	IBSRate2026Pct = decimal.NewFromFloat(0.1) // display only

	CClassTribDefault = "000001" // cClassTrib default per NT 2025.002-RTC table
)

// IsTaxReformActive reports whether the indirect tax reform is in effect for
// the given timestamp. Returns true from 2026-01-01 onwards (when NF-e must
// carry IBS/CBS groups). Pre-2026 emissions still emit a zero-valued <IBSCBS>
// block so downstream consumers see stable schema, but the totals stay zero.
func IsTaxReformActive(t time.Time) bool {
	return t.Year() >= 2026
}

// IsIBSCBSTributado reports whether the given CST is in the IBSCBS scheme
// (i.e. the per-item <IBSCBS> group must be emitted). For the 2026 symbolic
// phase all 10 NT 2025.002-RTC codes are treated as "in scheme" so the XML
// always carries the schema-required <gIBSCBS> subgroup, even when vIBS and
// vCBS are zero (exempt, immune, suspended, etc.).
func IsIBSCBSTributado(cst string) bool {
	for _, c := range []string{
		IBSCBSCSTTributadaIntegral, IBSCBSCSTTributadaParcial,
		IBSCBSCSTIsenta, IBSCBSCSTNaoTributada, IBSCBSCSTComSuspensao,
		IBSCBSCSTTributavelZero, IBSCBSCSTIsencaoCondicional,
		IBSCBSCSTSemIncidencia, IBSCBSCSTImunidade, IBSCBSCSTNaoTributavelFora,
	} {
		if cst == c {
			return true
		}
	}
	return false
}