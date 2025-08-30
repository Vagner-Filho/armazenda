package armazenda_database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	zerologadapter "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			name VARCHAR(255) NOT NULL,
			field SMALLINT,
			crop SMALLINT,
			vehicle VARCHAR(255),
			tare NUMERIC(10, 3),
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

func initAddGetEntryDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_entry_draft(
			in_name VARCHAR(255),
			in_field SMALLINT,
			in_crop SMALLINT,
			in_vehicle VARCHAR(255),
			in_tare NUMERIC(10, 3),
			in_farm INTEGER,
			OUT out_id INTEGER,
			OUT out_name VARCHAR(255),
			OUT out_field_name VARCHAR(255),
			OUT out_crop_name VARCHAR(255),
			OUT out_vehicle_plate VARCHAR(255),
			OUT out_tare NUMERIC(10, 3)
		)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO entry_draft (name, field, crop, vehicle, tare, farm) VALUES (in_name, in_field, in_crop, in_vehicle, in_tare, in_farm) RETURNING entry_draft.id INTO out_id;

			SELECT f.name FROM field f WHERE f.id = in_field INTO out_field_name;
			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
			
			out_name := in_name;
			out_vehicle_plate := in_vehicle;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_entry_draft:\n%v", err.Error())
	}
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
		FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
		FOREIGN KEY (field) REFERENCES field(id),
		FOREIGN KEY (crop) REFERENCES crop(id),
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry")
}

