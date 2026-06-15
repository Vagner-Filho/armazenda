package armazenda_database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"time"

	zerologadapter "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed schema/*.sql
var schemaFS embed.FS

func execSchemaFile(c *pgx.Conn, filename string) {
	sqlBytes, err := schemaFS.ReadFile(filename)
	if err != nil {
		fmt.Printf("failed to read %s: %v\n", filename, err)
		return
	}
	_, execErr := c.Exec(context.Background(), string(sqlBytes))
	if execErr != nil {
		fmt.Printf("failed to execute %s: %v\n", filename, execErr)
	}
}

func InitDb(c *pgx.Conn) {
	// 1. All tables
	execSchemaFile(c, "schema/tables.sql")

	// 2. All procedures
	execSchemaFile(c, "schema/procedures.sql")

	// 3. All triggers
	execSchemaFile(c, "schema/triggers.sql")
}

// PostMigrationSeeds runs all data seeding that depends on migration-created tables.
// This must be called AFTER RunMigrations completes successfully.
func PostMigrationSeeds(pool *pgxpool.Pool) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("failed to acquire connection for seeding: %v\n", err)
		return
	}
	defer conn.Release()

	c := conn.Conn()
	seedProducts(c)
	seedDefaultHumidityProgression(c)
	seedDefaultPersonConfig(c)
	SeedMunicipios(c)
}

func seedProducts(c *pgx.Conn) {
	var products uint8
	c.QueryRow(context.Background(), "SELECT COUNT(*) FROM product").Scan(&products)
	if products == 0 {
		_, err := c.Exec(context.Background(), "INSERT INTO product (name) VALUES ('Milho'), ('Soja')")
		if err != nil {
			panic(err.Error())
		}
	}
}

func seedDefaultHumidityProgression(c *pgx.Conn) {
	var count int
	err := c.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM humidity_progression WHERE is_system_default = TRUE").Scan(&count)

	if err == nil && count == 0 {
		_, insertErr := c.Exec(context.Background(), `
			INSERT INTO humidity_progression (name, farm_id, is_system_default)
			VALUES ('Padrão do Sistema', NULL, TRUE)
		`)
		if insertErr != nil {
			fmt.Printf("error inserting system default humidity progression: %v\n", insertErr.Error())
			return
		}

		_, insertErr = c.Exec(context.Background(), `
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value)
			SELECT id, 14, 1.7 FROM humidity_progression WHERE is_system_default = TRUE;
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value)
			SELECT id, 16, 1.8 FROM humidity_progression WHERE is_system_default = TRUE;
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value)
			SELECT id, 18, 2.0 FROM humidity_progression WHERE is_system_default = TRUE;
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value)
			SELECT id, 20, 2.2 FROM humidity_progression WHERE is_system_default = TRUE;
		`)
		if insertErr != nil {
			fmt.Printf("error inserting default humidity progression tiers: %v\n", insertErr.Error())
		}
	}
}

func seedDefaultPersonConfig(c *pgx.Conn) {
	var count int
	err := c.QueryRow(context.Background(), "SELECT COUNT(*) FROM default_person_config").Scan(&count)
	if err == nil && count == 0 {
		_, insertErr := c.Exec(context.Background(),
			`INSERT INTO default_person_config (id, humidity_progression_id, entry_soy_discount, entry_corn_discount)
			 SELECT 1, id, 3.5, 5.5 FROM humidity_progression WHERE is_system_default = TRUE`)
		if insertErr != nil {
			fmt.Printf("error inserting default person config: %v\n", insertErr.Error())
		}
	}
}

var dbPool *pgxpool.Pool

func GetDbPool() (*pgxpool.Pool, error) {
	if dbPool == nil {
		dbHost := os.Getenv("DB_HOST")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASS")
		dbName := os.Getenv("DB_NAME")
		dbPort := os.Getenv("DB_PORT")

		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		zlogger := zerolog.New(output).Level(zerolog.ErrorLevel).With().Timestamp().Logger()

		connString := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse connection string")
			fmt.Printf("host | user | psswd | name | port\n%v | %v | %v | %v | %v\n", dbHost, dbUser, dbPass, dbName, dbPort)
			os.Exit(1)

			return nil, errors.New("Falha em conectar ao banco")
		}
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			LogLevel: tracelog.LogLevelError,
			Logger:   zerologadapter.NewLogger(zlogger),
		}

		pool, err := pgxpool.NewWithConfig(context.Background(), config)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			fmt.Printf("host | user | psswd | name | port\n%v | %v | %v | %v | %v\n", dbHost, dbUser, dbPass, dbName, dbPort)
			os.Exit(1)

			return nil, errors.New("Falha em conectar ao banco")
		}

		dbPool = pool
		return dbPool, nil
	}
	return dbPool, nil
}
