package nfe_service

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/person_model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// dbContactPool opens the e2e test database when the deliberate double gate is
// satisfied: the full DB_* set and RUN_DB_TESTS=1. An inherited dev-shell DB_*
// environment must never cause an implicit write against a running database.
//
// The pool is shared by every DB-gated test of this package: the person model
// is a process-wide singleton, so a second test must not create (and close) a
// different pool behind it. The process exit releases it.
func dbContactPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("skip: needs the e2e test database explicitly (DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME) plus RUN_DB_TESTS=1, e.g. `cd test/e2e && bun run db:start && bun run db:seed` then run with the DB env and RUN_DB_TESTS=1")
	}

	contactPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("pgxpool.New failed: %v", err)
		}
		contactPool = pool
	})
	return contactPool
}

var (
	contactPoolOnce sync.Once
	contactPool     *pgxpool.Pool
)

// seedContactFarm creates a farm with a contact in it and returns both ids.
// The contact is a legal person, which is enough for the transport mapping.
func seedContactFarm(t *testing.T, pool *pgxpool.Pool, label string) (uint32, uint32) {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	var farmID uint32
	if err := pool.QueryRow(ctx,
		`INSERT INTO farm (inscricao_estadual, uf) VALUES (@inscricao, 'MT') RETURNING id`,
		pgx.NamedArgs{"inscricao": fmt.Sprintf("db%d%s", suffix, label)}).Scan(&farmID); err != nil {
		t.Fatalf("failed to seed farm %s: %v", label, err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM farm WHERE id = @id`, pgx.NamedArgs{"id": farmID})
	})

	var personID uint32
	if err := pool.QueryRow(ctx,
		`INSERT INTO person (ie, farm) VALUES ('', @farm) RETURNING id`,
		pgx.NamedArgs{"farm": farmID}).Scan(&personID); err != nil {
		t.Fatalf("failed to seed person: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM legal_person WHERE personId = @id`, pgx.NamedArgs{"id": personID})
		pool.Exec(context.Background(), `DELETE FROM person WHERE id = @id`, pgx.NamedArgs{"id": personID})
	})

	if _, err := pool.Exec(ctx,
		`INSERT INTO legal_person (cnpj, personId, companyName) VALUES (@cnpj, @person_id, @name)`,
		pgx.NamedArgs{
			"cnpj":      "11222333000181",
			"person_id": personID,
			"name":      "Transportadora DB " + label,
		}); err != nil {
		t.Fatalf("failed to seed legal person: %v", err)
	}

	return farmID, personID
}