func initEntryOrigin(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry origin table", `
	CREATE TABLE IF NOT EXISTS entry_origin (
		entry_id INTEGER UNIQUE NOT NULL,
		person_id INTEGER NOT NULL,
		FOREIGN KEY (entry_id) REFERENCES entry(id),
		FOREIGN KEY (person_id) REFERENCES person(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry origin")
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

func initPerson(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init person table", `
	CREATE TABLE IF NOT EXISTS person (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		ie TEXT,
		farm INTEGER NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id),
		CONSTRAINT unique_person_in_farm UNIQUE (farm, ie)
	);
	`)

	handleStmtExec(c, stmt, err, "create person")
}

func initPersonConfig(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init person_config table", `
	CREATE TABLE IF NOT EXISTS person_config (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		person_id INTEGER UNIQUE NOT NULL,
		ie TEXT NOT NULL,
		farm INTEGER NOT NULL,
		humidity_discount NUMERIC(6, 3) NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create person config")
}

func initNaturalPerson(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init natural_person table", `
	CREATE TABLE IF NOT EXISTS natural_person (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(255) NOT NULL,
		cpf VARCHAR(11) NOT NULL,
		personId INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (personId) REFERENCES person(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create natural_person")
}

func initLegalPerson(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init legal_person table", `
	CREATE TABLE IF NOT EXISTS legal_person (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		cnpj VARCHAR(14) NOT NULL,
		personId INTEGER UNIQUE NOT NULL,
		companyName VARCHAR(255) NOT NULL,
		fantasyName VARCHAR(255),
		FOREIGN KEY (personId) REFERENCES person(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create legal_person")
}

func initDepartureRecipient(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure_recipient table", `
	CREATE TABLE IF NOT EXISTS departure_recipient (
		departure_id INTEGER UNIQUE NOT NULL,
		person_id INTEGER NOT NULL,
		FOREIGN KEY (person_id) REFERENCES person(id),
		FOREIGN KEY (departure_id) REFERENCES departure(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create departure_recipient")
}

func initAddrress(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init address table", `
	CREATE TABLE IF NOT EXISTS address (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		street TEXT NOT NULL,
		cep CHARACTER(8) NOT NULL,
		number INTEGER,
		neighborhood TEXT NOT NULL,
		city TEXT NOT NULL,
		state CHARACTER(2) NOT NULL,
		person_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (person_id) REFERENCES person(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create address")
}

func initAddrressComplement(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init address_complement table", `
	CREATE TABLE IF NOT EXISTS address_complement (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		complement TEXT NOT NULL,
		address_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (address_id) REFERENCES address(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create address_complement")
}

func initContact(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init contact table", `
	CREATE TABLE IF NOT EXISTS contact (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		email TEXT,
		phone_number TEXT,
		person_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (person_id) REFERENCES person(id)
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
			IN personId INTEGER,
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
			INSERT INTO departure_recipient (departure_id, person_id) VALUES (departure_id, personId);

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = crop INTO productName;
			departureId := departure_id;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_departure:\n%v", err.Error())
	}
}

func initAddNaturalPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_natural_person;
		CREATE OR REPLACE FUNCTION add_get_natural_person(
			OUT person_type INTEGER,
			INOUT name VARCHAR(255),
			INOUT cpf VARCHAR(255),
			INOUT ie VARCHAR(255),
			OUT personId INTEGER,
			IN farm INTEGER,
			IN humidityDiscount NUMERIC(6, 3) DEFAULT NULL,
			IN street TEXT DEFAULT NULL,
			IN cep CHARACTER(8) DEFAULT NULL,
			IN number INTEGER DEFAULT NULL,
			IN neighborhood TEXT DEFAULT NULL,
			IN city TEXT DEFAULT NULL,
			IN state CHARACTER(2) DEFAULT NULL,
			IN complement TEXT DEFAULT NULL,
			IN email TEXT DEFAULT NULL,
			IN phoneNumber TEXT DEFAULT NULL
		)
		LANGUAGE plpgsql AS $$
		DECLARE person_id INTEGER;
		DECLARE addressId INTEGER;
		BEGIN
			INSERT INTO person (ie, farm) VALUES (ie, farm) RETURNING id INTO person_id;
			INSERT INTO natural_person (name, cpf, personId) VALUES (name, cpf, person_id);

			personId := person_id;
			person_type := 0;

			IF humidityDiscount IS NOT NULL THEN
				INSERT INTO person_config (person_id, ie, farm, humidity_discount) VALUES (person_id, ie, farm, humidityDiscount);
			END IF;

			IF street IS NOT NULL AND cep IS NOT NULL AND neighborhood IS NOT NULL AND city IS NOT NULL AND state IS NOT NULL THEN
				INSERT INTO address (street, cep, number, neighborhood, city, state, person_id) VALUES (street, cep, number, neighborhood, city, state, person_id) RETURNING id INTO addressId;

				IF complement IS NOT NULL THEN
					INSERT INTO address_complement (complement, address_id) VALUES (complement, addressId);
				END IF;
			END IF;

			IF email IS NOT NULL OR phoneNumber IS NOT NULL THEN
				INSERT INTO contact (email, phone_number, person_id) VALUES (email, phoneNumber, person_id);
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_natural_person:\n%v", err.Error())
	}
}

func initAddLegalPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_legal_person;
		CREATE OR REPLACE FUNCTION add_get_legal_person(
			OUT person_type INTEGER,
			INOUT companyName VARCHAR(255),
			INOUT cnpj VARCHAR(14),
			INOUT ie VARCHAR(255),
			IN fantasyName VARCHAR(255),
			IN farm INTEGER,
			OUT personId INTEGER,
			IN humidityDiscount NUMERIC(6, 3) DEFAULT NULL,
			IN street TEXT DEFAULT NULL,
			IN cep CHARACTER(8) DEFAULT NULL,
			IN number INTEGER DEFAULT NULL,
			IN neighborhood TEXT DEFAULT NULL,
			IN city TEXT DEFAULT NULL,
			IN state CHARACTER(2) DEFAULT NULL,
			IN complement TEXT DEFAULT NULL,
			IN email TEXT DEFAULT NULL,
			IN phoneNumber TEXT DEFAULT NULL
		)
		LANGUAGE plpgsql AS $$
		DECLARE person_id INTEGER;
		DECLARE addressId INTEGER;
		BEGIN
			INSERT INTO person (ie, farm) VALUES (ie, farm) RETURNING id INTO person_id;
			INSERT INTO legal_person (cnpj, companyname, fantasyname, personid) VALUES (cnpj, companyName, fantasyName, person_id);

			personId := person_id;
			person_type := 1;
			
			IF humidityDiscount IS NOT NULL THEN
				INSERT INTO person_config (person_id, ie, farm, humidity_discount) VALUES (person_id, ie, farm, humidityDiscount);
			END IF;

			IF street IS NOT NULL AND cep IS NOT NULL AND neighborhood IS NOT NULL AND city IS NOT NULL AND state IS NOT NULL THEN
				INSERT INTO address (street, cep, number, neighborhood, city, state, person_id) VALUES (street, cep, number, neighborhood, city, state, person_id) RETURNING id INTO addressId;

				IF complement IS NOT NULL THEN
					INSERT INTO address_complement (complement, address_id) VALUES (complement, addressId);
				END IF;
			END IF;

			IF email IS NOT NULL OR phoneNumber IS NOT NULL THEN
				INSERT INTO contact (email, phone_number, person_id) VALUES (email, phoneNumber, person_id);
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_legal_person:\n%v", err.Error())
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
			IN impurity NUMERIC(6, 3),
			IN origin INTEGER
		)
		LANGUAGE plpgsql AS $$
		DECLARE entry_id INTEGER;
		BEGIN
			INSERT INTO entry (field, crop, vehicle, grossweight, tare, netweight, arrivalDate, farm) VALUES (field, crop, vehicle, grossWeight, tare, netWeight, arrivalDate, farm) RETURNING id INTO entry_id;

			IF humidity IS NOT NULL OR damage IS NOT NULL OR impurity IS NOT NULL THEN
				INSERT INTO entry_analysis (humidity, damage, impurity, entryid) VALUES (humidity, damage, impurity, entry_id);
			END IF;
			
			IF origin IS NOT NULL THEN
				INSERT INTO entry_origin (entry_id, person_id) VALUES (entry_id, origin);
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
			email TEXT NOT NULL,
			name TEXT NOT NULL,
			passwd TEXT NOT NULL,
			inscricao_estadual TEXT NOT NULL,
			farm INTEGER NOT NULL,
			cpf VARCHAR(11) NOT NULL,
			FOREIGN KEY (farm) REFERENCES farm(id),
			CONSTRAINT unique_user_in_farm UNIQUE (farm, cpf, email)
		);
	`)
	handleStmtExec(c, stmt, err, "init user table")
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

func initFarmConfig(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init farm_config table", `
			CREATE TABLE IF NOT EXISTS farm_config (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			farm_id INTEGER NOT NULL UNIQUE,
			name TEXT NOT NULL,
			humidity_discount NUMERIC(6, 3) DEFAULT 1.15,
			FOREIGN KEY (farm_id) REFERENCES farm(id)
		);
	`)

	handleStmtExec(c, stmt, err, "create farm_config")
}

func initFarmAddrress(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init farm_address table", `
	CREATE TABLE IF NOT EXISTS farm_address (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		street TEXT,
		cep CHARACTER(8) NOT NULL,
		number INTEGER,
		neighborhood TEXT,
		city TEXT NOT NULL,
		state CHARACTER(2) NOT NULL,
		farm_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (farm_id) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create farm_address")
}

func initFarmAddrressComplement(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init farm_address_complement table", `
	CREATE TABLE IF NOT EXISTS farm_address_complement (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		complement TEXT NOT NULL,
		farm_address_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (farm_address_id) REFERENCES farm_address(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create farm_address_complement")
}

func initFarmContact(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init farm_contact table", `
	CREATE TABLE IF NOT EXISTS farm_contact (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		email TEXT,
		phone_number TEXT,
		farm_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (farm_id) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create contact")
}

func initFarmUpdateFunc(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_farm;
		CREATE OR REPLACE FUNCTION update_get_farm(
			INOUT f_id INTEGER,
			INOUT f_inscricao_estadual TEXT,
			INOUT f_name TEXT,
			INOUT f_street TEXT,
			INOUT f_cep CHARACTER(8),
			INOUT f_number INTEGER,
			INOUT f_neighborhood TEXT,
			INOUT f_city TEXT,
			INOUT f_state CHARACTER(2),
			INOUT f_complement TEXT,
			INOUT f_email TEXT,
			INOUT f_phone_number TEXT,
			INOUT f_humidity_discount NUMERIC(6, 3) DEFAULT 1.15
		)
		LANGUAGE plpgsql AS $$
		DECLARE var_farm_address_id INTEGER;
		DECLARE config_exists BOOLEAN;
		DECLARE address_exists BOOLEAN;
		DECLARE address_complement_exists BOOLEAN;
		DECLARE farm_contact_exists BOOLEAN;
		BEGIN
			UPDATE farm SET inscricao_estadual = f_inscricao_estadual WHERE id = f_id;

			IF f_name IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM farm_config fc WHERE fc.farm_id = f_id) INTO config_exists;

				IF config_exists THEN
					UPDATE farm_config SET name = f_name, humidity_discount = f_humidity_discount WHERE farm_id = f_id;
				ELSE
					INSERT INTO farm_config (farm_id, name, humidity_discount) VALUES (f_id, f_name, f_humidity_discount);
				END IF;
			END IF;

			IF f_cep IS NOT NULL AND f_city IS NOT NULL AND f_state IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM farm_address fa WHERE fa.farm_id = f_id) INTO address_exists;

				IF address_exists THEN
					UPDATE farm_address SET street = f_street, cep = f_cep, number = f_number, neighborhood = f_neighborhood, city = f_city, state = f_state WHERE farm_id = f_id RETURNING id INTO var_farm_address_id;
				ELSE
					INSERT INTO farm_address (street, cep, number, neighborhood, city, state, farm_id) VALUES (f_street, f_cep, f_number, f_neighborhood, f_city, f_state, f_id) RETURNING id INTO var_farm_address_id;
				END IF;

				IF f_complement IS NOT NULL AND var_farm_address_id IS NOT NULL THEN
					SELECT EXISTS (SELECT 1 FROM farm_address_complement fac WHERE fac.farm_address_id = var_farm_address_id) INTO address_complement_exists;
					IF address_complement_exists THEN
						UPDATE farm_address_complement SET complement = f_complement, farm_address_id = var_farm_address_id WHERE farm_address_id = var_farm_address_id;
					ELSE
						INSERT INTO farm_address_complement (complement, farm_address_id) VALUES (f_complement, var_farm_address_id);
					END IF;
				END IF;
			END IF;

			IF f_email IS NOT NULL OR f_phone_number IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM farm_contact fc WHERE fc.farm_id = f_id) INTO farm_contact_exists;
				IF farm_contact_exists THEN
					UPDATE farm_contact SET email = f_email, phone_number = f_phone_number WHERE farm_id = f_id;
				ELSE
					INSERT INTO farm_contact (email, phone_number, farm_id) VALUES (f_email, f_phone_number, f_id);
				END IF;
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_farm:\n%v", err.Error())
	}
}

func initUpdateDepartureProc(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init update departure stmt", `
		CREATE OR REPLACE FUNCTION update_get_departure(
			IN d_crop SMALLINT,
			IN d_personId INTEGER,
			INOUT departureId INTEGER,
			OUT productName VARCHAR(255),
			INOUT d_vehicle VARCHAR(255),
			INOUT departure_Date TIMESTAMP WITHOUT TIME ZONE,
			IN d_grossWeight NUMERIC,
			IN d_tare NUMERIC,
			INOUT d_netWeight NUMERIC,
			OUT farm INTEGER
		)
		LANGUAGE plpgsql AS $$
		BEGIN
			UPDATE departure d SET
				 departureDate = departure_Date,
				 vehicle = d_vehicle,
				 crop = d_crop,
				 grossweight = d_grossWeight,
				 tare = d_tare,
				 netweight = d_netWeight
			WHERE d.id = departureId;

			SELECT p.name, d.farm INTO productName, farm FROM departure d JOIN crop c ON d.crop = c.id JOIN product p ON c.product = p.id WHERE d.id = departureId;
		END;
		$$;
	`)
	handleStmtExec(c, stmt, err, "init update departure proc")
}

func initDepartureDraft(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure draft table", `
		CREATE TABLE IF NOT EXISTS departure_draft (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name VARCHAR(255) NOT NULL,
			person INTEGER,
			crop SMALLINT,
			vehicle VARCHAR(255),
			tare NUMERIC(10, 3),
			farm INTEGER NOT NULL,
			startedAt TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
			FOREIGN KEY (vehicle) REFERENCES vehicle(plate),
			FOREIGN KEY (person) REFERENCES person(id),
			FOREIGN KEY (crop) REFERENCES crop(id),
			FOREIGN KEY (farm) REFERENCES farm(id)
		)
	`)
	handleStmtExec(c, stmt, err, "create departure_draft")
}

func initAddGetDepartureDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION add_get_departure_draft(
			in_name VARCHAR(255),
			in_person INTEGER,
			in_crop SMALLINT,
			in_vehicle VARCHAR(255),
			in_tare NUMERIC(10, 3),
			in_farm INTEGER,
			OUT out_id INTEGER,
			OUT out_name VARCHAR(255),
			OUT out_person_name VARCHAR(255),
			OUT out_crop_name VARCHAR(255),
			OUT out_vehicle_plate VARCHAR(255),
			OUT out_tare NUMERIC(10, 3)
		)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO departure_draft (name, person, crop, vehicle, tare, farm) VALUES (in_name, in_person, in_crop, in_vehicle, in_tare, in_farm) RETURNING departure_draft.id INTO out_id;

			SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p 
			LEFT JOIN natural_person np ON p.id = np.personid
			LEFT JOIN legal_person lp ON p.id = lp.personid
			WHERE p.id = in_person INTO out_person_name;

			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
			
			out_name := in_name;
			out_vehicle_plate := in_vehicle;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_departure_draft:\n%v", err.Error())
	}
}

func InitDb(c *pgx.Conn) {
	initFarm(c)
	initFarmConfig(c)
	initFarmAddrress(c)
	initFarmAddrressComplement(c)
	initFarmContact(c)
	initFarmUpdateFunc(c)
	initUser(c)
	initUserApproval(c)
	initProduct(c)
	initCrop(c)
	initVehicle(c)
	initField(c)
	initEntry(c)
	initEntryOrigin(c)
	initPreEntry(c)
	initEntryAnalysis(c)
	initDeparture(c)
	initDepartureDraft(c)
	initPerson(c)
	initPersonConfig(c)
	initDepartureRecipient(c)
	initNaturalPerson(c)
	initLegalPerson(c)
	initContact(c)
	initAddrress(c)
	initAddrressComplement(c)
	initLogTable(c)
	initInactiveDeparture(c)
	initInactiveEntry(c)
	initAddEntry(c)
	initUpdateEntry(c)
	initAddDepartureProcedure(c)
	initAddNaturalPerson(c)
	initAddLegalPerson(c)
	initUpdateDepartureProc(c)
	initAddGetEntryDraft(c)
	initAddGetDepartureDraft(c)
}

func initUserApproval(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init user approval stmt", `
		CREATE TABLE IF NOT EXISTS user_approval (

			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			passwd TEXT NOT NULL,
			inscricao_estadual TEXT NOT NULL,
			farm_id INTEGER NOT NULL,
			cpf VARCHAR(11) UNIQUE NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			FOREIGN KEY (farm_id) REFERENCES farm(id)
		);
	`)
	handleStmtExec(c, stmt, err, "init user approval table")
}

var dbc *pgx.Conn

func GetDbConnection() (*pgx.Conn, error) {
	if dbc == nil {
		dbHost := os.Getenv("DB_HOST")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASS")
		dbName := os.Getenv("DB_NAME")
		dbPort := os.Getenv("DB_PORT")

		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		zlogger := zerolog.New(output).Level(zerolog.ErrorLevel).With().Timestamp().Logger()

		connString := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName
		config, err := pgx.ParseConfig(connString)
		if err != nil {
			// Handle error
			log.Fatal().Err(err).Msg("Failed to parse connection string")
			fmt.Printf("host | user | psswd | name | port\n%v | %v | %v | %v | %v\n", dbHost, dbUser, dbPass, dbName, dbPort)
			os.Exit(1)

			return nil, errors.New("Falha em conectar ao banco")
		}
		config.Tracer = &tracelog.TraceLog{
			LogLevel: tracelog.LogLevelError,
			Logger:   zerologadapter.NewLogger(zlogger),
		}

		dbc, err := pgx.ConnectConfig(context.Background(), config)

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
