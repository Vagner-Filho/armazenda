package nfe_model_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// dbRoundTripPool returns a pool against the explicitly-pointed test database.
// Mirrors the double-gating of TestDetachedInvoiceTransport_DBRoundTrip: the
// full DB_* set must be present AND RUN_DB_TESTS=1 must be set explicitly, so
// an inherited dev-shell DB_* environment never causes an implicit write
// against a running database.
func dbRoundTripPool(t *testing.T) (*pgxpool.Pool, context.Context, context.CancelFunc) {
	t.Helper()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("skip: needs the e2e test database explicitly (DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME) plus RUN_DB_TESTS=1, e.g. `cd test/e2e && bun run db:start && bun run db:seed` then run with the DB env and RUN_DB_TESTS=1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		cancel()
		t.Fatalf("pgxpool.New failed: %v", err)
	}
	return pool, ctx, cancel
}

// seedPersonForInvoice inserts a person row usable as the invoice recipient/FK
// target, returning its ID with a cleanup registered on t.
func seedPersonForInvoice(t *testing.T, pool *pgxpool.Pool, ctx context.Context, farmID int) int {
	t.Helper()

	var personID int
	if err := pool.QueryRow(ctx, `INSERT INTO person (farm) VALUES (@farm) RETURNING id`,
		pgx.NamedArgs{"farm": farmID}).Scan(&personID); err != nil {
		t.Fatalf("failed to seed person row: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM person WHERE id = @id`, pgx.NamedArgs{"id": personID})
	})
	return personID
}

// sentAtText reads sent_at as text so idempotency can be asserted exactly (a
// re-stamp would overwrite the sentinel with CURRENT_TIMESTAMP).
func sentAtText(t *testing.T, pool *pgxpool.Pool, ctx context.Context, table string, id int) string {
	t.Helper()

	var sentAt string
	query := fmt.Sprintf(`SELECT COALESCE(sent_at::text, '') FROM %s WHERE id = @id`, table)
	if err := pool.QueryRow(ctx, query, pgx.NamedArgs{"id": id}).Scan(&sentAt); err != nil {
		t.Fatalf("failed to read sent_at from %s: %v", table, err)
	}
	return sentAt
}

