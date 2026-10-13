package nfe_model_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// This file follows the invoice_lifecycle_model_test.go pattern: the tests are
// double-gated (full DB_* set + RUN_DB_TESTS=1) and MUST be run in separate
// `go test -run` invocations — InitNFeModel is a singleton, so two DB tests
// sharing one process would reuse the first test's (closed) pool.

// snapshotRates is the effective-rate snapshot a service prepare step would
// pass: all five axes resolved (user > farm — there is no fallback anymore).
func snapshotRates() *entity.TaxRates {
	icms := decimal.RequireFromString("0.17")
	pis := decimal.RequireFromString("0.0165")
	cofins := decimal.RequireFromString("0.076")
	ibs := decimal.RequireFromString("0.001")
	cbs := decimal.RequireFromString("0.009")
	return &entity.TaxRates{
		ICMSRate:   &icms,
		PISRate:    &pis,
		COFINSRate: &cofins,
		IBSRate:    &ibs,
		CBSRate:    &cbs,
	}
}

// snapshotCodes is the resolved tax-code snapshot a service prepare step
// would pass (MergeEffectiveOverrides output shape).
func snapshotCodes() *entity.InvoiceOverrides {
	icmsCST := "00"
	pisCST := "01"
	cofinsCST := "01"
	ibsCST := "000"
	cbsCST := "000"
	cClassTrib := "000001"
	return &entity.InvoiceOverrides{
		ICMSCST:    &icmsCST,
		PISCST:     &pisCST,
		COFINSCST:  &cofinsCST,
		IBSCST:     &ibsCST,
		CBSCST:     &cbsCST,
		CClassTrib: &cClassTrib,
	}
}

// assertInvoiceSnapshotRow reads the invoice row directly (the INSERT-side
// truth) plus the tax-rates row and asserts the full snapshot persisted.
func assertInvoiceSnapshotRow(t *testing.T, pool *pgxpool.Pool, ctx context.Context, invoiceID int) {
	t.Helper()

	var icmsCST, pisCST, cofinsCST, ibsCST, cbsCST, cClassTrib *string
	if err := pool.QueryRow(ctx,
		`SELECT icms_cst, pis_cst, cofins_cst, ibs_cst, cbs_cst, c_class_trib
		 FROM nfe_invoice WHERE id = @id`,
		pgx.NamedArgs{"id": invoiceID},
	).Scan(&icmsCST, &pisCST, &cofinsCST, &ibsCST, &cbsCST, &cClassTrib); err != nil {
		t.Fatalf("failed to read invoice code columns: %v", err)
	}
	codeColumns := map[string]*string{
		"icms_cst":     icmsCST,
		"pis_cst":      pisCST,
		"cofins_cst":   cofinsCST,
		"ibs_cst":      ibsCST,
		"cbs_cst":      cbsCST,
		"c_class_trib": cClassTrib,
	}
	for column, value := range codeColumns {
		if value == nil || *value == "" {
			t.Errorf("%s: expected the resolved code to be persisted, got %v", column, value)
		}
	}

	var rICMS, rPIS, rCOFINS, rIBS, rCBS *decimal.Decimal
	if err := pool.QueryRow(ctx,
		`SELECT icms_rate, pis_rate, cofins_rate, ibs_rate, cbs_rate
		 FROM nfe_invoice_tax_rates WHERE invoice_id = @id`,
		pgx.NamedArgs{"id": invoiceID},
	).Scan(&rICMS, &rPIS, &rCOFINS, &rIBS, &rCBS); err != nil {
		t.Fatalf("expected a complete snapshots row in nfe_invoice_tax_rates: %v", err)
	}
	rateColumns := map[string]*decimal.Decimal{
		"icms_rate":   rICMS,
		"pis_rate":    rPIS,
		"cofins_rate": rCOFINS,
		"ibs_rate":    rIBS,
		"cbs_rate":    rCBS,
	}
	for column, value := range rateColumns {
		if value == nil {
			t.Errorf("%s: expected a resolved rate, got nil (the fallback arm is gone)", column)
		}
	}
}

