package armazenda_database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func handleStmtExec(c *pgx.Conn, stmt *pgconn.StatementDescription, err error) {
	if err != nil {
		fmt.Printf("prepare stmt err %v\n", err.Error())
	}

	_, execErr := c.Exec(context.Background(), stmt.SQL)

	if execErr != nil {
		fmt.Printf("prepare stmt err %v\n", execErr.Error())
	}
}

func initProduct(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init product table", `
	CREATE TABLE IF NOT EXISTS product (
    		id SMALLINT PRIMARY KEY,
    		name VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initField(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init field table", `
	CREATE TABLE IF NOT EXISTS field (
    		id SMALLINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initCrop(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init crop table", `
	CREATE TABLE IF NOT EXISTS crop (
		id SMALLINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		startDate DATE NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initVehicle(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init vehicle table", `
	CREATE TABLE IF NOT EXISTS vehicle (
		plate VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry table", `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		product SMALLINT NOT NULL,
		field SMALLINT NOT NULL,
		crop SMALLINT NOT NULL,
		vehicle VARCHAR(255) NOT NULL,
		grossWeight DOUBLE PRECISION,
		tare DOUBLE PRECISION,
		netWeight DOUBLE PRECISION NOT NULL,
		humidity DOUBLE PRECISION,
		arrivalDate DATE NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id),
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (field) REFERENCES field(id),
		FOREIGN KEY (crop) REFERENCES crop(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initDeparture(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure table", `
	CREATE TABLE IF NOT EXISTS departure (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		departureDate DATE NOT NULL,
		product SMALLINT NOT NULl,
		vehicle VARCHAR(255),
		crop SMALLINT NOT NULL,
		weight DOUBLE PRECISION NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id),
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (crop) REFERENCES crop(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initBuyer(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init buyer table", `
	CREATE TABLE IF NOT EXISTS buyer (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		ie VARCHAR(255) UNIQUE
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initBuyerPerson(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init buyerPerson table", `
	CREATE TABLE IF NOT EXISTS buyerPerson (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(255) NOT NULL,
		cpf VARCHAR(255) UNIQUE NOT NULL,
		buyerId INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (buyerId) REFERENCES buyer(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initBuyerCompany(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init buyerCompany table", `
	CREATE TABLE IF NOT EXISTS buyerCompany (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		cnpj VARCHAR(255) UNIQUE NOT NULL,
		buyerId INTEGER UNIQUE NOT NULL,
		companyName VARCHAR(255) NOT NULL,
		fantasyName VARCHAR(255),
		FOREIGN KEY (buyerId) REFERENCES buyer(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initDepartureBuyer(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure_buyer table", `
	CREATE TABLE IF NOT EXISTS departureBuyer (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		departureId INTEGER UNIQUE NOT NULL,
		buyerId VARCHAR(255) NOT NULL,
		FOREIGN KEY (buyerId) REFERENCES buyer(id),
		FOREIGN KEY (departureId) REFERENCES departure(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initAddrress(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init address table", `
	CREATE TABLE IF NOT EXISTS address (
		id SMALLINT PRIMARY KEY,
		street VARCHAR(255) NOT NULL,
		cep VARCHAR(255) NOT NULL,
		number INTEGER,
		neighborhood VARCHAR(255) NOT NULL,
		city VARCHAR(255) NOT NULL,
		state VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initAddrressComplement(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init addressComplement table", `
	CREATE TABLE IF NOT EXISTS addressComplement (
		id SMALLINT PRIMARY KEY,
		complement TEXT NOT NULL,
		addressId SMALLINT UNIQUE NOT NULL,
		FOREIGN KEY (addressId) REFERENCES address(id)
	);
	`)

	handleStmtExec(c, stmt, err)
}

func initContact(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init contact table", `
	CREATE TABLE IF NOT EXISTS contact (
		id SMALLINT PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		phoneNumber VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err)
}

func InitDb(c *pgx.Conn) {
	initProduct(c)
	initCrop(c)
	initVehicle(c)
	initField(c)
	initEntry(c)
	initDeparture(c)
	initDeparture(c)
	initBuyer(c)
	initDepartureBuyer(c)
	initBuyerPerson(c)
	initBuyerCompany(c)
	initContact(c)
	initAddrress(c)
	initAddrressComplement(c)
}