// TestDetachedInvoice_ContingencyFieldsAndSentStamp_DBRoundTrip verifies against
// a live database that:
//
//   - CreateDetachedInvoice persists dh_cont/x_just for a contingency emission
//     (tpEmis 6|7) and leaves them NULL/empty for a normal emission, matching
//     the departure flow;
//   - UpdateDetachedInvoiceSent stamps sent_at (NULL before the first attempt)
//     and refuses to overwrite an existing stamp.
func TestDetachedInvoice_ContingencyFieldsAndSentStamp_DBRoundTrip(t *testing.T) {
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

	reason := "Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"
	contAccessKey := fmt.Sprintf("lifecycle-svc-%d", time.Now().UnixNano())

	// The CST columns are NOT NULL and have no value fallback in the INSERT;
	// pass the values a configured form emission posts (prefilled from the
	// farm config defaults).
	icmsCST, pisCST, cofinsCST := "00", "01", "01"
	ibsCST, cbsCST, cClassTrib := "000", "000", "000001"
	overrides := &entity.InvoiceOverrides{
		ICMSCST:    &icmsCST,
		PISCST:     &pisCST,
		COFINSCST:  &cofinsCST,
		IBSCST:     &ibsCST,
		CBSCST:     &cbsCST,
		CClassTrib: &cClassTrib,
	}

	// Contingency emission: dh_cont/x_just must be persisted on the row.
	contID, createErr := model.CreateDetachedInvoice(
		uint32(farmID), uint32(personID), contAccessKey, 929, 1,
		nil, nil,
		decimal.NewFromInt(100), decimal.Zero, decimal.Zero,
		6, time.Now(), reason, nil, overrides, []nfe_model.DetachedInvoiceItem{},
		"1", nil, nil,
	)
	if createErr != nil {
		t.Fatalf("CreateDetachedInvoice (contingency) failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": contID})
	})

	var tpEmis int
	var dhCont *time.Time
	var xJust *string
	var sentAt *time.Time
	if err := pool.QueryRow(ctx,
		`SELECT tp_emis, dh_cont, x_just, sent_at FROM detached_nfe_invoice WHERE id = @id`,
		pgx.NamedArgs{"id": contID},
	).Scan(&tpEmis, &dhCont, &xJust, &sentAt); err != nil {
		t.Fatalf("failed to read contingency invoice: %v", err)
	}
	if tpEmis != 6 {
		t.Errorf("expected tp_emis 6, got %d", tpEmis)
	}
	if dhCont == nil {
		t.Error("expected dh_cont to be persisted for a contingency emission")
	}
	if xJust == nil || *xJust != reason {
		t.Errorf("expected x_just %q, got %v", reason, xJust)
	}
	if sentAt != nil {
		t.Errorf("expected sent_at NULL before the first transmission attempt, got %v", sentAt)
	}

	// Full SELECT + scan parity: read the row back through the same accessor
	// the UI uses, which selects the (edited) column list into
	// scanDetachedInvoice.
	loaded, getErr := model.GetDetachedInvoiceByAccessKey(contAccessKey)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceByAccessKey failed: %v", getErr)
	}
	if loaded == nil {
		t.Fatal("expected the contingency invoice to be readable by access key")
	}
	if loaded.TpEmis != 6 {
		t.Errorf("expected loaded tp_emis 6, got %d", loaded.TpEmis)
	}
	if loaded.DhCont == nil {
		t.Error("expected loaded dh_cont to be set for a contingency emission")
	}
	if loaded.XJust == nil || *loaded.XJust != reason {
		t.Errorf("expected loaded x_just %q, got %v", reason, loaded.XJust)
	}
	if loaded.SentAt != nil {
		t.Errorf("expected loaded sent_at NULL before the first transmission attempt, got %v", loaded.SentAt)
	}

	// First transmission attempt: stamp.
	if err := model.UpdateDetachedInvoiceSent(contID); err != nil {
		t.Fatalf("UpdateDetachedInvoiceSent failed: %v", err)
	}
	if got := sentAtText(t, pool, ctx, "detached_nfe_invoice", contID); got == "" {
		t.Error("expected sent_at to be set after the first stamp")
	}
	loaded, getErr = model.GetDetachedInvoiceByAccessKey(contAccessKey)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceByAccessKey (after stamp) failed: %v", getErr)
	}
	if loaded == nil || loaded.SentAt == nil {
		t.Error("expected loaded sent_at to be set after the first stamp")
	}

	// Idempotency: a re-entered send path must not overwrite the first stamp.
	const sentinel = "2000-01-01 00:00:00"
	if _, err := pool.Exec(ctx,
		`UPDATE detached_nfe_invoice SET sent_at = TIMESTAMP '2000-01-01 00:00:00' WHERE id = @id`,
		pgx.NamedArgs{"id": contID}); err != nil {
		t.Fatalf("failed to set sentinel sent_at: %v", err)
	}
	if err := model.UpdateDetachedInvoiceSent(contID); err != nil {
		t.Fatalf("UpdateDetachedInvoiceSent (second call) failed: %v", err)
	}
	if got := sentAtText(t, pool, ctx, "detached_nfe_invoice", contID); got != sentinel {
		t.Errorf("sent_at guard failed: second stamp overwrote %s with %s", sentinel, got)
	}

	// Normal emission: no contingency fields, dh_cont NULL and x_just empty.
	normalAccessKey := fmt.Sprintf("lifecycle-normal-%d", time.Now().UnixNano())
	normalID, createErr := model.CreateDetachedInvoice(
		uint32(farmID), uint32(personID), normalAccessKey, 929, 2,
		nil, nil,
		decimal.NewFromInt(100), decimal.Zero, decimal.Zero,
		1, nil, "", nil, overrides, []nfe_model.DetachedInvoiceItem{},
		"1", nil, nil,
	)
	if createErr != nil {
		t.Fatalf("CreateDetachedInvoice (normal) failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": normalID})
	})

	var normalDhCont *time.Time
	var normalXJust *string
	if err := pool.QueryRow(ctx,
		`SELECT dh_cont, COALESCE(x_just, '') FROM detached_nfe_invoice WHERE id = @id`,
		pgx.NamedArgs{"id": normalID},
	).Scan(&normalDhCont, &normalXJust); err != nil {
		t.Fatalf("failed to read normal invoice: %v", err)
	}
	if normalDhCont != nil {
		t.Errorf("expected dh_cont NULL for a normal emission, got %v", normalDhCont)
	}
	if normalXJust == nil || *normalXJust != "" {
		t.Errorf("expected x_just empty for a normal emission, got %v", normalXJust)
	}

	// Runtime scan parity for the remaining detached SELECTs (farm list, draft
	// retry, pending retry): each has its own column list feeding
	// scanDetachedInvoice, so a stale column would fail the scan at runtime.
	if _, total, err := model.GetDetachedInvoicesByFarm(uint32(farmID), 1); err != nil || total < 2 {
		t.Errorf("GetDetachedInvoicesByFarm: err=%v total=%d", err, total)
	}
	if drafts, err := model.GetDraftDetachedInvoicesForRetry(); err != nil || len(drafts) == 0 {
		t.Errorf("GetDraftDetachedInvoicesForRetry: err=%v rows=%d", err, len(drafts))
	}
	if _, err := pool.Exec(ctx, `UPDATE detached_nfe_invoice SET status = 'pending' WHERE id = @id`,
		pgx.NamedArgs{"id": normalID}); err != nil {
		t.Fatalf("failed to set pending status: %v", err)
	}
	if pending, err := model.GetPendingDetachedInvoicesForRetry(); err != nil || len(pending) == 0 {
		t.Errorf("GetPendingDetachedInvoicesForRetry: err=%v rows=%d", err, len(pending))
	}
}