// TestResolveTransportContact_DB exercises the read-only contact resolution
// against the e2e database: a contact of another farm is refused with the
// warning toast and no write, a contact of the farm is mapped into the
// Transportadora group, and a missing contact is refused the same way.
func TestResolveTransportContact_DB(t *testing.T) {
	pool := dbContactPool(t)

	if _, err := person_model.InitPersonModel(pool); err != nil {
		t.Fatalf("InitPersonModel failed: %v", err)
	}

	farmID, personID := seedContactFarm(t, pool, "a")
	otherFarmID, _ := seedContactFarm(t, pool, "b")

	// Diagnostic: the resolution depends on the full-person lookup working.
	if _, err := person_model.GetPersonModel().GetFullPersonById(personID); err != nil {
		t.Fatalf("GetFullPersonById failed: %s", err.Error())
	}

	svc := NewNFeService()

	// No contact: the parsed input is returned unchanged.
	in := DetachedInvoiceInput{FarmID: farmID, Transport: DetachedTransportInput{}}
	got, toast := svc.resolveTransportContact(in)
	if toast.Type != entity_public.SuccessToast {
		t.Fatalf("no contact must not warn, got %+v", toast)
	}
	if got.PersonID != nil {
		t.Error("no contact must stay nil")
	}

	// Contact of the farm: resolved into the Transportadora group.
	in.Transport.PersonID = &personID
	got, toast = svc.resolveTransportContact(in)
	if toast.Type != entity_public.SuccessToast {
		t.Fatalf("own contact must resolve, got %+v", toast)
	}
	if got.Nome == nil || *got.Nome != "Transportadora DB a" {
		t.Errorf("contact name must be mapped, got %v", got.Nome)
	}
	if got.Document == nil || *got.Document != "11222333000181" {
		t.Errorf("contact document must be mapped, got %v", got.Document)
	}

	// Contact of another farm: rejected with the pt-br warning.
	in.FarmID = otherFarmID
	got, toast = svc.resolveTransportContact(in)
	if toast.Type != entity_public.WarningToast {
		t.Fatalf("cross-farm contact must warn, got %+v", toast)
	}
	if toast.Message != "Transportadora inválida" ||
		toast.Hint != "A transportadora selecionada não pertence à sua fazenda" {
		t.Errorf("unexpected cross-farm toast: %+v", toast)
	}
	if got.Nome != nil || got.Document != nil {
		t.Error("a rejected contact must not be mapped into the transport group")
	}

	// No person row is created or removed by the resolution.
	var personCount int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM person WHERE farm = @farm`, pgx.NamedArgs{"farm": farmID}).Scan(&personCount); err != nil {
		t.Fatalf("failed to count persons: %v", err)
	}
	if personCount != 1 {
		t.Errorf("expected exactly the seeded contact, got %d", personCount)
	}

	// Missing contact: rejected with the same warning.
	missing := personID + 1_000_000
	in = DetachedInvoiceInput{FarmID: farmID, Transport: DetachedTransportInput{PersonID: &missing}}
	if _, toast = svc.resolveTransportContact(in); toast.Type != entity_public.WarningToast {
		t.Fatalf("a missing contact must warn, got %+v", toast)
	}
}

// TestResolveOrCreateRecipient_RejectsCrossFarm_DB exercises the recipient
// tenancy guard: a recipient of another farm is refused before any lookup or
// write, and a recipient of the farm resolves.
func TestResolveOrCreateRecipient_RejectsCrossFarm_DB(t *testing.T) {
	pool := dbContactPool(t)

	if _, err := person_model.InitPersonModel(pool); err != nil {
		t.Fatalf("InitPersonModel failed: %v", err)
	}

	farmID, personID := seedContactFarm(t, pool, "c")
	otherFarmID, _ := seedContactFarm(t, pool, "d")

	svc := NewNFeService()
	recipient := DetachedRecipient{PersonID: &personID}

	// Same farm: resolved (dry-run keeps it read-only).
	resolvedID, resolved, toast := svc.resolveOrCreateRecipient(recipient, farmID, nil, true)
	if toast.Type != entity_public.SuccessToast {
		t.Fatalf("own recipient must resolve, got %+v", toast)
	}
	if resolvedID != personID || resolved.XNome != "Transportadora DB c" {
		t.Errorf("unexpected resolved recipient: id=%d name=%q", resolvedID, resolved.XNome)
	}

	// Another farm: warning before any write.
	before := countPersons(t, pool, otherFarmID)
	_, _, toast = svc.resolveOrCreateRecipient(recipient, otherFarmID, nil, true)
	if toast.Type != entity_public.WarningToast {
		t.Fatalf("cross-farm recipient must warn, got %+v", toast)
	}
	if toast.Message != "Destinatário inválido" ||
		toast.Hint != "O destinatário selecionado não pertence à sua fazenda" {
		t.Errorf("unexpected cross-farm toast: %+v", toast)
	}
	if after := countPersons(t, pool, otherFarmID); after != before {
		t.Errorf("the rejection must not write a person row: before=%d after=%d", before, after)
	}
}

func countPersons(t *testing.T, pool *pgxpool.Pool, farmID uint32) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM person WHERE farm = @farm`, pgx.NamedArgs{"farm": farmID}).Scan(&count); err != nil {
		t.Fatalf("failed to count persons: %v", err)
	}
	return count
}
