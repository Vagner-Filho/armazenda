package nfe_service

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
)

// snapshotRates builds an entity.TaxRates with all five axes explicitly set.
func snapshotRates(icms, pis, cofins, ibs, cbs string) entity.TaxRates {
	return entity.TaxRates{
		ICMSRate:   ratePtr(icms),
		PISRate:    ratePtr(pis),
		COFINSRate: ratePtr(cofins),
		IBSRate:    ratePtr(ibs),
		CBSRate:    ratePtr(cbs),
	}
}

func ratePtr(s string) *decimal.Decimal {
	d := decimal.RequireFromString(s)
	return &d
}

func codePtr(s string) *string {
	v := s
	return &v
}

// snapshotFarmConfig returns a farm config with every rate configured and
// the standard default CSTs, the happy-path resolution source.
func snapshotFarmConfig() *entity_public.FarmConfig {
	return &entity_public.FarmConfig{
		ICMSRate:         decimal.RequireFromString("0.17"),
		PISRate:          decimal.RequireFromString("0.0165"),
		COFINSRate:       decimal.RequireFromString("0.076"),
		IBSRate:          decimal.RequireFromString("0.001"),
		CBSRate:          decimal.RequireFromString("0.009"),
		DefaultICMSCST:   codePtr("00"),
		DefaultPISCST:    codePtr("01"),
		DefaultCOFINSCST: codePtr("01"),
	}
}

// unconfiguredFarmConfig has zero rates (NULL columns scan as zero) and no
// default CSTs — the "never configured" state.
func unconfiguredFarmConfig() *entity_public.FarmConfig {
	return &entity_public.FarmConfig{}
}

// ratesEqual compares two resolved rate sets field by field.
func ratesEqual(t *testing.T, got, want entity.TaxRates) {
	t.Helper()
	fields := []struct {
		name      string
		got, want *decimal.Decimal
	}{
		{"ICMS", got.ICMSRate, want.ICMSRate},
		{"PIS", got.PISRate, want.PISRate},
		{"COFINS", got.COFINSRate, want.COFINSRate},
		{"IBS", got.IBSRate, want.IBSRate},
		{"CBS", got.CBSRate, want.CBSRate},
	}
	for _, f := range fields {
		if (f.got == nil) != (f.want == nil) {
			t.Errorf("%s: nil mismatch — got %v, want %v", f.name, f.got, f.want)
			continue
		}
		if f.got != nil && !f.got.Equal(*f.want) {
			t.Errorf("%s: got %s, want %s", f.name, f.got, *f.want)
		}
	}
}

func TestMergeRates_SnapshotIsIdempotent(t *testing.T) {
	cfg := snapshotFarmConfig()

	// Resolve once: user rates win over the farm arm.
	user := snapshotRates("0.07", "0.0165", "0.076", "0.001", "0.009")
	merged := MergeRates(user, cfg)

	// Feeding the resolved snapshot back through MergeRates (as the worker
	// rebuild and the SVC attempt do) must return it unchanged.
	again := MergeRates(merged, cfg)
	ratesEqual(t, again, merged)

	// A farm-only resolution must also be idempotent (the snapshot overrides
	// everything when complete).
	farmOnly := MergeRates(entity.TaxRates{}, cfg)
	againFarm := MergeRates(farmOnly, cfg)
	ratesEqual(t, againFarm, farmOnly)

	// The user-typed rate must win over the farm config in the resolution.
	if !merged.ICMSRate.Equal(decimal.RequireFromString("0.07")) {
		t.Errorf("user rate should win over farm config: got %s", merged.ICMSRate)
	}
	if !farmOnly.ICMSRate.Equal(decimal.RequireFromString("0.17")) {
		t.Errorf("farm rate should be used when the user types nothing: got %s", farmOnly.ICMSRate)
	}
}

func TestMergeRates_MissingAxesSurfaceInValidation(t *testing.T) {
	merged := MergeRates(entity.TaxRates{}, unconfiguredFarmConfig())

	axes := map[string]*decimal.Decimal{
		"ICMS":   merged.ICMSRate,
		"PIS":    merged.PISRate,
		"COFINS": merged.COFINSRate,
		"IBS":    merged.IBSRate,
		"CBS":    merged.CBSRate,
	}
	for name, rate := range axes {
		if rate != nil {
			t.Errorf("%s: rate must come back nil (unresolvable), got %s", name, rate)
		}
	}

	errs := validateResolvedRates(merged, defaults.TaxRegimeLucroReal)
	if len(errs) != 5 {
		t.Fatalf("expected 5 missing-rate errors for an unconfigured farm, got %d: %v", len(errs), errs)
	}
}

