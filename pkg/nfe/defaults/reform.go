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

// 2026 symbolic rates per Ato Conjunto RFB/CGIBS.
//
// Decimal-rate convention: the value stored here and in nfe_farm_config.ibs_rate
// / cbs_rate is the *decimal* rate (percentage / 100). The XML schema
// expects the rate as a decimal (e.g. 0.001 for 0.1 %); the builder
// multiplies by 100 at render time to produce the percentage value the
// schema requires (e.g. 0.10 in <pIBSUF>).
//
// This convention is enforced by:
//   - router/nfe_router/router.go::parsePercentRateOrNil: divides the
//     form-submitted percentage string by 100 before storing.
//   - migration 000019: nfe_farm_config.ibs_rate/cbs_rate columns default
//     to 0.0010 / 0.0090 (not 0.1000 / 0.9000).
//
// CAUTION: the previous migration defaults (0.1000/0.9000) were 100x too
// large and produced SEFAZ rejection 1026 ("Alíquota do IBS da UF inválida").
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

// Canonical element orders for the IBS/CBS XML groups. SEFAZ rejects the NF-e
// (status 215, cvc-complex-type.2.4.a) when children are missing or out of
// order. These constants anchor the schema to the source of truth — verify any
// changes against NT 2025.002-RTC v1.51 §6.7.4 (Grupo UB / Grupo W03) and
// the NFe_Util 2Gv5.02b reference output.
//
// Per-item <IBSCBS>/<gIBSCBS> (Group UB, NT 2025.002-RTC v1.51):
//
//	IBSCBS:    CST, cClassTrib, [indDoacao 0-1], gIBSCBS
//	gIBSCBS:   vBC, gIBSUF, gIBSMun, vIBS, gCBS                  (FLAT sequence — no <gIBS> wrapper)
//	gIBSUF:    pIBSUF, [optional inner seq], vIBSUF
//	gIBSMun:   pIBSMun, [optional inner seq], vIBSMun
//	gCBS:      pCBS, [optional inner seq], vCBS
//
// KEY POINT: the per-item block has NO <gIBS> wrapper. <gIBSUF>, <gIBSMun>,
// <vIBS>, and <gCBS> are direct siblings of each other inside <gIBSCBS>.
// Adding a <gIBS> wrapper here produces SEFAZ 215 with cvc-complex-type.2.4.a
// ("Invalid content starting with element 'gIBS'. One of 'gIBSUF' is
// expected"). <vIBS> (UB54a, per-item IBS total = vIBSUF + vIBSMun) is
// mandatory 1-1 as a direct sibling.
//
// Per-item <vCredPres> and <vCredPresCondSus> do NOT appear as direct
// children of <gCBS> — those fields are totals-only. Per-item credit-
// presumption fields belong inside the optional <gIBSCredPres> and
// <gCBSCredPres> subgroups (skipped in the 2026 symbolic phase).
//
// Totals <IBSCBSTot>/<gIBS>/<gCBS> (Group W03, NT 2025.002-RTC v1.51):
//
//	IBSCBSTot: vBCIBSCBS, gIBS, gCBS, [gMono OPTIONAL 0-1]
//	gIBS:      gIBSUF, gIBSMun, vIBS, vCredPres, vCredPresCondSus
//	gIBSUF:    vDif, vDevTrib, vIBSUF
//	gIBSMun:   vDif, vDevTrib, vIBSMun
//	gCBS:      vDif, vDevTrib, vCBS, vCredPres, vCredPresCondSus
//
// KEY POINT: the totals block DOES use a <gIBS> wrapper. This asymmetry with
// the per-item block is by design in the NT schema.
const (
	IBSCBSTotGCBSOrder       = "vDif,vDevTrib,vCBS,vCredPres,vCredPresCondSus"
	IBSCBSTotGIBSUFOrder     = "vDif,vDevTrib,vIBSUF"
	IBSCBSTotGIBSMunOrder    = "vDif,vDevTrib,vIBSMun"
	IBSCBSTotGIBSOrder       = "gIBSUF,gIBSMun,vIBS,vCredPres,vCredPresCondSus"
	IBSCBSTotRootOrder       = "vBCIBSCBS,gIBS,gCBS"
	// Per-item <gIBSCBS> is a FLAT sequence per NT v1.51 §6.7.4 UB15:
	//   gIBSCBS: vBC, gIBSUF, gIBSMun, vIBS, gCBS
	// <vIBS> (UB54a) is the per-item IBS total, mandatory 1-1, direct sibling.
	PerItemGIBSCBSOrder       = "vBC,gIBSUF,gIBSMun,vIBS,gCBS"
	PerItemGIBSUFOrder        = "pIBSUF,vIBSUF"
	PerItemGIBSMunOrder       = "pIBSMun,vIBSMun"
	PerItemGCBSOrder          = "pCBS,vCBS"
	// PerItemGIBSOrder removed: per-item has no <gIBS> wrapper.
)