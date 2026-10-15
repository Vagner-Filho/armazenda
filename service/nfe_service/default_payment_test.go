package nfe_service

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestDefaultPayment_IndPagOmitted pins the à-vista default regressed by the
// homolog comparison (docs/research_nfe_use_case/HOMOLOG_REPLAY_COMPARISON.md,
// Findings #3): the previously hardcoded indPag=1 produced "a prazo" detPag
// tags on à-vista (tPag 01) and no-payment (tPag 90) documents — accepted by
// MT but semantically wrong, and unlike the reference emitter corpus
// (indPag in 3/115 docs). The builder omits indPag when 0, so the service
// must leave it zeroed.
func TestDefaultPayment_IndPagOmitted(t *testing.T) {
	pg := defaultPayment("01", decimal.RequireFromString("600.00"))

	if pg.IndPag != 0 {
		t.Errorf("Payment.IndPag = %d, want 0 (à vista default; the builder omits the tag)", pg.IndPag)
	}
	if len(pg.Detalhes) != 1 {
		t.Fatalf("expected 1 detPag detail, got %d", len(pg.Detalhes))
	}
	det := pg.Detalhes[0]
	if det.IndPag != 0 {
		t.Errorf("detPag.IndPag = %d, want 0 (à vista default)", det.IndPag)
	}
	if det.TPag != "01" {
		t.Errorf("detPag.TPag = %q, want 01", det.TPag)
	}
	if !det.VPag.Equal(decimal.RequireFromString("600.00")) {
		t.Errorf("detPag.VPag = %s, want 600.00", det.VPag)
	}

	// tPag 90 (sem pagamento), the departure-flow shape, must behave the same.
	pg90 := defaultPayment("90", decimal.Zero)
	if pg90.IndPag != 0 || pg90.Detalhes[0].IndPag != 0 {
		t.Error("tPag 90 must also keep indPag zeroed (tag omitted)")
	}
}
