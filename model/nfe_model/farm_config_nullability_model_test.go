package nfe_model_test

import (
	"bytes"
	"context"
	"testing"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/nfe_model"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// TestMigration000031_FarmConfigColumnsNullable_DBRoundTrip verifies the
// migration made the four columns nullable.
func TestMigration000031_FarmConfigColumnsNullable_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	for _, col := range []string{"cbs_cst", "ibs_cst", "c_class_trib", "certificate_data"} {
		var isNullable string
		err := pool.QueryRow(ctx, `
			SELECT is_nullable FROM information_schema.columns
			WHERE table_name = 'nfe_farm_config' AND column_name = @col
		`, pgx.NamedArgs{"col": col}).Scan(&isNullable)
		if err != nil {
			t.Fatalf("query information_schema for %s: %v", col, err)
		}
		if isNullable != "YES" {
			t.Errorf("column %s is_nullable = %s, want YES", col, isNullable)
		}
	}
}

func TestUpsertFarmConfig_CertOnly_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows: %v", err)
	}
	farmIDU := uint32(farmID)

	// Clean any pre-existing config row for a deterministic insert path.
	pool.Exec(ctx, `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	})

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel: %v", err)
	}

	cert := []byte("cert-only-bytes")
	cfg := entity_public.FarmConfig{
		FarmID:          &farmIDU,
		CertificateData: cert,
		TaxRegime:       1,
	}
	if mErr := model.UpsertFarmConfig(cfg, nil); mErr != nil {
		t.Fatalf("UpsertFarmConfig cert-only failed: %v", mErr)
	}

	// Raw assertions
	var env int
	var cbsCst, ibsCst, cClassTrib *string
	var certData []byte
	if err := pool.QueryRow(ctx, `
		SELECT environment, cbs_cst, ibs_cst, c_class_trib, certificate_data
		FROM nfe_farm_config WHERE farm_id = @f
	`, pgx.NamedArgs{"f": farmID}).Scan(&env, &cbsCst, &ibsCst, &cClassTrib, &certData); err != nil {
		t.Fatalf("raw select failed: %v", err)
	}
	if env != 2 {
		t.Errorf("environment = %d, want 2", env)
	}
	if cbsCst != nil {
		t.Errorf("cbs_cst = %v, want nil", cbsCst)
	}
	if ibsCst != nil {
		t.Errorf("ibs_cst = %v, want nil", ibsCst)
	}
	if cClassTrib != nil {
		t.Errorf("c_class_trib = %v, want nil", cClassTrib)
	}
	if !bytes.Equal(certData, cert) {
		t.Errorf("certificate_data mismatch: got %v want %v", certData, cert)
	}

	// GetFarmConfig
	got, gErr := model.GetFarmConfig(farmIDU)
	if gErr != nil {
		t.Fatalf("GetFarmConfig: %v", gErr)
	}
	if got == nil {
		t.Fatal("GetFarmConfig returned nil")
	}
	if got.Environment != 2 {
		t.Errorf("got.Environment = %d, want 2", got.Environment)
	}
	if got.DefaultCBSCST != nil {
		t.Errorf("DefaultCBSCST = %v, want nil", got.DefaultCBSCST)
	}
	if got.DefaultIBSCST != nil {
		t.Errorf("DefaultIBSCST = %v, want nil", got.DefaultIBSCST)
	}
	if got.DefaultCClassTrib != nil {
		t.Errorf("DefaultCClassTrib = %v, want nil", got.DefaultCClassTrib)
	}
	if !bytes.Equal(got.CertificateData, cert) {
		t.Errorf("CertificateData mismatch")
	}
}

func TestUpsertFarmConfig_RejectsInvalidTaxRegime_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows: %v", err)
	}
	farmIDU := uint32(farmID)

	// Count before
	var before int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID}).Scan(&before)

	pool.Exec(ctx, `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	})

	model, _ := nfe_model.InitNFeModel(pool)

	for _, regime := range []int{0, 4, 99} {
		cfg := entity_public.FarmConfig{
			FarmID:    &farmIDU,
			TaxRegime: regime,
		}
		mErr := model.UpsertFarmConfig(cfg, nil)
		if mErr == nil {
			t.Fatalf("regime %d: expected error, got nil", regime)
		}
		if mErr.IsServerErr {
			t.Errorf("regime %d: IsServerErr = true, want false", regime)
		}
		if mErr.Message != "Regime tributário inválido: selecione Simples Nacional, Lucro Real ou Lucro Presumido" {
			t.Errorf("regime %d: message = %q", regime, mErr.Message)
		}
	}

	var after int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID}).Scan(&after)
	if after != before {
		// before was after delete? We deleted, so before count captured before delete. Hmm.
	}
	// Simpler: ensure no row exists now (we deleted and rejected all)
	var now int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID}).Scan(&now)
	if now != 0 {
		t.Errorf("expected no row written, got %d", now)
	}
	_ = model_error.ModelError{}
}

