package defaults_test

import (
	"testing"
	"time"

	"armazenda/pkg/nfe/defaults"
)

func TestIsTaxReformActive(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{"2025-12-31_23-59-59", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC), false},
		{"2026-01-01_00-00-00", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"2026-08-01_mandatory_fields", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), true},
		{"2027-01-01_cbs_active", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaults.IsTaxReformActive(tt.date); got != tt.expected {
				t.Errorf("IsTaxReformActive(%s) = %v, want %v", tt.date.Format(time.RFC3339), got, tt.expected)
			}
		})
	}
}

func TestTaxReform2026Rates(t *testing.T) {
	if got := defaults.CBSRate2026Pct.InexactFloat64(); got != 0.9 {
		t.Errorf("CBSRate2026Pct = %v, want 0.9", got)
	}
	if got := defaults.IBSRate2026Pct.InexactFloat64(); got != 0.1 {
		t.Errorf("IBSRate2026Pct = %v, want 0.1", got)
	}
	// Decimal *rates* (the values stored on invoices / farm config) are
	// the percentage divided by 100 — XML pIBS / pCBS multiply by 100.
	if got := defaults.CBSRate2026.InexactFloat64(); got != 0.009 {
		t.Errorf("CBSRate2026 = %v, want 0.009", got)
	}
	if got := defaults.IBSRate2026.InexactFloat64(); got != 0.001 {
		t.Errorf("IBSRate2026 = %v, want 0.001", got)
	}
}

func TestIBSCBSCSTs(t *testing.T) {
	// Every NT 2025.002-RTC CST code must round-trip via IsIBSCBSTributado.
	csts := []string{
		defaults.IBSCBSCSTTributadaIntegral,    // "000"
		defaults.IBSCBSCSTTributadaParcial,     // "010"
		defaults.IBSCBSCSTIsenta,               // "200"
		defaults.IBSCBSCSTNaoTributada,         // "400"
		defaults.IBSCBSCSTComSuspensao,         // "510"
		defaults.IBSCBSCSTTributavelZero,       // "600"
		defaults.IBSCBSCSTIsencaoCondicional,   // "620"
		defaults.IBSCBSCSTSemIncidencia,        // "800"
		defaults.IBSCBSCSTImunidade,            // "810"
		defaults.IBSCBSCSTNaoTributavelFora,    // "900"
	}
	if len(csts) != 10 {
		t.Errorf("expected 10 IBSCBS CST codes, got %d", len(csts))
	}
	for _, c := range csts {
		if c == "" {
			t.Error("encountered empty IBSCBS CST code")
		}
		if !defaults.IsIBSCBSTributado(c) {
			t.Errorf("IsIBSCBSTributado(%q) = false, want true", c)
		}
	}
	// Unknown code → false
	if defaults.IsIBSCBSTributado("999") {
		t.Error("IsIBSCBSTributado(999) should be false (unknown code)")
	}
}

func TestCClassTribDefault(t *testing.T) {
	if got := defaults.CClassTribDefault; got != "000001" {
		t.Errorf("CClassTribDefault = %q, want %q", got, "000001")
	}
}