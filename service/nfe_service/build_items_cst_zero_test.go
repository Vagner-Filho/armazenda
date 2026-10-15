package nfe_service

import (
	"testing"

	"armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// TestBuildItems_NonDeclaringCSTValuesAreZero pins the same totals
// contract as the detached flow (docs/sdd/nfe_cst_zero_values) on the
// departure flow: items with ICMS 40/41/50 or PIS/COFINS 04–09 carry
// zero entity values because their XML groups declare nothing; CSTs
// whose groups do declare values keep the resolved rates.
func TestBuildItems_NonDeclaringCSTValuesAreZero(t *testing.T) {
	svc := NewNFeService()

	departure := entity_public.Departure{
		CargoWeight: entity_public.CargoWeight{
			GrossWeight: decimal.RequireFromString("11"),
			Tare:        decimal.RequireFromString("1"),
			NetWeight:   decimal.RequireFromString("10"),
		},
	}
	unitPrice := decimal.RequireFromString("100")
	// Item total = NetWeight × unitPrice = 1000.

	rate := func(s string) *decimal.Decimal {
		d := decimal.RequireFromString(s)
		return &d
	}
	rates := entity.TaxRates{
		ICMSRate:   rate("0.17"),
		PISRate:    rate("0.01"),
		COFINSRate: rate("0.07"),
		IBSRate:    rate("0.001"),
		CBSRate:    rate("0.009"),
	}
	strPtr := func(s string) *string { return &s }

	for _, cst := range []string{"40", "41", "50"} {
		t.Run("icms_cst_"+cst+"_declares_nothing", func(t *testing.T) {
			codes := entity.InvoiceOverrides{
				ICMSCST:   strPtr(cst),
				PISCST:    strPtr("08"),
				COFINSCST: strPtr("08"),
			}
			items := svc.buildItems(departure, unitPrice, "84339090", "", "UN",
				"Veneziana", defaults.TaxRegime(3), "5913", rates, codes)
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			icms := items[0].Imposto.ICMS
			if !icms.VBC.IsZero() || !icms.VICMS.IsZero() || !icms.PICMS.IsZero() {
				t.Errorf("CST %s: ICMS values must be zero, got VBC=%s VICMS=%s PICMS=%s",
					cst, icms.VBC, icms.VICMS, icms.PICMS)
			}

			pis := items[0].Imposto.PIS
			if !pis.VBC.IsZero() || !pis.PPIS.IsZero() || !pis.VPIS.IsZero() {
				t.Errorf("CST %s: PIS values must be zero, got VBC=%s PPIS=%s VPIS=%s",
					cst, pis.VBC, pis.PPIS, pis.VPIS)
			}

			cofins := items[0].Imposto.COFINS
			if !cofins.VBC.IsZero() || !cofins.PCOFINS.IsZero() || !cofins.VCOFINS.IsZero() {
				t.Errorf("CST %s: COFINS values must be zero, got VBC=%s PCOFINS=%s VCOFINS=%s",
					cst, cofins.VBC, cofins.PCOFINS, cofins.VCOFINS)
			}

			// IBS/CBS values follow their own (always-declaring) group.
			if !items[0].Imposto.IBSCBS.VBC.Equal(decimal.RequireFromString("1000")) {
				t.Errorf("CST %s: IBS vBC must stay 1000, got %s", cst, items[0].Imposto.IBSCBS.VBC)
			}
		})
	}

	t.Run("declaring_csts_keep_rates", func(t *testing.T) {
		codes := entity.InvoiceOverrides{
			ICMSCST:   strPtr("00"),
			PISCST:    strPtr("01"),
			COFINSCST: strPtr("01"),
		}
		items := svc.buildItems(departure, unitPrice, "84339090", "", "UN",
			"Veneziana", defaults.TaxRegime(3), "5913", rates, codes)
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}

		icms := items[0].Imposto.ICMS
		if !icms.VBC.Equal(decimal.RequireFromString("1000")) {
			t.Errorf("declaring CST: VBC = %s, want 1000", icms.VBC)
		}
		if !icms.VICMS.Equal(decimal.RequireFromString("170")) {
			t.Errorf("declaring CST: VICMS = %s, want 170 (1000 × 17%%)", icms.VICMS)
		}
		if !icms.PICMS.Equal(decimal.RequireFromString("17")) {
			t.Errorf("declaring CST: PICMS = %s, want 17", icms.PICMS)
		}

		pis := items[0].Imposto.PIS
		if !pis.VBC.Equal(decimal.RequireFromString("1000")) || !pis.VPIS.Equal(decimal.RequireFromString("10")) {
			t.Errorf("declaring PIS: VBC/VPIS = %s/%s, want 1000/10", pis.VBC, pis.VPIS)
		}

		cofins := items[0].Imposto.COFINS
		if !cofins.VBC.Equal(decimal.RequireFromString("1000")) || !cofins.VCOFINS.Equal(decimal.RequireFromString("70")) {
			t.Errorf("declaring COFINS: VBC/VCOFINS = %s/%s, want 1000/70", cofins.VBC, cofins.VCOFINS)
		}
	})
}
