package nfe_router

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"armazenda/model/cfop_model"
	"armazenda/model/farm_product_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// detachedViewPool opens the e2e test database when the deliberate double gate
// is satisfied: the full DB_* set and RUN_DB_TESTS=1. An inherited dev-shell
// DB_* environment must never cause an implicit connection against a running
// database. The pool is shared by the DB-gated tests of this package: the
// models are process-wide singletons, so a second test must not create (and
// close) a different pool behind them.
func dbViewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("skip: needs the e2e test database explicitly (DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME) plus RUN_DB_TESTS=1, e.g. `cd test/e2e && bun run db:start && bun run db:seed` then run with the DB env and RUN_DB_TESTS=1")
	}

	viewPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("pgxpool.New failed: %v", err)
		}
		viewPool = pool
	})
	return viewPool
}

var (
	viewPoolOnce sync.Once
	viewPool     *pgxpool.Pool
)

// TestBuildDetachedPageView_DB is the regression guard for the typed-nil trap
// that made every render of /nfe/emitir return HTTP 500: GetPeopleByFarm
// returns *model_error.ModelError, and checking its nil success value through
// a shared interface-typed error variable reported a phantom non-nil error
// on every successful page load. The success path must return a nil error
// with the loaded view data.
func TestBuildDetachedPageView_DB(t *testing.T) {
	pool := dbViewPool(t)

	if _, err := nfe_model.InitNFeModel(pool); err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}
	if _, err := person_model.InitPersonModel(pool); err != nil {
		t.Fatalf("InitPersonModel failed: %v", err)
	}
	if _, err := farm_product_model.InitFarmProductModel(pool); err != nil {
		t.Fatalf("InitFarmProductModel failed: %v", err)
	}
	if _, err := product_model.InitProductModel(pool); err != nil {
		t.Fatalf("InitProductModel failed: %v", err)
	}
	if _, err := cfop_model.InitCfopModel(pool); err != nil {
		t.Fatalf("InitCfopModel failed: %v", err)
	}

	// Ad-hoc farm with one legal-person contact: GetFarmConfig finds no row
	// (nil config, nil error) and GetPeopleByFarm returns the contact — the
	// all-success path the phantom check used to reject with a page 500.
	suffix := time.Now().UnixNano()

	var farmID uint32
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO farm (inscricao_estadual, uf) VALUES (@inscricao, 'MT') RETURNING id`,
		pgx.NamedArgs{"inscricao": fmt.Sprintf("view%d", suffix)}).Scan(&farmID); err != nil {
		t.Fatalf("failed to seed farm: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM farm WHERE id = @id`, pgx.NamedArgs{"id": farmID})
	})

	var personID uint32
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO person (ie, farm) VALUES ('', @farm) RETURNING id`,
		pgx.NamedArgs{"farm": farmID}).Scan(&personID); err != nil {
		t.Fatalf("failed to seed person: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM legal_person WHERE personid = @id`, pgx.NamedArgs{"id": personID})
		pool.Exec(context.Background(), `DELETE FROM person WHERE id = @id`, pgx.NamedArgs{"id": personID})
	})

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO legal_person (cnpj, personId, companyName) VALUES (@cnpj, @person_id, @name)`,
		pgx.NamedArgs{
			"cnpj":      "11222333000199",
			"person_id": personID,
			"name":      "Destinatário DB View",
		}); err != nil {
		t.Fatalf("failed to seed legal person: %v", err)
	}

	view, viewErr := buildDetachedPageView(farmID, nil)
	if viewErr != nil {
		t.Fatalf("expected no error on the success path, got: %v", viewErr)
	}
	if view.FarmID != farmID {
		t.Errorf("expected FarmID %d, got %d", farmID, view.FarmID)
	}
	// personOptionDisplays skips the id-less "Própria" row, so the seeded
	// contact is the only option in both the recipient and transportadora
	// selects.
	if len(view.People) != 1 || view.People[0].Id != personID || view.People[0].Name != "Destinatário DB View" {
		t.Errorf("expected the seeded contact in People, got %+v", view.People)
	}
	if len(view.TransportPeople) != 1 || view.TransportPeople[0].Id != personID {
		t.Errorf("expected the seeded contact in TransportPeople, got %+v", view.TransportPeople)
	}
	if len(view.Items) == 0 {
		t.Errorf("expected one initial item row, got none")
	}
}
