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
    		name VARCHAR(255) UNIQUE NOT NULL
	);
	`)
	handleStmtExec(c, stmt, err, "create product")

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
}

func initField(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init field table", `
	CREATE TABLE IF NOT EXISTS field (
    		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(255) NOT NULL,
		farm INTEGER NOT NULL,
		hectares NUMERIC(10, 4) NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id)
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
		farm INTEGER NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id),
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create crop")
}

func initVehicle(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init vehicle table", `
	CREATE TABLE IF NOT EXISTS vehicle (
		plate VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255),
		farm INTEGER NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create vehicle")
}

func initPreEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry draft table", `
		CREATE TABLE IF NOT EXISTS entry_draft (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			field SMALLINT,
			crop SMALLINT,
			vehicle VARCHAR(255),
			grossWeight NUMERIC(10, 3),
			tare NUMERIC(10, 3),
			netWeight NUMERIC(10, 3),
			arrivalDate TIMESTAMP WITHOUT TIME ZONE,
			farm INTEGER NOT NULL,
			startedAt TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
			FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
			FOREIGN KEY (field) REFERENCES field(id),
			FOREIGN KEY (crop) REFERENCES crop(id),
			FOREIGN KEY (farm) REFERENCES farm(id)
		)
	`)
	handleStmtExec(c, stmt, err, "create entry_draft")
}

func initEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry table", `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		field SMALLINT NOT NULL,
		crop SMALLINT NOT NULL,
		vehicle VARCHAR(255) NOT NULL,
		grossWeight NUMERIC(10, 3) NOT NULL,
		tare NUMERIC(10, 3) NOT NULL,
		netWeight NUMERIC(10, 3) NOT NULL,
		arrivalDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		farm INTEGER NOT NULL,
		finishedAt TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (field) REFERENCES field(id),
		FOREIGN KEY (crop) REFERENCES crop(id),
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry")
}

func initEntryAnalysis(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry_analysis table", `
	CREATE TABLE IF NOT EXISTS entry_analysis (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		humidity NUMERIC(6, 3),
		damage NUMERIC(6, 3),
		impurity NUMERIC(6, 3),
		entryId INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (entryId) REFERENCES entry(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry_analysis")
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
		grossWeight NUMERIC(10, 3) NOT NULL,
		tare NUMERIC(10, 3) NOT NULL,
		netWeight NUMERIC(10, 3) NOT NULL,
		farm INTEGER NOT NULL,
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (crop) REFERENCES crop(id),
		FOREIGN KEY (farm) REFERENCES farm(id)
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
		ie VARCHAR(255) UNIQUE,
		farm INTEGER NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id)
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
			INOUT departureDate TIMESTAMP WITHOUT TIME ZONE,
			IN farm INTEGER,
			IN grossWeight NUMERIC,
			IN tare NUMERIC,
			INOUT netWeight NUMERIC
		)
		LANGUAGE plpgsql AS $$
		DECLARE departure_id INTEGER;
		BEGIN
			INSERT INTO departure (departureDate, vehicle, crop, farm, tare, grossWeight, netWeight) VALUES (departureDate, vehicle, crop, farm, tare, grossWeight, netWeight) RETURNING id INTO departure_id;
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
			INOUT name VARCHAR(255),
			IN farm INTEGER
		)
		LANGUAGE plpgsql AS $$
		DECLARE buyer_id INTEGER;
		BEGIN
			INSERT INTO buyer (ie, farm) VALUES (ie, farm) RETURNING id INTO buyer_id;
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
			IN farm INTEGER,
			OUT buyerId INTEGER,
			INOUT companyName VARCHAR(255)
		)
		LANGUAGE plpgsql AS $$
		DECLARE buyer_id INTEGER;
		BEGIN
			INSERT INTO buyer (ie, farm) VALUES (ie, farm) RETURNING id INTO buyer_id;
			INSERT INTO buyercompany (cnpj, companyname, fantasyname, buyerid) VALUES (cnpj, companyName, fantasyName, buyer_id);
			
			buyerId := buyer_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_buyer_company:\n%v", err.Error())
	}
}

func initAddEntry(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_entry(
			IN field SMALLINT,
			IN crop SMALLINT,
			IN grossWeight NUMERIC(10, 3),
			IN tare NUMERIC(10, 3),
		 	IN humidity NUMERIC(6, 3),
	
			OUT entryId INTEGER,
			OUT productName VARCHAR(255),
			OUT fieldName VARCHAR(255),

			INOUT vehicle VARCHAR(255),
			INOUT netWeight NUMERIC(10, 3),
			INOUT arrivalDate TIMESTAMP WITHOUT TIME ZONE,
			INOUT farm INTEGER,
			IN damage NUMERIC(6, 3),
			IN impurity NUMERIC(6, 3)
		)
		LANGUAGE plpgsql AS $$
		DECLARE entry_id INTEGER;
		BEGIN
			INSERT INTO entry (field, crop, vehicle, grossweight, tare, netweight, arrivalDate, farm) VALUES (field, crop, vehicle, grossWeight, tare, netWeight, arrivalDate, farm) RETURNING id INTO entry_id;

			IF humidity IS NOT NULL OR damage IS NOT NULL OR impurity IS NOT NULL THEN
				INSERT INTO entry_analysis (humidity, damage, impurity, entryid) VALUES (humidity, damage, impurity, entry_id);
			END IF;

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = crop INTO productName;
			SELECT f.name FROM field f WHERE f.id = field INTO fieldName;
			entryId := entry_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_entry:\n%v", err.Error())
	}
}

