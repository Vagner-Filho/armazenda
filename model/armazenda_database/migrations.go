package armazenda_database

import (
	"context"
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version     int
	name        string
	description string
}

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// Ensure schema_migrations table exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			applied_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Load already-applied versions
	applied := make(map[int]bool)
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[v] = true
	}
	rows.Close()

	// Read migration files from embedded FS
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".sql")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue // skip files without version prefix
		}

		version, parseErr := strconv.Atoi(parts[0])
		if parseErr != nil {
			continue // skip files with non-numeric prefix
		}

		description := strings.ReplaceAll(parts[1], "_", " ")
		migrations = append(migrations, migration{
			version:     version,
			name:        entry.Name(),
			description: description,
		})
	}

	// Sort by version ascending
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Apply pending migrations
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		sqlBytes, readErr := migrationsFS.ReadFile(path.Join("migrations", m.name))
		if readErr != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.name, readErr)
		}

		tx, beginErr := pool.Begin(ctx)
		if beginErr != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", m.version, beginErr)
		}

		_, execErr := tx.Exec(ctx, string(sqlBytes))
		if execErr != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %d: %w", m.version, execErr)
		}

		_, recordErr := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, description) VALUES ($1, $2)`,
			m.version, m.description)
		if recordErr != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %d: %w", m.version, recordErr)
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.version, commitErr)
		}

		fmt.Printf("applied migration %06d: %s\n", m.version, m.description)
	}

	return nil
}
