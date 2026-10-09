package nfe_model_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"armazenda/model/nfe_model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// TestDetachedInvoiceTransport_JSONRoundTrip verifies the JSONB document
// serializes every transport group and deserializes back with exact values.
func TestDetachedInvoiceTransport_JSONRoundTrip(t *testing.T) {
	pesoL := decimal.RequireFromString("25000.500")
	pesoB := decimal.RequireFromString("26000.250")

	in := nfe_model.DetachedInvoiceTransport{
		Transportadora: &nfe_model.DetachedTransportadora{
			TpDoc:     2,
			DocNumber: "12345678909",
			XNome:     "Carreteiro Fulano",
			IE:        "ISENTO",
			XEnder:    "Rodovia BR-163, Km 12",
			XMun:      "Sorriso",
			UF:        "MT",
		},
		Veiculo: &nfe_model.DetachedVeiculo{
			Placa: "ABC1D23",
			UF:    "MT",
			RNTC:  "12345678901234567890",
		},
		Vol: &nfe_model.DetachedVol{
			Qtd:   1,
			Esp:   "Granel",
			Marca: "",
			PesoL: pesoL,
			PesoB: pesoB,
		},
	}

	payload, marshalErr := json.Marshal(in)
	if marshalErr != nil {
		t.Fatalf("marshal failed: %v", marshalErr)
	}

	var out nfe_model.DetachedInvoiceTransport
	if unmarshalErr := json.Unmarshal(payload, &out); unmarshalErr != nil {
		t.Fatalf("unmarshal failed: %v", unmarshalErr)
	}

	if out.Transportadora == nil || out.Veiculo == nil || out.Vol == nil {
		t.Fatalf("expected all three groups, got %+v", out)
	}

	tr := out.Transportadora
	if tr.TpDoc != 2 || tr.DocNumber != "12345678909" || tr.XNome != "Carreteiro Fulano" ||
		tr.IE != "ISENTO" || tr.XEnder != "Rodovia BR-163, Km 12" || tr.XMun != "Sorriso" || tr.UF != "MT" {
		t.Errorf("transportadora round trip mismatch: %+v", tr)
	}

	va := out.Veiculo
	if va.Placa != "ABC1D23" || va.UF != "MT" || va.RNTC != "12345678901234567890" {
		t.Errorf("veiculo round trip mismatch: %+v", va)
	}

	vol := out.Vol
	if vol.Qtd != 1 || vol.Esp != "Granel" || vol.Marca != "" {
		t.Errorf("vol round trip mismatch: %+v", vol)
	}
	if !vol.PesoL.Equal(pesoL) {
		t.Errorf("expected pesoL %s, got %s", pesoL, vol.PesoL)
	}
	if !vol.PesoB.Equal(pesoB) {
		t.Errorf("expected pesoB %s, got %s", pesoB, vol.PesoB)
	}
}

// TestDetachedInvoiceTransport_OmittedGroupsAreNullOnce verifies that groups
// not informed are omitted from the JSON document (stable shape, no noise).
func TestDetachedInvoiceTransport_OmittedGroupsAreNullOnce(t *testing.T) {
	in := nfe_model.DetachedInvoiceTransport{
		Vol: &nfe_model.DetachedVol{
			Qtd:   1,
			Esp:   "Granel",
			PesoL: decimal.RequireFromString("0"),
			PesoB: decimal.RequireFromString("0"),
		},
	}

	payload, marshalErr := json.Marshal(in)
	if marshalErr != nil {
		t.Fatalf("marshal failed: %v", marshalErr)
	}

	var raw map[string]json.RawMessage
	if unmarshalErr := json.Unmarshal(payload, &raw); unmarshalErr != nil {
		t.Fatalf("unmarshal failed: %v", unmarshalErr)
	}

	if _, present := raw["transportadora"]; present {
		t.Error("transportadora must be omitted when not informed")
	}
	if _, present := raw["veiculo"]; present {
		t.Error("veiculo must be omitted when not informed")
	}
	if _, present := raw["vol"]; !present {
		t.Error("vol must be present in the document")
	}
}