// TestInvoice_TaxSnapshot_DBRoundTrip verifies against a live database that
// CreateInvoiceWithEmission persists the effective-rate/code snapshot:
// the complete tax-rates row (after the fallback arm is gone, the service
// always passes resolved rates) and the resolved code columns, both readable
// through the public getter.
func TestInvoice_TaxSnapshot_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows available for the round trip: %v", err)
	}

	// Minimal FK chain: product → crop → vehicle → departure.
	var productID, cropID, vehicleID, departureID int
	if err := pool.QueryRow(ctx, `INSERT INTO product (name) VALUES (@name) RETURNING id`,
		pgx.NamedArgs{"name": fmt.Sprintf("snapshot-test-product-%d", time.Now().UnixNano())}).Scan(&productID); err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM product WHERE id = @id`, pgx.NamedArgs{"id": productID})
	})

	if err := pool.QueryRow(ctx,
		`INSERT INTO crop (name, product, startDate, farm) VALUES (@name, @product, CURRENT_DATE, @farm) RETURNING id`,
		pgx.NamedArgs{"name": fmt.Sprintf("snapshot-test-crop-%d", time.Now().UnixNano()), "product": productID, "farm": farmID},
	).Scan(&cropID); err != nil {
		t.Fatalf("failed to seed crop: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM crop WHERE id = @id`, pgx.NamedArgs{"id": cropID})
	})

	if err := pool.QueryRow(ctx,
		`INSERT INTO vehicle (plate, name, farm) VALUES (@plate, @name, @farm) RETURNING id`,
		pgx.NamedArgs{"plate": fmt.Sprintf("SPT%d", time.Now().UnixNano()%100000), "name": "tx snapshot test", "farm": farmID},
	).Scan(&vehicleID); err != nil {
		t.Fatalf("failed to seed vehicle: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM vehicle WHERE id = @id`, pgx.NamedArgs{"id": vehicleID})
	})

	if err := pool.QueryRow(ctx,
		`INSERT INTO departure (departureDate, vehicle, crop, grossWeight, tare, netWeight, farm)
		 VALUES (CURRENT_TIMESTAMP, @vehicle, @crop, 110, 10, 100, @farm) RETURNING id`,
		pgx.NamedArgs{"vehicle": vehicleID, "crop": cropID, "farm": farmID},
	).Scan(&departureID); err != nil {
		t.Fatalf("failed to seed departure: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM departure WHERE id = @id`, pgx.NamedArgs{"id": departureID})
	})

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}

	accessKey := fmt.Sprintf("taxsnapshot-dep-%d", time.Now().UnixNano())
	invoiceID, createErr := model.CreateInvoiceWithEmission(
		uint32(departureID), accessKey, 929, 1,
		"5101", "12019000",
		decimal.RequireFromString("100"), decimal.RequireFromString("150.25"),
		decimal.RequireFromString("15025"), decimal.Zero, decimal.Zero,
		1, nil, "", snapshotRates(), snapshotCodes(),
	)
	if createErr != nil {
		t.Fatalf("CreateInvoiceWithEmission failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	})

	// INSERT-side truth: code columns + the complete tax-rates row.
	assertInvoiceSnapshotRow(t, pool, ctx, invoiceID)

	// Getter-side truth (scan parity over the edited SELECT list).
	loaded, getErr := model.GetInvoiceByAccessKey(accessKey)
	if getErr != nil {
		t.Fatalf("GetInvoiceByAccessKey failed: %v", getErr)
	}
	if loaded == nil {
		t.Fatal("expected the invoice to be readable by access key")
	}
	if loaded.TaxRates == nil {
		t.Fatal("expected loaded TaxRates to be non-nil for a complete snapshot")
	}
	wantRates := *snapshotRates()
	gotRates := *loaded.TaxRates
	fields := []struct {
		name      string
		got, want *decimal.Decimal
	}{
		{"ICMS", gotRates.ICMSRate, wantRates.ICMSRate},
		{"PIS", gotRates.PISRate, wantRates.PISRate},
		{"COFINS", gotRates.COFINSRate, wantRates.COFINSRate},
		{"IBS", gotRates.IBSRate, wantRates.IBSRate},
		{"CBS", gotRates.CBSRate, wantRates.CBSRate},
	}
	for _, f := range fields {
		if f.got == nil {
			t.Errorf("%s: expected a resolved rate on the loaded row, got nil", f.name)
			continue
		}
		if !f.got.Equal(*f.want) {
			t.Errorf("%s: got %s, want %s", f.name, f.got, *f.want)
		}
	}
	if loaded.ICMSCST == nil || *loaded.ICMSCST != "00" {
		t.Errorf("expected loaded icms_cst '00', got %v", loaded.ICMSCST)
	}
	if loaded.CClassTrib == nil || *loaded.CClassTrib != "000001" {
		t.Errorf("expected loaded c_class_trib '000001', got %v", loaded.CClassTrib)
	}
}

// TestDetachedInvoice_TaxSnapshot_DBRoundTrip verifies the detached flow's
// snapshot persistence (rates row + resolved code columns), and that the
// legacy shape (nil rates + nil codes) still produces no tax-rates row.
func TestDetachedInvoice_TaxSnapshot_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows available for the round trip: %v", err)
	}
	personID := seedPersonForInvoice(t, pool, ctx, farmID)

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}

	// Snapshot emission.
	accessKey := fmt.Sprintf("taxsnapshot-det-%d", time.Now().UnixNano())
	invoiceID, createErr := model.CreateDetachedInvoice(
		uint32(farmID), uint32(personID), accessKey, 929, 3,
		nil, nil,
		decimal.NewFromInt(100), decimal.Zero, decimal.Zero,
		1, nil, "", snapshotRates(), snapshotCodes(),
		[]nfe_model.DetachedInvoiceItem{},
	)
	if createErr != nil {
		t.Fatalf("CreateDetachedInvoice failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	})

	var dICMSCST, dPISCST, dCOFINSCST, dIBSCST, dCBSCST, dCClassTrib *string
	if err := pool.QueryRow(ctx,
		`SELECT icms_cst, pis_cst, cofins_cst, ibs_cst, cbs_cst, c_class_trib
		 FROM detached_nfe_invoice WHERE id = @id`,
		pgx.NamedArgs{"id": invoiceID},
	).Scan(&dICMSCST, &dPISCST, &dCOFINSCST, &dIBSCST, &dCBSCST, &dCClassTrib); err != nil {
		t.Fatalf("failed to read detached code columns: %v", err)
	}
	for column, value := range map[string]*string{
		"icms_cst":     dICMSCST,
		"pis_cst":      dPISCST,
		"cofins_cst":   dCOFINSCST,
		"ibs_cst":      dIBSCST,
		"cbs_cst":      dCBSCST,
		"c_class_trib": dCClassTrib,
	} {
		if value == nil || *value == "" {
			t.Errorf("%s: expected the resolved code to be persisted, got %v", column, value)
		}
	}

	var drICMS, drPIS, drCOFINS, drIBS, drCBS *decimal.Decimal
	if err := pool.QueryRow(ctx,
		`SELECT icms_rate, pis_rate, cofins_rate, ibs_rate, cbs_rate
		 FROM detached_nfe_tax_rates WHERE invoice_id = @id`,
		pgx.NamedArgs{"id": invoiceID},
	).Scan(&drICMS, &drPIS, &drCOFINS, &drIBS, &drCBS); err != nil {
		t.Fatalf("expected a complete snapshot row in detached_nfe_tax_rates: %v", err)
	}
	for column, value := range map[string]*decimal.Decimal{
		"icms_rate":   drICMS,
		"pis_rate":    drPIS,
		"cofins_rate": drCOFINS,
		"ibs_rate":    drIBS,
		"cbs_rate":    drCBS,
	} {
		if value == nil {
			t.Errorf("%s: expected a resolved rate, got nil (the fallback arm is gone)", column)
		}
	}

	loaded, getErr := model.GetDetachedInvoiceByAccessKey(accessKey)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceByAccessKey failed: %v", getErr)
	}
	if loaded == nil || loaded.TaxRates == nil {
		t.Fatal("expected loaded TaxRates to be non-nil for a complete snapshot")
	}
	if got := *loaded.TaxRates.ICMSRate; !got.Equal(decimal.RequireFromString("0.17")) {
		t.Errorf("expected loaded icms_rate 0.17, got %s", got)
	}

	// Legacy shape: a legacy caller may persist with nil rates (old rows did
	// not always carry a tax-rates row) but the NOT NULL code columns must
	// always bind values (an explicit NULL bypasses the column DEFAULTs).
	// The skip-if-all-nil guard still leaves no tax-rates row for nil rates.
	legacyAccessKey := fmt.Sprintf("taxsnapshot-det-legacy-%d", time.Now().UnixNano())
	legacyID, createErr := model.CreateDetachedInvoice(
		uint32(farmID), uint32(personID), legacyAccessKey, 929, 4,
		nil, nil,
		decimal.NewFromInt(100), decimal.Zero, decimal.Zero,
		1, nil, "", nil, snapshotCodes(),
		[]nfe_model.DetachedInvoiceItem{},
	)
	if createErr != nil {
		t.Fatalf("CreateDetachedInvoice (legacy shape) failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": legacyID})
	})

	var rateRowCount int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM detached_nfe_tax_rates WHERE invoice_id = @id`,
		pgx.NamedArgs{"id": legacyID},
	).Scan(&rateRowCount); err != nil {
		t.Fatalf("failed to count legacy tax-rates rows: %v", err)
	}
	if rateRowCount != 0 {
		t.Errorf("expected no tax-rates row for the legacy shape, got %d", rateRowCount)
	}
}