func initUpdateEntry(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION update_get_display_entry(
			in e_field SMALLINT,
			IN e_crop SMALLINT,
			in e_grossWeight DOUBLE PRECISION,
			in e_tare DOUBLE PRECISION,
		 	in e_humidity NUMERIC(6, 3),

			INOUT e_id INTEGER,
			OUT productName VARCHAR(255),
			OUT fieldName VARCHAR(255),
			INOUT e_vehicle VARCHAR(255),
			INOUT e_netWeight DOUBLE PRECISION,
			INOUT e_arrivalDate TIMESTAMP WITHOUT TIME ZONE,
			OUT e_farm INTEGER,
			IN e_damage NUMERIC(6, 3),
			IN e_impurity NUMERIC(6, 3)
		)
		LANGUAGE plpgsql AS $$
		DECLARE analysis_exists BOOLEAN;
		BEGIN
			UPDATE entry e SET
				field = e_field,
				crop = e_crop,
				vehicle = e_vehicle,
				grossweight = e_grossWeight,
				tare = e_tare,
				netweight = e_netWeight,
				arrivalDate = e_arrivalDate
			WHERE e.id = e_id RETURNING e.farm INTO e_farm;

			SELECT EXISTS (SELECT 1 FROM entry_analysis ea WHERE ea.entryid = e_id) INTO analysis_exists;
			
			IF analysis_exists THEN
				UPDATE entry_analysis ea SET
					humidity = e_humidity,
					damage = e_damage,
					impurity = e_impurity
				WHERE ea.entryid = e_id;
			ELSIF e_humidity IS NOT NULL OR e_damage IS NOT NULL OR e_impurity IS NOT NULL THEN
				INSERT INTO entry_analysis (humidity, damage, impurity, entryid) VALUES (e_humidity, e_damage, e_impurity, e_id);
			END IF;

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = e_crop INTO productName;
			SELECT f.name FROM field f WHERE f.id = e_field INTO fieldName;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_display_entry:\n%v", err.Error())
	}
}
func initUser(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init user stmt", `
		CREATE TABLE IF NOT EXISTS app_user (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			passwd TEXT NOT NULL,
			inscricao_estadual TEXT NOT NULL,
			farm INTEGER NOT NULL,
			cpf VARCHAR(11) UNIQUE NOT NULL,
			FOREIGN KEY (farm) REFERENCES farm(id)
		);
	`)
	handleStmtExec(c, stmt, err, "init user table")
}

// inserts to farm if there is no other farm with the same inscricao estadual
func initAddUserAndFarm(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_app_user(
			IN email TEXT,
			IN name TEXT,
			IN passwd TEXT,
			IN ie TEXT,
			IN cpf VARCHAR(11)
		)
		RETURNS VOID
		LANGUAGE plpgsql AS $$
		DECLARE farm_exists BOOLEAN; farm_id INTEGER;
		BEGIN
			SELECT EXISTS (SELECT 1 FROM farm WHERE inscricao_estadual = ie) INTO farm_exists;
		
			IF NOT farm_exists THEN
				INSERT INTO farm (inscricao_estadual) VALUES (ie) RETURNING id INTO farm_id;
				INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf) VALUES (email, name, passwd, ie, farm_id, cpf);
			ELSE
				SELECT id FROM farm WHERE inscricao_estadual = ie INTO farm_id;
				INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf) VALUES (email, name, passwd, ie, farm_id, cpf);
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_app_user:\n%v", err.Error())
	}
}

func initFarm(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init farm stmt", `
		CREATE TABLE IF NOT EXISTS farm (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			inscricao_estadual TEXT UNIQUE NOT NULL
		);
	`)
	handleStmtExec(c, stmt, err, "init farm table")
}

func InitDb(c *pgx.Conn) {
	initFarm(c)
	initUser(c)
	initProduct(c)
	initCrop(c)
	initVehicle(c)
	initField(c)
	initEntry(c)
	initPreEntry(c)
	initEntryAnalysis(c)
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
	initAddEntry(c)
	initUpdateEntry(c)
	initAddDepartureProcedure(c)
	initAddBuyerPerson(c)
	initAddBuyerCompany(c)
	initAddUserAndFarm(c)
}

var dbc *pgx.Conn

func GetDbConnection() (*pgx.Conn, error) {
	if dbc == nil {
		dbHost := os.Getenv("DB_HOST")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASS")
		dbName := os.Getenv("DB_NAME")
		dbPort := os.Getenv("DB_PORT")
		//dbc, err := pgx.Connect(context.Background(), "postgres://armazenda_user:y34xEy2HR09pibXFA6ngrku7@localhost:5432/armazenda_db")
		dbc, err := pgx.Connect(context.Background(), "postgres://"+dbUser+":"+dbPass+"@"+dbHost+":"+dbPort+"/"+dbName)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			fmt.Printf("host | user | psswd | name | port\n%v | %v | %v | %v | %v\n", dbHost, dbUser, dbPass, dbName, dbPort)
			os.Exit(1)

			return nil, errors.New("Falha em conectar ao banco")
		}

		return dbc, nil
	}
	return dbc, nil
}