// TestDetachedInvoiceTransport_DBRoundTrip exercises the JSONB persistence
// against a live database. It is deliberately double-gated: (1) the full set
// of DB_* variables must be set, and (2) RUN_DB_TESTS=1 must be set
// explicitly — an inherited dev-shell DB_* environment must never cause an
// implicit write against a running database.
func TestDetachedInvoiceTransport_DBRoundTrip(t *testing.T) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("skip: needs the e2e test database explicitly (DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME) plus RUN_DB_TESTS=1, e.g. `cd test/e2e && bun run db:start && bun run db:seed` then run with the DB env and RUN_DB_TESTS=1")
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

	// Seed a minimal invoice row (1:1 FK target) with cleanup.
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

	var invoiceID int
	accessKey := fmt.Sprintf("dbroundtrip%d", time.Now().UnixNano())
	if err := pool.QueryRow(ctx,
		`INSERT INTO detached_nfe_invoice (farm_id, recipient_id, access_key, serie, number, status, total_value, items_json)
		 VALUES (@farm_id, @recipient_id, @access_key, 929, 1, 'draft', 100, @items)
		 RETURNING id`,
		pgx.NamedArgs{
			"farm_id":      farmID,
			"recipient_id": personID,
			"access_key":   accessKey,
			"items":        []byte(`[]`),
		},
	).Scan(&invoiceID); err != nil {
		t.Fatalf("failed to seed invoice row: %v", err)
	}
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM detached_invoice_transportation WHERE invoice_id = @id`, pgx.NamedArgs{"id": invoiceID})
		pool.Exec(context.Background(),
			`DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	}()

	record := nfe_model.DetachedInvoiceTransport{
		Transportadora: &nfe_model.DetachedTransportadora{
			TpDoc:     2,
			DocNumber: "52998224725",
			XNome:     "Carreteiro DB Round Trip",
			IE:        "ISENTO",
			XEnder:    "Rodovia BR-163, Km 12",
			XMun:      "Sorriso",
			UF:        "MT",
		},
		Veiculo: &nfe_model.DetachedVeiculo{Placa: "ABC1D23", UF: "MT", RNTC: "12345"},
		Vol: &nfe_model.DetachedVol{
			Qtd:   20,
			Esp:   "Big Bag",
			Marca: "BigFarm",
			PesoL: decimal.RequireFromString("15000.500"),
			PesoB: decimal.RequireFromString("15020.250"),
		},
	}

	if createErr := model.CreateDetachedInvoiceTransport(invoiceID, record); createErr != nil {
		t.Fatalf("CreateDetachedInvoiceTransport failed: %v", createErr)
	}

	loaded, getErr := model.GetDetachedInvoiceTransport(invoiceID)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceTransport failed: %v", getErr)
	}
	if loaded == nil {
		t.Fatal("expected a transport document, got nil")
	}

	payload, _ := json.Marshal(record)
	loadedPayload, _ := json.Marshal(*loaded)
	if string(loadedPayload) != string(payload) {
		t.Errorf("round trip mismatch:\n want %s\n got  %s", payload, loadedPayload)
	}

	// An invoice without transport data reads back as nil (no row).
	var missingID int
	if err := pool.QueryRow(ctx,
		`INSERT INTO detached_nfe_invoice (farm_id, recipient_id, access_key, serie, number, status, total_value, items_json)
		 VALUES (@farm_id, @recipient_id, @access_key, 929, 1, 'draft', 100, @items)
		 RETURNING id`,
		pgx.NamedArgs{
			"farm_id":      farmID,
			"recipient_id": personID,
			"access_key":   accessKey + "none",
			"items":        []byte(`[]`),
		},
	).Scan(&missingID); err != nil {
		t.Fatalf("failed to seed second invoice row: %v", err)
	}
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM detached_nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": missingID})
	}()

	empty, getErr := model.GetDetachedInvoiceTransport(missingID)
	if getErr != nil {
		t.Fatalf("GetDetachedInvoiceTransport (absent) failed: %v", getErr)
	}
	if empty != nil {
		t.Error("expected nil transport document when no row exists")
	}
}
