package armazenda_database

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func handleStmtExec(c *pgx.Conn, stmt *pgconn.StatementDescription, err error, operationName string) {
	if c == nil {
		fmt.Printf("\nmissing db connection: %v\n", stmt.Name)
		return
	}

	if stmt == nil {
		fmt.Printf("\nmissing db statement for: %v\n", operationName)
		if err != nil {
			fmt.Printf("prepare stmt err %v\n", err.Error())
		}
		return
	}

	if err != nil {
		fmt.Printf("stmt name: %v\n", stmt.Name)
		fmt.Printf("prepare stmt err %v\n", err.Error())
		return
	}

	_, execErr := c.Exec(context.Background(), stmt.SQL)

	if execErr != nil {
		fmt.Printf("stmt name: %v\n", stmt.Name)
		fmt.Printf("exec stmt err %v\n", execErr.Error())
	}
}

func initProduct(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init product table", `
	CREATE TABLE IF NOT EXISTS product (
    		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    		name VARCHAR(255) NOT NULL
	);
	`)

	if err == nil {
		var products uint8
		c.QueryRow(context.Background(), "SELECT COUNT(*) FROM product").Scan(&products)
		if products == 0 {
			_, insertProductErr := c.Exec(context.Background(), "INSERT INTO product (name) VALUES ('Milho'), ('Soja')")
			if insertProductErr != nil {
				panic(insertProductErr.Error())
			}
		}
	}
	handleStmtExec(c, stmt, err, "create product")
}

func initField(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init field table", `
	CREATE TABLE IF NOT EXISTS field (
    		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err, "create field")
}

