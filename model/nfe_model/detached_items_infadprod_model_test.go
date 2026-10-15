package nfe_model_test

import (
	"context"
	"encoding/json"
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

// TestDetachedInvoiceItem_InfAdProdJSON pins the items_json contract of the
// per-item additional information (inf_ad_prod): the key round-trips when
// informed, is omitted when empty, and a document written before the key
// existed (pre-existing rows) parses with an empty value.
func TestDetachedInvoiceItem_InfAdProdJSON(t *testing.T) {
	t.Run("key round-trips", func(t *testing.T) {
		item := nfe_model.DetachedInvoiceItem{
			ProductName: "Soja",
			NCM:         "12019000",
			CFOP:        "5101",
			Unit:        "KG",
			Quantity:    decimal.RequireFromString("10000"),
			GrossWeight: decimal.RequireFromString("10200"),
			UnitPrice:   decimal.RequireFromString("150.25"),
			TotalValue:  decimal.RequireFromString("1502500"),
			InfAdProd:   "SERIE ABC123",
		}

		payload, marshalErr := json.Marshal(item)
		if marshalErr != nil {
			t.Fatalf("marshal failed: %v", marshalErr)
		}
		if !containsKey(payload, "inf_ad_prod") {
			t.Fatalf("expected inf_ad_prod in the payload, got %s", payload)
		}

		var out nfe_model.DetachedInvoiceItem
		if unmarshalErr := json.Unmarshal(payload, &out); unmarshalErr != nil {
			t.Fatalf("unmarshal failed: %v", unmarshalErr)
		}
		if out.InfAdProd != "SERIE ABC123" {
			t.Errorf("InfAdProd = %q, want %q", out.InfAdProd, "SERIE ABC123")
		}
	})

	t.Run("empty value omits the key", func(t *testing.T) {
		payload, marshalErr := json.Marshal(nfe_model.DetachedInvoiceItem{
			ProductName: "Soja",
			NCM:         "12019000",
			CFOP:        "5101",
			Unit:        "KG",
			Quantity:    decimal.RequireFromString("10000"),
			GrossWeight: decimal.RequireFromString("10200"),
			UnitPrice:   decimal.RequireFromString("150.25"),
			TotalValue:  decimal.RequireFromString("1502500"),
		})
		if marshalErr != nil {
			t.Fatalf("marshal failed: %v", marshalErr)
		}
		if containsKey(payload, "inf_ad_prod") {
			t.Errorf("expected inf_ad_prod to be omitted when empty, got %s", payload)
		}
	})

	t.Run("legacy document without the key parses empty", func(t *testing.T) {
		// Shape of a row persisted before the inf_ad_prod key existed
		// (no inf_ad_prod key at all).
		legacy := []byte(`{"product_name":"Soja","ncm":"12019000","cest":"0100101","cfop":"5101",` +
			`"unit":"KG","quantity":"10000","gross_weight":"10200","unit_price":"150.25",` +
			`"discount":"0","total_value":"1502500"}`)

		var out nfe_model.DetachedInvoiceItem
		if unmarshalErr := json.Unmarshal(legacy, &out); unmarshalErr != nil {
			t.Fatalf("legacy document must parse without the new key: %v", unmarshalErr)
		}
		if out.InfAdProd != "" {
			t.Errorf("legacy row InfAdProd = %q, want empty", out.InfAdProd)
		}
	})
}

// containsKey reports whether a JSON object payload carries the given key.
func containsKey(payload []byte, key string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false
	}
	_, present := raw[key]
	return present
}