func TestValidateResolvedRates_SimpleNacionalICMSOptional(t *testing.T) {
	// Simples Nacional never consumes the ICMS rate: unconfigured ICMS with
	// every other axis resolved must validate clean.
	rates := entity.TaxRates{
		ICMSRate:   nil,
		PISRate:    ratePtr("0.0165"),
		COFINSRate: ratePtr("0.076"),
		IBSRate:    ratePtr("0.001"),
		CBSRate:    ratePtr("0.009"),
	}
	if errs := validateResolvedRates(rates, defaults.TaxRegimeSimplesNacional); len(errs) != 0 {
		t.Errorf("SN regime must not require the ICMS rate, got errors: %v", errs)
	}

	// Regime Normal (Lucro Real here) does require ICMS.
	errs := validateResolvedRates(rates, defaults.TaxRegimeLucroReal)
	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error (ICMS) for a normal-regime farm, got %d: %v", len(errs), errs)
	}
	if want := "Alíquota ICMS"; !strings.Contains(errs[0], want) {
		t.Errorf("error should mention the ICMS axis: %v", errs)
	}
}

func TestValidateResolvedRates_UserZeroIsExempt(t *testing.T) {
	// A user-typed zero is an explicit exemption, not a missing rate.
	rates := entity.TaxRates{ICMSRate: ratePtr("0")}
	errs := validateResolvedRates(rates, defaults.TaxRegimeLucroReal)
	for _, e := range errs {
		if strings.Contains(e, "ICMS") {
			t.Errorf("a typed zero must not be reported as missing ICMS: %v", errs)
		}
	}
	if len(errs) != 4 {
		t.Fatalf("expected the 4 remaining axes to be reported missing, got %d: %v", len(errs), errs)
	}
}

func TestValidateResolvedRates_RequiredInBothRegimes(t *testing.T) {
	for _, regime := range []defaults.TaxRegime{defaults.TaxRegimeSimplesNacional, defaults.TaxRegimeLucroReal} {
		// An entirely empty rate set reports every axis the regime uses:
		// 5 for the normal regimes, 4 for SN (ICMS not consumed there).
		want := 4
		if regime != defaults.TaxRegimeSimplesNacional {
			want = 5
		}
		errs := validateResolvedRates(entity.TaxRates{}, regime)
		if len(errs) != want {
			t.Errorf("regime %v: expected %d errors, got %d: %v", regime, want, len(errs), errs)
		}
	}
}

func TestMergeEffectiveOverrides_PriorityUserOverFarmOverDefault(t *testing.T) {
	cfg := snapshotFarmConfig()

	// User codes win over farm defaults.
	resolved := MergeEffectiveOverrides(&entity.InvoiceOverrides{
		ICMSCST:   codePtr("20"),
		PISCST:    codePtr("04"),
		COFINSCST: codePtr("07"),
		CBSCST:    codePtr("200"),
	}, cfg, defaults.TaxRegimeLucroReal)
	assertCode(t, resolved.ICMSCST, "20", "ICMSCST user")
	assertCode(t, resolved.PISCST, "04", "PISCST user")
	assertCode(t, resolved.COFINSCST, "07", "COFINSCST user")
	assertCode(t, resolved.CBSCST, "200", "CBSCST user")

	// Farm defaults fill in where the user typed nothing.
	resolved = MergeEffectiveOverrides(nil, cfg, defaults.TaxRegimeLucroReal)
	assertCode(t, resolved.ICMSCST, "00", "ICMSCST farm default")
	assertCode(t, resolved.PISCST, "01", "PISCST farm default")
	assertCode(t, resolved.COFINSCST, "01", "COFINSCST farm default")

	// Builder defaults fill in where the farm has no default.
	resolved = MergeEffectiveOverrides(nil, unconfiguredFarmConfig(), defaults.TaxRegimeLucroReal)
	assertCode(t, resolved.ICMSCST, "00", "ICMSCST builder default")
	assertCode(t, resolved.PISCST, "01", "PISCST builder default")
	assertCode(t, resolved.COFINSCST, "01", "COFINSCST builder default")

	// The shared IBS/CBS CST default applies in every case.
	assertCode(t, resolved.CBSCST, "000", "CBSCST default")
	assertCode(t, resolved.CClassTrib, "000001", "CClassTrib default")
}