func initCrop(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init crop table", `
	CREATE TABLE IF NOT EXISTS crop (
		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(255) UNIQUE NOT NULL,
		product SMALLINT NOT NULL,
		startDate DATE NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create crop")
}

func initVehicle(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init vehicle table", `
	CREATE TABLE IF NOT EXISTS vehicle (
		plate VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255)
	);
	`)

	handleStmtExec(c, stmt, err, "create vehicle")
}

func initEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry table", `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		field SMALLINT NOT NULL,
		crop SMALLINT NOT NULL,
		vehicle VARCHAR(255) NOT NULL,
		grossWeight DOUBLE PRECISION,
		tare DOUBLE PRECISION,
		netWeight DOUBLE PRECISION NOT NULL,
		humidity DOUBLE PRECISION,
		arrivalDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (field) REFERENCES field(id),
		FOREIGN KEY (crop) REFERENCES crop(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry")
}

func initInactiveEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init inactive_entry table", `
	CREATE TABLE IF NOT EXISTS inactive_entry (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		entry_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (entry_id) REFERENCES entry(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create inactive entry")
}

func initDeparture(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure table", `
	CREATE TABLE IF NOT EXISTS departure (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		departureDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		vehicle VARCHAR(255),
		crop SMALLINT NOT NULL,
		weight DOUBLE PRECISION NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id),
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (crop) REFERENCES crop(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create departure")
}

func initInactiveDeparture(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init inactive_departure table", `
	CREATE TABLE IF NOT EXISTS inactive_departure (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		departure_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (departure_id) REFERENCES departure(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create inactive departure")
}

func initBuyer(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init buyer table", `
	CREATE TABLE IF NOT EXISTS buyer (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		ie VARCHAR(255) UNIQUE
	);
	`)

	handleStmtExec(c, stmt, err, "create buyer")
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

	handleStmtExec(c, stmt, err, "create buyerPerson")
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

	handleStmtExec(c, stmt, err, "create buyerCompany")
}

func initDepartureBuyer(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure_buyer table", `
	CREATE TABLE IF NOT EXISTS departureBuyer (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		departureId INTEGER UNIQUE NOT NULL,
		buyerId INTEGER NOT NULL,
		FOREIGN KEY (buyerId) REFERENCES buyer(id),
		FOREIGN KEY (departureId) REFERENCES departure(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create departureBuyer")
}

func initAddrress(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init address table", `
	CREATE TABLE IF NOT EXISTS address (
		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		street VARCHAR(255) NOT NULL,
		cep VARCHAR(255) NOT NULL,
		number INTEGER,
		neighborhood VARCHAR(255) NOT NULL,
		city VARCHAR(255) NOT NULL,
		state VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err, "create address")
}

func initAddrressComplement(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init addressComplement table", `
	CREATE TABLE IF NOT EXISTS addressComplement (
		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		complement TEXT NOT NULL,
		addressId SMALLINT UNIQUE NOT NULL,
		FOREIGN KEY (addressId) REFERENCES address(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create addressComplement")
}

func initContact(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init contact table", `
	CREATE TABLE IF NOT EXISTS contact (
		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		email VARCHAR(255) NOT NULL,
		phoneNumber VARCHAR(255) NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err, "create contact")
}

func initLogTable(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init log table", `
	CREATE TABLE IF NOT EXISTS sys_log (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		content VARCHAR(255) NOT NULL,
		at TIMESTAMP WITHOUT TIME ZONE NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err, "create sys_log")
}

func initAddDepartureProcedure(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_departure(
			IN crop SMALLINT,
			IN buyerId INTEGER,
			OUT departureId INTEGER,
			OUT productName VARCHAR(255),
			INOUT vehicle VARCHAR(255),
			INOUT weight FLOAT,
			INOUT departureDate TIMESTAMP WITHOUT TIME ZONE
		)
		LANGUAGE plpgsql AS $$
		DECLARE departure_id INTEGER;
		BEGIN
			INSERT INTO departure (departureDate, vehicle, crop, weight) VALUES (departureDate, vehicle, crop, weight) RETURNING id INTO departure_id;
			INSERT INTO departurebuyer (departureId, buyerId) VALUES (departure_id, buyerId);

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = crop INTO productName;
			departureId := departure_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_departure:\n%v", err.Error())
	}
}

func initAddBuyerPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_buyer_person(
			IN ie VARCHAR(255),
			IN cpf VARCHAR(255),
			OUT buyerId INTEGER,
			INOUT name VARCHAR(255)
		)
		LANGUAGE plpgsql AS $$
		DECLARE buyer_id INTEGER;
		BEGIN
			INSERT INTO buyer (ie) VALUES (ie) RETURNING id INTO buyer_id;
			INSERT INTO buyerperson (name, cpf, buyerid) VALUES (name, cpf, buyer_id);
			
			buyerId := buyer_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_buyer_person:\n%v", err.Error())
	}
}

func initAddBuyerCompany(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_buyer_company(
			IN ie VARCHAR(255),
			IN cnpj VARCHAR(255),
			IN fantasyName VARCHAR(255),
			OUT buyerId INTEGER,
			INOUT companyName VARCHAR(255)
		)
		LANGUAGE plpgsql AS $$
		DECLARE buyer_id INTEGER;
		BEGIN
			INSERT INTO buyer (ie) VALUES (ie) RETURNING id INTO buyer_id;
			INSERT INTO buyercompany (cnpj, companyname, fantasyname, buyerid) VALUES (cnpj, companyName, fantasyName, buyer_id);
			
			buyerId := buyer_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_buyer_company:\n%v", err.Error())
	}
}

func InitDb(c *pgx.Conn) {
	initProduct(c)
	initCrop(c)
	initVehicle(c)
	initField(c)
	initEntry(c)
	initDeparture(c)
	initBuyer(c)
	initDepartureBuyer(c)
	initBuyerPerson(c)
	initBuyerCompany(c)
	initContact(c)
	initAddrress(c)
	initAddrressComplement(c)
	initLogTable(c)
	initInactiveDeparture(c)
	initInactiveEntry(c)
	initAddDepartureProcedure(c)
	initAddBuyerPerson(c)
	initAddBuyerCompany(c)
}

var dbc *pgx.Conn

func GetDbConnection() (*pgx.Conn, error) {
	if dbc == nil {
		dbc, err := pgx.Connect(context.Background(), "postgres://postgres:armazendapsswd@localhost:5432/postgres")

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)

			return nil, errors.New("Falha em conectar ao banco")
		}

		return dbc, nil
	}
	return dbc, nil
}
