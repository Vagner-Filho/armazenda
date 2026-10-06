package cfop_model_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// migrationPath is read directly (no DB) so the seed list integrity is checked
// offline, as the SDD prescribes.
const migrationPath = "../armazenda_database/migrations/000025_cfop_catalog.sql"

// expectedSeedCount is the number of distinct codes extracted from the
// Contabilizei table plus the system-assumed extras. Update it only when the
// catalog is intentionally extended.
const expectedSeedCount = 167

var seedTupleRe = regexp.MustCompile(`\('(\d{4})', '((?:[^']|'')*)', '((?:[^']|'')*)'\)`)

func readMigration(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(content)
}

func parseSeedRows(t *testing.T, migration string) map[string][2]string {
	t.Helper()
	rows := map[string][2]string{}
	for _, match := range seedTupleRe.FindAllStringSubmatch(migration, -1) {
		code := match[1]
		description := strings.ReplaceAll(match[2], "''", "'")
		origin := strings.ReplaceAll(match[3], "''", "'")
		if _, exists := rows[code]; exists {
			t.Errorf("duplicate seed code %s", code)
		}
		rows[code] = [2]string{description, origin}
	}
	return rows
}

func TestCfopSeed_StructuralContracts(t *testing.T) {
	migration := readMigration(t)
	for _, table := range []string{"cfop", "farm_cfop", "farm_cfop_use"} {
		if !strings.Contains(migration, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Errorf("migration must create %q idempotently", table)
		}
	}
	if !strings.Contains(migration, "ON CONFLICT (code) DO NOTHING") {
		t.Error("seed must be idempotent (ON CONFLICT (code) DO NOTHING)")
	}
	if strings.Contains(migration, "ALTER TABLE") {
		t.Error("migration must not alter existing NF-e tables")
	}
}

func TestCfopSeed_AllCodesAndCount(t *testing.T) {
	rows := parseSeedRows(t, readMigration(t))
	if len(rows) != expectedSeedCount {
		t.Fatalf("expected %d seeded codes, got %d", expectedSeedCount, len(rows))
	}
}

func TestCfopSeed_OriginDestinationMatchesFirstDigit(t *testing.T) {
	rows := parseSeedRows(t, readMigration(t))
	derivations := map[byte]string{
		'1': "Mesmo estado", '5': "Mesmo estado",
		'2': "Outro estado", '6': "Outro estado",
		'3': "Exterior", '7': "Exterior",
	}
	for code, row := range rows {
		expected, ok := derivations[code[0]]
		if !ok {
			t.Errorf("code %s has an invalid first digit", code)
			continue
		}
		if row[1] != expected {
			t.Errorf("code %s origin_destination = %q, want %q", code, row[1], expected)
		}
		if row[0] == "" {
			t.Errorf("code %s has an empty description", code)
		}
	}
}

func TestCfopSeed_SystemAssumedCodesPresent(t *testing.T) {
	rows := parseSeedRows(t, readMigration(t))
	required := []string{
		// Agriculture defaults.
		"5101", "5102", "5901", "6202",
		// Every NaturezaOpForCFOP key (pkg/nfe/defaults).
		"1101", "7101", "7102", "1901", "7901",
		"1202", "2202", "5202",
		"1103", "5103", "7103", "1104", "5104", "7104",
	}
	for _, code := range required {
		if _, ok := rows[code]; !ok {
			t.Errorf("required system CFOP %s missing from the seed", code)
		}
	}
}

func TestCfopSeed_SourceSpotChecks(t *testing.T) {
	rows := parseSeedRows(t, readMigration(t))
	if got := rows["1102"][0]; got != "Compra para comercialização" {
		t.Errorf("1102 description = %q", got)
	}
	if got := rows["5102"][0]; got != "Venda de mercadoria adquirida ou recebida de terceiros" {
		t.Errorf("5102 description = %q", got)
	}
	// The page lists 6905 as "Mesmo estado"; the first digit (6, exit to
	// another state) is the source of truth per the SDD.
	if got := rows["6905"][1]; got != "Outro estado" {
		t.Errorf("6905 origin_destination = %q, want Outro estado", got)
	}
	if _, ok := rows["0000"]; ok {
		t.Error("seed must not contain the invalid code 0000")
	}
}