func TestMergeEffectiveOverrides_ICMSResolvesPerRegime(t *testing.T) {
	// Without any user or farm value, the ICMS code resolves to CSOSN under
	// Simples Nacional and to CST under the normal regimes.
	resolved := MergeEffectiveOverrides(nil, unconfiguredFarmConfig(), defaults.TaxRegimeSimplesNacional)
	assertCode(t, resolved.ICMSCST, "102", "ICMSCST SN default (CSOSN)")

	resolved = MergeEffectiveOverrides(nil, unconfiguredFarmConfig(), defaults.TaxRegimeLucroPresumido)
	assertCode(t, resolved.ICMSCST, "00", "ICMSCST normal default (CST)")
}

func TestMergeEffectiveOverrides_OnlyIBSCSTTypedFeedsSharedCST(t *testing.T) {
	// A user typing only the IBS CST must see it carried into the shared
	// per-item <CST> element (builder precedence: CBSCST > IBSCST).
	resolved := MergeEffectiveOverrides(&entity.InvoiceOverrides{IBSCST: codePtr("111")}, nil, defaults.TaxRegimeLucroReal)
	assertCode(t, resolved.CBSCST, "111", "IBSCST typed alone")
}

func TestMergeEffectiveOverrides_UserFieldsCopiedThrough(t *testing.T) {
	// The helper resolves only the tax fields; every other user field must
	// survive the copy so the create functions persist them.
	user := &entity.InvoiceOverrides{
		NaturezaOp:  codePtr("Venda de produção"),
		ProductDesc: codePtr("Soja em grão"),
		Unit:        codePtr("SC"),
		ModFrete:    intPtr(2),
		ICMSCST:     codePtr("20"),
	}
	resolved := MergeEffectiveOverrides(user, snapshotFarmConfig(), defaults.TaxRegimeLucroReal)

	if resolved.NaturezaOp == nil || *resolved.NaturezaOp != "Venda de produção" {
		t.Errorf("NaturezaOp must be copied through, got %v", resolved.NaturezaOp)
	}
	if resolved.ProductDesc == nil || *resolved.ProductDesc != "Soja em grão" {
		t.Errorf("ProductDesc must be copied through, got %v", resolved.ProductDesc)
	}
	if resolved.Unit == nil || *resolved.Unit != "SC" {
		t.Errorf("Unit must be copied through, got %v", resolved.Unit)
	}
	if resolved.ModFrete == nil || *resolved.ModFrete != 2 {
		t.Errorf("ModFrete must be copied through, got %v", resolved.ModFrete)
	}
	assertCode(t, resolved.ICMSCST, "20", "ICMSCST user")
}

func TestMergeEffectiveOverrides_IdempotentOnResolvedStruct(t *testing.T) {
	cfg := snapshotFarmConfig()
	first := MergeEffectiveOverrides(&entity.InvoiceOverrides{PISCST: codePtr("04")}, cfg, defaults.TaxRegimeLucroReal)
	second := MergeEffectiveOverrides(&first, cfg, defaults.TaxRegimeLucroReal)

	assertCode(t, second.PISCST, "04", "resolved PISCST stable")
	assertCode(t, second.ICMSCST, *first.ICMSCST, "resolved ICMSCST stable")
	assertCode(t, second.CBSCST, *first.CBSCST, "resolved CBSCST stable")
	assertCode(t, second.CClassTrib, *first.CClassTrib, "resolved CClassTrib stable")
}

func assertCode(t *testing.T, got *string, want, label string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: resolved code must never be nil, want %q", label, want)
		return
	}
	if *got != want {
		t.Errorf("%s: got %q, want %q", label, *got, want)
	}
}

func intPtr(i int) *int {
	v := i
	return &v
}

// resolvedSnapshot runs the exact preparation the service prepare steps use
// (merge → validation → codes) so `buildDetachedItems`/`buildItems` test
// fixtures mirror the production path. Tests with unresolvable fixture rates
// fail here with the validation message instead of dereferencing nils.
func resolvedSnapshot(t *testing.T, userRates entity.TaxRates, overrides *entity.InvoiceOverrides, cfg *entity_public.FarmConfig) emissionTaxSnapshot {
	t.Helper()
	rates := MergeRates(userRates, cfg)
	regime := defaults.TaxRegime(cfg.TaxRegime)
	if errs := validateResolvedRates(rates, regime); len(errs) > 0 {
		t.Fatalf("test fixture has unresolvable rates: %v", errs)
	}
	return emissionTaxSnapshot{Rates: rates, Codes: MergeEffectiveOverrides(overrides, cfg, regime)}
}
