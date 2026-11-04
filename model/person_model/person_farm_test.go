package person_model_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"armazenda/model/person_model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestPersonBelongsToFarm verifies the tenancy guard used before a contact is
// resolved into an NF-e group. It is deliberately double-gated: (1) the full
// set of DB_* variables must be set, and (2) RUN_DB_TESTS=1 must be set
// explicitly — an inherited dev-shell DB_* environment must never cause an
// implicit write against a running database.
func TestPersonBelongsToFarm(t *testing.T) {
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

	model, err := person_model.InitPersonModel(pool)
	if err != nil {
		t.Fatalf("InitPersonModel failed: %v", err)
	}

	// Two farms: the contact belongs to one of them only.
	suffix := time.Now().UnixNano()
	createFarm := func(label string) int {
		var farmID int
		inscricao := fmt.Sprintf("db%d%s", suffix, label)
		if err := pool.QueryRow(ctx,
			`INSERT INTO farm (inscricao_estadual, uf) VALUES (@inscricao, 'MT') RETURNING id`,
			pgx.NamedArgs{"inscricao": inscricao}).Scan(&farmID); err != nil {
			t.Fatalf("failed to seed farm %s: %v", label, err)
		}
		return farmID
	}

	farmID := createFarm("a")
	otherFarmID := createFarm("b")
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM farm WHERE id = @id`, pgx.NamedArgs{"id": farmID})
		pool.Exec(context.Background(),
			`DELETE FROM farm WHERE id = @id`, pgx.NamedArgs{"id": otherFarmID})
	}()

	var personID int
	if err := pool.QueryRow(ctx,
		`INSERT INTO person (ie, farm) VALUES ('', @farm) RETURNING id`,
		pgx.NamedArgs{"farm": farmID}).Scan(&personID); err != nil {
		t.Fatalf("failed to seed person row: %v", err)
	}
	defer func() {
		pool.Exec(context.Background(),
			`DELETE FROM person WHERE id = @id`, pgx.NamedArgs{"id": personID})
	}()

	belongs, modelErr := model.PersonBelongsToFarm(uint32(personID), uint32(farmID))
	if modelErr != nil {
		t.Fatalf("PersonBelongsToFarm failed: %v", modelErr.Error())
	}
	if !belongs {
		t.Error("a contact of the farm must belong to it")
	}

	belongs, modelErr = model.PersonBelongsToFarm(uint32(personID), uint32(otherFarmID))
	if modelErr != nil {
		t.Fatalf("PersonBelongsToFarm (other farm) failed: %v", modelErr.Error())
	}
	if belongs {
		t.Error("a contact of another farm must not belong to this farm")
	}

	belongs, modelErr = model.PersonBelongsToFarm(uint32(personID+1_000_000), uint32(farmID))
	if modelErr != nil {
		t.Fatalf("PersonBelongsToFarm (missing id) failed: %v", modelErr.Error())
	}
	if belongs {
		t.Error("a missing contact must not belong to the farm")
	}
}
