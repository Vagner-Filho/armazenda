package nfe_service

import (
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

func TestMergeRates_UnconfiguredFarmResolvesDefaults(t *testing.T) {
	merged := MergeRates(entity.TaxRates{}, unconfiguredFarmConfig())

	if merged.ICMSRate == nil || !merged.ICMSRate.IsZero() {
		t.Errorf("ICMS: got %v, want 0", merged.ICMSRate)
	}
	if merged.PISRate == nil || !merged.PISRate.IsZero() {
		t.Errorf("PIS: got %v, want 0", merged.PISRate)
	}
	if merged.COFINSRate == nil || !merged.COFINSRate.IsZero() {
		t.Errorf("COFINS: got %v, want 0", merged.COFINSRate)
	}
	if merged.IBSRate == nil || !merged.IBSRate.Equal(defaults.IBSRate2026) {
		t.Errorf("IBS: got %v, want %s", merged.IBSRate, defaults.IBSRate2026)
	}
	if merged.CBSRate == nil || !merged.CBSRate.Equal(defaults.CBSRate2026) {
		t.Errorf("CBS: got %v, want %s", merged.CBSRate, defaults.CBSRate2026)
	}
}

func TestMergeRates_PrecedenceUserZeroBeatsFarmAndConstants(t *testing.T) {
	cfg := snapshotFarmConfig()
	// User-typed zero wins over a non-zero farm config and the constants.
	user := snapshotRates("0", "0", "0", "0", "0")
	merged := MergeRates(user, cfg)
	if !merged.ICMSRate.Equal(decimal.Zero) {
		t.Errorf("ICMS user zero should win, got %s", merged.ICMSRate)
	}
	if !merged.IBSRate.Equal(decimal.Zero) {
		t.Errorf("IBS user zero should win over constant, got %s", merged.IBSRate)
	}
}

func TestMergeRates_FarmNonZeroBeatsConstants(t *testing.T) {
	cfg := snapshotFarmConfig()
	merged := MergeRates(entity.TaxRates{}, cfg)
	if !merged.ICMSRate.Equal(decimal.RequireFromString("0.17")) {
		t.Errorf("ICMS farm should win, got %s", merged.ICMSRate)
	}
	if !merged.IBSRate.Equal(decimal.RequireFromString("0.001")) {
		t.Errorf("IBS farm should win, got %s", merged.IBSRate)
	}
}

func TestMergeRates_UserBeatsFarm(t *testing.T) {
	cfg := snapshotFarmConfig()
	user := snapshotRates("0.07", "0.01", "0.02", "0.002", "0.018")
	merged := MergeRates(user, cfg)
	if !merged.ICMSRate.Equal(decimal.RequireFromString("0.07")) {
		t.Errorf("ICMS user should beat farm, got %s", merged.ICMSRate)
	}
	if !merged.IBSRate.Equal(decimal.RequireFromString("0.002")) {
		t.Errorf("IBS user should beat farm, got %s", merged.IBSRate)
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
// (merge → codes) so `buildDetachedItems`/`buildItems` test fixtures mirror
// the production path. The rates waterfall never leaves an axis unresolved,
// so no rate validation is needed here.
func resolvedSnapshot(t *testing.T, userRates entity.TaxRates, overrides *entity.InvoiceOverrides, cfg *entity_public.FarmConfig) emissionTaxSnapshot {
	t.Helper()
	rates := MergeRates(userRates, cfg)
	regime := defaults.TaxRegime(cfg.TaxRegime)
	return emissionTaxSnapshot{Rates: rates, Codes: MergeEffectiveOverrides(overrides, cfg, regime)}
}
