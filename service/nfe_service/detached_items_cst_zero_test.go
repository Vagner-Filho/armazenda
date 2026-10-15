package nfe_service

import (
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// TestBuildDetachedItems_NonDeclaringCSTValuesAreZero pins the totals
// contract regressed by SEFAZ Rej. 531 (G4 homolog replay): ICMS 40/41/50
// items declare no vBC/pICMS/vICMS and PIS/COFINS 04–09 declare no
// vBC/percentage/value, so the entity values must stay zero (the XML
// layer omits them and ICMSTot sums only declared ones). CSTs whose
// groups do declare values (00) keep the farm-rate behavior.
func TestBuildDetachedItems_NonDeclaringCSTValuesAreZero(t *testing.T) {
	svc := NewNFeService()

	cfg := &entity_public.FarmConfig{
		TaxRegime:       3, // normal (non-Simples) regime
		DefaultModFrete: 9,
		ICMSRate:        decimal.RequireFromString("0.17"),
		PISRate:         decimal.RequireFromString("0.01"),
		COFINSRate:      decimal.RequireFromString("0.07"),
		IBSRate:         decimal.RequireFromString("0.001"),
		CBSRate:         decimal.RequireFromString("0.009"),
	}

	newItem := func() []DetachedItemInput {
		return []DetachedItemInput{
			{
				ProductName: "Veneziana",
				NCM:         "84339090",
				CFOP:        "5913",
				Quantity:    decimal.RequireFromString("1"),
				UnitPrice:   decimal.RequireFromString("600"),
				Unit:        "UN",
			},
		}
	}

	strPtr := func(s string) *string { return &s }

	for _, cst := range []string{"40", "41", "50"} {
		t.Run("icms_cst_"+cst+"_declares_nothing", func(t *testing.T) {
			overrides := &entity.InvoiceOverrides{
				ICMSCST:   strPtr(cst),
				PISCST:    strPtr("08"),
				COFINSCST: strPtr("08"),
			}
			items, _, _, _, _ := svc.buildDetachedItems(
				newItem(), cfg, resolvedSnapshot(t, entity.TaxRates{}, overrides, cfg), overrides)
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
			if !items[0].Imposto.IBSCBS.VBC.Equal(decimal.RequireFromString("600")) {
				t.Errorf("CST %s: IBS vBC must stay 600, got %s", cst, items[0].Imposto.IBSCBS.VBC)
			}
		})
	}

	t.Run("icms_cst_and_pis_cofins_declaring_keep_farm_rate", func(t *testing.T) {
		overrides := &entity.InvoiceOverrides{
			ICMSCST:   strPtr("00"),
			PISCST:    strPtr("01"),
			COFINSCST: strPtr("01"),
		}
		items, _, _, _, _ := svc.buildDetachedItems(
			newItem(), cfg, resolvedSnapshot(t, entity.TaxRates{}, overrides, cfg), overrides)
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}

		icms := items[0].Imposto.ICMS
		if !icms.VBC.Equal(decimal.RequireFromString("600")) {
			t.Errorf("declaring CST: VBC = %s, want 600", icms.VBC)
		}
		if !icms.VICMS.Equal(decimal.RequireFromString("102")) {
			t.Errorf("declaring CST: VICMS = %s, want 102 (600 × 17%%)", icms.VICMS)
		}
		if !icms.PICMS.Equal(decimal.RequireFromString("17")) {
			t.Errorf("declaring CST: PICMS = %s, want 17", icms.PICMS)
		}

		pis := items[0].Imposto.PIS
		if !pis.VBC.Equal(decimal.RequireFromString("600")) || !pis.VPIS.Equal(decimal.RequireFromString("6")) {
			t.Errorf("declaring PIS: VBC/VPIS = %s/%s, want 600/6", pis.VBC, pis.VPIS)
		}

		cofins := items[0].Imposto.COFINS
		if !cofins.VBC.Equal(decimal.RequireFromString("600")) || !cofins.VCOFINS.Equal(decimal.RequireFromString("42")) {
			t.Errorf("declaring COFINS: VBC/VCOFINS = %s/%s, want 600/42", cofins.VBC, cofins.VCOFINS)
		}
	})
}