// TestInvoice_SentStamp_DBRoundTrip verifies the departure-flow mirror of the
// stamp: UpdateInvoiceSent stamps sent_at and refuses to overwrite an existing
// stamp. The nfe_invoice row is created through CreateInvoiceWithEmission (the
// production writer) over a directly seeded departure.
func TestInvoice_SentStamp_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows available for the round trip: %v", err)
	}

	// departure is the nfe_invoice FK target; it needs a vehicle and a crop of
	// the same farm. The e2e fixture seeds both for farm 1.
	var vehicleID, cropID int
	if err := pool.QueryRow(ctx, `SELECT id FROM vehicle WHERE farm = @farm ORDER BY id LIMIT 1`,
		pgx.NamedArgs{"farm": farmID}).Scan(&vehicleID); err != nil {
		t.Skipf("skip: no vehicle row for farm %d: %v", farmID, err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM crop WHERE farm = @farm ORDER BY id LIMIT 1`,
		pgx.NamedArgs{"farm": farmID}).Scan(&cropID); err != nil {
		t.Skipf("skip: no crop row for farm %d: %v", farmID, err)
	}

	var departureID int
	if err := pool.QueryRow(ctx,
		`INSERT INTO departure (departureDate, vehicle, crop, grossWeight, tare, netWeight, farm)
		 VALUES (CURRENT_TIMESTAMP, @vehicle, @crop, 100, 10, 90, @farm)
		 RETURNING id`,
		pgx.NamedArgs{"vehicle": vehicleID, "crop": cropID, "farm": farmID},
	).Scan(&departureID); err != nil {
		t.Fatalf("failed to seed departure row: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM departure WHERE id = @id`, pgx.NamedArgs{"id": departureID})
	})

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}

	accessKey := fmt.Sprintf("lifecycle-nfe-%d", time.Now().UnixNano())
	icmsCST, pisCST, cofinsCST := "00", "01", "01"
	ibsCST, cbsCST, cClassTrib := "000", "000", "000001"
	overrides := &entity.InvoiceOverrides{
		ICMSCST:    &icmsCST,
		PISCST:     &pisCST,
		COFINSCST:  &cofinsCST,
		IBSCST:     &ibsCST,
		CBSCST:     &cbsCST,
		CClassTrib: &cClassTrib,
	}
	invoiceID, createErr := model.CreateInvoiceWithEmission(
		uint32(departureID), accessKey, 929, 1, "5101", "12019000",
		decimal.NewFromInt(90), decimal.NewFromInt(10), decimal.NewFromInt(900),
		decimal.Zero, decimal.Zero,
		1, nil, "", nil, overrides, nil,
	)
	if createErr != nil {
		t.Fatalf("CreateInvoiceWithEmission failed: %v", createErr)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	})

	if got := sentAtText(t, pool, ctx, "nfe_invoice", invoiceID); got != "" {
		t.Errorf("expected sent_at NULL before the first transmission attempt, got %s", got)
	}

	// Full SELECT + scan parity through the accessor used by the UI.
	loaded, getErr := model.GetInvoiceByAccessKey(accessKey)
	if getErr != nil {
		t.Fatalf("GetInvoiceByAccessKey failed: %v", getErr)
	}
	if loaded == nil {
		t.Fatal("expected the invoice to be readable by access key")
	}
	if loaded.SentAt != nil {
		t.Errorf("expected loaded sent_at NULL before the first transmission attempt, got %v", loaded.SentAt)
	}

	// Runtime scan parity for the remaining nfe_invoice SELECTs (departure,
	// farm list, draft retry, pending retry): each has its own column list
	// feeding scanInvoice, so a stale column would fail the scan at runtime.
	if byDeparture, err := model.GetInvoiceByDeparture(uint32(departureID)); err != nil || byDeparture == nil {
		t.Errorf("GetInvoiceByDeparture: err=%v invoice=%v", err, byDeparture)
	}
	if _, total, err := model.GetInvoicesByFarm(uint32(farmID), 1); err != nil || total < 1 {
		t.Errorf("GetInvoicesByFarm: err=%v total=%d", err, total)
	}
	if drafts, err := model.GetDraftInvoicesForRetry(); err != nil || len(drafts) == 0 {
		t.Errorf("GetDraftInvoicesForRetry: err=%v rows=%d", err, len(drafts))
	}
	if _, err := pool.Exec(ctx, `UPDATE nfe_invoice SET status = 'pending' WHERE id = @id`,
		pgx.NamedArgs{"id": invoiceID}); err != nil {
		t.Fatalf("failed to set pending status: %v", err)
	}
	if pending, err := model.GetPendingInvoicesForRetry(); err != nil || len(pending) == 0 {
		t.Errorf("GetPendingInvoicesForRetry: err=%v rows=%d", err, len(pending))
	}

	if err := model.UpdateInvoiceSent(invoiceID); err != nil {
		t.Fatalf("UpdateInvoiceSent failed: %v", err)
	}
	if got := sentAtText(t, pool, ctx, "nfe_invoice", invoiceID); got == "" {
		t.Error("expected sent_at to be set after the first stamp")
	}

	const sentinel = "2000-01-01 00:00:00"
	if _, err := pool.Exec(ctx,
		`UPDATE nfe_invoice SET sent_at = TIMESTAMP '2000-01-01 00:00:00' WHERE id = @id`,
		pgx.NamedArgs{"id": invoiceID}); err != nil {
		t.Fatalf("failed to set sentinel sent_at: %v", err)
	}
	if err := model.UpdateInvoiceSent(invoiceID); err != nil {
		t.Fatalf("UpdateInvoiceSent (second call) failed: %v", err)
	}
	if got := sentAtText(t, pool, ctx, "nfe_invoice", invoiceID); got != sentinel {
		t.Errorf("sent_at guard failed: second stamp overwrote %s with %s", sentinel, got)
	}
}