// TestDetachedInvoiceItems_InfAdProdDBRoundTrip exercises the inf_ad_prod
// persistence against a live database: CreateDetachedInvoice →
// GetDetachedInvoiceByAccessKey reproduces the value, and a row seeded with a
// legacy items_json document (no key) reads back empty. Like the transport
// DB test it is deliberately double-gated: (1) the full set of DB_* variables
// must be set, and (2) RUN_DB_TESTS=1 must be set explicitly — an inherited
// dev-shell DB_* environment must never cause an implicit write against a
// running database.
func TestDetachedInvoiceItems_InfAdProdDBRoundTrip(t *testing.T) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("skip: needs the e2e test database explicitly (DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME) plus RUN_DB_TESTS=1, e.g. `cd test/e2e && bun run db:start` then run the app once so migrations apply")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New failed: %v", err)
	}
	defer pool.Close()

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}

	var farmID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows available for the round trip: %v", err)
	}

	// person has no seeded rows on a fresh database: create one for the FK.
	var personID int
	if err := pool.QueryRow(ctx, `INSERT INTO person (farm) VALUES (@farm) RETURNING id`,
		pgx.NamedArgs{"farm": farmID}).Scan(&personID); err != nil {
		t.Fatalf("failed to seed person row: %v", err)
	}
	defer func() {
		pool.Exec(context.Background(), `DELETE FROM person WHERE id = @id`, pgx.NamedArgs{"id": personID})
	}()

	accessKey := fmt.Sprintf("infadprod%d", time.Now().UnixNano())
	items := []nfe_model.DetachedInvoiceItem{
		{
			ProductName: "Soja",
			NCM:         "12019000",
			CFOP:        "5101",
			Unit:        "KG",
			Quantity:    decimal.RequireFromString("10"),
			GrossWeight: decimal.RequireFromString("10"),
			UnitPrice:   decimal.RequireFromString("150"),
			InfAdProd:   "SERIE ABC123",
			TotalValue:  decimal.RequireFromString("1500"),
		},
	}

	// The production emission always persists the resolved tax codes; the
	// IBS/CBS code columns are NOT NULL, so the fixture mirrors that call.
	ibsCST, cbsCST, cClassTrib := "000", "000", "000001"
	codes := &entity.InvoiceOverrides{IBSCST: &ibsCST, CBSCST: &cbsCST, CClassTrib: &cClassTrib}

	invoiceID, createErr := model.CreateDetachedInvoice(
		uint32(farmID), uint32(personID), accessKey, 929, 1,
		nil, nil, decimal.RequireFromString("1500"), decimal.Zero, decimal.Zero,
		1, nil, "", nil, codes, items, "1", nil, nil,
	)
	if createErr != nil {
		t.Fatalf("CreateDetachedInvoice failed: %v", createErr)
	}
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	}()

	loaded, getErr := model.GetDetachedInvoiceByAccessKey(accessKey)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceByAccessKey failed: %v", getErr)
	}
	if loaded == nil || len(loaded.Items) != 1 {
		t.Fatalf("expected the invoice with 1 item, got %+v", loaded)
	}
	if loaded.Items[0].InfAdProd != "SERIE ABC123" {
		t.Errorf("persisted InfAdProd = %q, want %q", loaded.Items[0].InfAdProd, "SERIE ABC123")
	}

	// A legacy row (items_json without the key) reads back empty, no error.
	legacyKey := accessKey + "legacy"
	var legacyID int
	if err := pool.QueryRow(ctx,
		`INSERT INTO detached_nfe_invoice (farm_id, recipient_id, access_key, serie, number, status, total_value, items_json)
		 VALUES (@farm_id, @recipient_id, @access_key, 929, 2, 'draft', 100, @items)
		 RETURNING id`,
		pgx.NamedArgs{
			"farm_id":      farmID,
			"recipient_id": personID,
			"access_key":   legacyKey,
			"items": []byte(`[{"product_name":"Soja","ncm":"12019000","cfop":"5101","unit":"KG",` +
				`"quantity":"10","gross_weight":"10","unit_price":"150","discount":"0","total_value":"1500"}]`),
		},
	).Scan(&legacyID); err != nil {
		t.Fatalf("failed to seed legacy invoice row: %v", err)
	}
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": legacyID})
	}()

	legacy, getErr := model.GetDetachedInvoiceByAccessKey(legacyKey)
	if getErr != nil {
		t.Fatalf("legacy row must read back without error: %v", getErr)
	}
	if legacy == nil || len(legacy.Items) != 1 {
		t.Fatalf("expected the legacy invoice with 1 item, got %+v", legacy)
	}
	if legacy.Items[0].InfAdProd != "" {
		t.Errorf("legacy row InfAdProd = %q, want empty", legacy.Items[0].InfAdProd)
	}
}