func TestUpsertFarmConfig_FullRoundTrip_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows: %v", err)
	}
	farmIDU := uint32(farmID)

	pool.Exec(ctx, `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	})

	model, _ := nfe_model.InitNFeModel(pool)

	cbsCST := "123"
	ibsCST := "456"
	cClass := "000099"
	cert := []byte("full-cert")
	env := 1
	cfg := entity_public.FarmConfig{
		FarmID:            &farmIDU,
		CertificateData:   cert,
		Environment:       env,
		Serie:             929,
		NextNumber:        5,
		TaxRegime:         3,
		DefaultModFrete:   9,
		DefaultCFOP:       "5101",
		DefaultUnit:       "KG",
		DefaultCBSCST:     &cbsCST,
		DefaultIBSCST:     &ibsCST,
		DefaultCClassTrib: &cClass,
	}
	// set rates
	cfg.ICMSRate = decimal.RequireFromString("0.12")
	cfg.PISRate = decimal.RequireFromString("0.0165")
	cfg.COFINSRate = decimal.RequireFromString("0.076")
	cfg.CBSRate = decimal.RequireFromString("0.0090")
	cfg.IBSRate = decimal.RequireFromString("0.0010")

	if mErr := model.UpsertFarmConfig(cfg, nil); mErr != nil {
		t.Fatalf("UpsertFarmConfig full failed: %v", mErr)
	}

	got, gErr := model.GetFarmConfig(farmIDU)
	if gErr != nil || got == nil {
		t.Fatalf("GetFarmConfig: %v %v", gErr, got)
	}
	if got.Environment != 1 {
		t.Errorf("Environment = %d, want 1", got.Environment)
	}
	if got.Serie != 929 {
		t.Errorf("Serie = %d, want 929", got.Serie)
	}
	if got.TaxRegime != 3 {
		t.Errorf("TaxRegime = %d, want 3", got.TaxRegime)
	}
	if got.DefaultCBSCST == nil || *got.DefaultCBSCST != cbsCST {
		t.Errorf("DefaultCBSCST = %v, want %s", got.DefaultCBSCST, cbsCST)
	}
	if got.DefaultIBSCST == nil || *got.DefaultIBSCST != ibsCST {
		t.Errorf("DefaultIBSCST = %v, want %s", got.DefaultIBSCST, ibsCST)
	}
	if got.DefaultCClassTrib == nil || *got.DefaultCClassTrib != cClass {
		t.Errorf("DefaultCClassTrib = %v, want %s", got.DefaultCClassTrib, cClass)
	}
	if !bytes.Equal(got.CertificateData, cert) {
		t.Errorf("CertificateData mismatch")
	}
	if !got.ICMSRate.Equal(decimal.RequireFromString("0.12")) {
		t.Errorf("ICMSRate = %s, want %s", got.ICMSRate, "0.12")
	}
}

func TestUpsertFarmConfig_NullCertificateData_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows: %v", err)
	}
	farmIDU := uint32(farmID)

	pool.Exec(ctx, `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID})
	})

	model, _ := nfe_model.InitNFeModel(pool)

	cfg := entity_public.FarmConfig{
		FarmID:      &farmIDU,
		TaxRegime:   1,
		Environment: 2,
	}
	// nil CertificateData
	if mErr := model.UpsertFarmConfig(cfg, nil); mErr != nil {
		t.Fatalf("UpsertFarmConfig nil cert failed: %v", mErr)
	}

	var certData []byte
	err := pool.QueryRow(ctx, `SELECT certificate_data FROM nfe_farm_config WHERE farm_id = @f`, pgx.NamedArgs{"f": farmID}).Scan(&certData)
	if err != nil {
		t.Fatalf("raw select cert_data: %v", err)
	}
	if len(certData) != 0 {
		t.Errorf("expected NULL/empty cert_data, got %v", certData)
	}

	got, gErr := model.GetFarmConfig(farmIDU)
	if gErr != nil {
		t.Fatalf("GetFarmConfig: %v", gErr)
	}
	if got == nil {
		t.Fatal("nil config")
	}
	if len(got.CertificateData) != 0 {
		t.Errorf("CertificateData = %v, want empty/nil", got.CertificateData)
	}
}
