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
	"github.com/jackc/pgx/v5/pgxpool"
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
    		name TEXT UNIQUE NOT NULL
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
		name TEXT NOT NULL,
		farm INTEGER NOT NULL,
		hectares NUMERIC(10, 4) NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id),
		CONSTRAINT unique_field_in_farm UNIQUE (name, farm)
	);
	`)

	handleStmtExec(c, stmt, err, "create field")
}

func initCrop(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init crop table", `
	CREATE TABLE IF NOT EXISTS crop (
		id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name TEXT NOT NULL,
		product SMALLINT NOT NULL,
		startDate DATE NOT NULL,
		farm INTEGER NOT NULL,
		FOREIGN KEY (product) REFERENCES product(id),
		FOREIGN KEY (farm) REFERENCES farm(id),
		CONSTRAINT unique_crop_in_farm UNIQUE (name, farm)
	);
	`)

	handleStmtExec(c, stmt, err, "create crop")
}

func initVehicle(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init vehicle table", `
	CREATE TABLE IF NOT EXISTS vehicle (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		plate TEXT NOT NULL,
		name TEXT,
		farm INTEGER NOT NULL,
		FOREIGN KEY (farm) REFERENCES farm(id),
		CONSTRAINT unique_vehicle_in_farm UNIQUE (plate, farm)
	);
	`)

	handleStmtExec(c, stmt, err, "create vehicle")
}

func initPreEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry draft table", `
		CREATE TABLE IF NOT EXISTS entry_draft (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name TEXT NOT NULL,
			field SMALLINT,
			crop SMALLINT,
			vehicle INTEGER,
			tare NUMERIC(10, 3),
			farm INTEGER NOT NULL,
			FOREIGN KEY (vehicle) REFERENCES vehicle(id),
			FOREIGN KEY (field) REFERENCES field(id),
			FOREIGN KEY (crop) REFERENCES crop(id),
			FOREIGN KEY (farm) REFERENCES farm(id)
		)
	`)
	handleStmtExec(c, stmt, err, "create entry_draft")
}

func initAddGetEntryDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_entry_draft;
		CREATE OR REPLACE FUNCTION add_get_entry_draft(
			in_name TEXT,
			in_field SMALLINT,
			in_crop SMALLINT,
			in_vehicle INTEGER,
			in_tare NUMERIC(10, 3),
			in_farm INTEGER,
			in_origin INTEGER,
			OUT out_id INTEGER,
			OUT out_name TEXT,
			OUT out_field_name TEXT,
			OUT out_crop_name TEXT,
			OUT out_vehicle_plate TEXT,
			OUT out_tare NUMERIC(10, 3),
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO entry_draft (name, field, crop, vehicle, tare, farm) VALUES (in_name, in_field, in_crop, in_vehicle, in_tare, in_farm) RETURNING entry_draft.id INTO out_id;

			IF in_origin IS NOT NULL THEN
				INSERT INTO entry_draft_origin (entry_draft_id, person_id) VALUES (out_id, in_origin);
				SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
				LEFT JOIN natural_person np ON p.id = np.personid
				LEFT JOIN legal_person lp ON p.id = lp.personid
				WHERE p.id = in_origin INTO out_origin;
			ELSE
				out_origin := 'Própria';
			END IF;

			SELECT f.name FROM field f WHERE f.id = in_field INTO out_field_name;
			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
			
			out_name := in_name;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_entry_draft:\n%v", err.Error())
	}
}

func initUpdateEntryDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_entry_draft;
		CREATE OR REPLACE FUNCTION update_get_entry_draft(
			INOUT draft_id INTEGER,
			IN in_name TEXT,
			IN in_field SMALLINT,
			IN in_crop SMALLINT,
			IN in_vehicle INTEGER,
			IN in_tare NUMERIC(10, 3),
			IN in_farm INTEGER,
			IN in_origin INTEGER,
			OUT out_name TEXT,
			OUT out_field_name TEXT,
			OUT out_crop_name TEXT,
			OUT out_vehicle_plate TEXT,
			OUT out_tare NUMERIC(10, 3),
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		DECLARE origin_exists BOOLEAN;
		BEGIN
			UPDATE entry_draft SET
				name = in_name,
				field = in_field,
				crop = in_crop,
				vehicle = in_vehicle,
				tare = in_tare
			WHERE id = draft_id;

			-- Handle origin relationship
			SELECT EXISTS (SELECT 1 FROM entry_draft_origin edo WHERE edo.entry_draft_id = draft_id) INTO origin_exists;
			
			IF in_origin IS NOT NULL THEN
				IF origin_exists THEN
					UPDATE entry_draft_origin SET person_id = in_origin WHERE entry_draft_id = draft_id;
					-- Get origin name
					SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
					LEFT JOIN natural_person np ON p.id = np.personid
					LEFT JOIN legal_person lp ON p.id = lp.personid
					WHERE p.id = in_origin INTO out_origin;
				ELSE
					INSERT INTO entry_draft_origin (entry_draft_id, person_id) VALUES (draft_id, in_origin);
					-- Get origin name
					SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
					LEFT JOIN natural_person np ON p.id = np.personid
					LEFT JOIN legal_person lp ON p.id = lp.personid
					WHERE p.id = in_origin INTO out_origin;
				END IF;
			ELSE
				-- Remove origin if it exists and new origin is null
				IF origin_exists THEN
					DELETE FROM entry_draft_origin WHERE entry_draft_id = draft_id;
				END IF;
				out_origin := 'Própria';
			END IF;

			-- Get related names
			SELECT f.name FROM field f WHERE f.id = in_field INTO out_field_name;
			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
			
			out_name := in_name;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_entry_draft:\n%v", err.Error())
	}
}

func initEntry(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry table", `
	CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		field SMALLINT NOT NULL,
		crop SMALLINT NOT NULL,
		vehicle INTEGER NOT NULL,
		grossWeight NUMERIC(10, 3) NOT NULL,
		tare NUMERIC(10, 3) NOT NULL,
		netWeight NUMERIC(10, 3) NOT NULL,
		arrivalDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		farm INTEGER NOT NULL,
		modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (vehicle) REFERENCES vehicle(id),
		FOREIGN KEY (field) REFERENCES field(id),
		FOREIGN KEY (crop) REFERENCES crop(id),
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry")
}

func initEntryTax(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry tax table", `
	CREATE TABLE IF NOT EXISTS entry_tax (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		entry_id INTEGER UNIQUE NOT NULL,
		weight NUMERIC(10, 3) NOT NULL,
		applied_tax NUMERIC(5, 2) NOT NULL,
		FOREIGN KEY (entry_id) REFERENCES entry(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry tax")
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
		humidity NUMERIC(5, 2),
		damage NUMERIC(5, 2),
		impurity NUMERIC(5, 2),
		entryId INTEGER UNIQUE NOT NULL,
		humidity_discount_modifier NUMERIC(5, 2),
		FOREIGN KEY (entryId) REFERENCES entry(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create entry_analysis")
}

func initDepartureAnalysis(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure_analysis table", `
	CREATE TABLE IF NOT EXISTS departure_analysis (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		humidity NUMERIC(6, 3),
		damage NUMERIC(6, 3),
		impurity NUMERIC(6, 3),
		departure_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (departure_id) REFERENCES departure(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create departure_analysis")
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
		vehicle INTEGER NOT NULL,
		crop SMALLINT NOT NULL,
		grossWeight NUMERIC(10, 3) NOT NULL,
		tare NUMERIC(10, 3) NOT NULL,
		netWeight NUMERIC(10, 3) NOT NULL,
		farm INTEGER NOT NULL,
		modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (vehicle) REFERENCES vehicle(id),
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
		modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
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
		humidity_discount NUMERIC(5, 2),
		entry_soy_discount NUMERIC (5, 2),
		entry_corn_discount NUMERIC (5, 2),
		FOREIGN KEY (farm) REFERENCES farm(id)
	);
	`)

	handleStmtExec(c, stmt, err, "create person config")
}

func initDefaultPersonConfig(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init default_person_config table", `
	CREATE TABLE IF NOT EXISTS default_person_config (
		id INTEGER PRIMARY KEY DEFAULT 1,
		humidity_discount NUMERIC(5, 2) NOT NULL DEFAULT 1.7,
		entry_soy_discount NUMERIC (5, 2) NOT NULL DEFAULT 3.5,
		entry_corn_discount NUMERIC (5, 2) NOT NULL DEFAULT 5.5,
		CONSTRAINT single_row CHECK (id = 1)
	);
	`)

	handleStmtExec(c, stmt, err, "create default person config")

	if err == nil {
		var count int
		err = c.QueryRow(context.Background(), "SELECT COUNT(*) FROM default_person_config").Scan(&count)
		if err == nil && count == 0 {
			_, insertErr := c.Exec(context.Background(),
				"INSERT INTO default_person_config (id, humidity_discount, entry_soy_discount, entry_corn_discount) VALUES (1, 1.7, 3.5, 5.5)")
			if insertErr != nil {
				fmt.Printf("error inserting default person config: %v\n", insertErr.Error())
			}
		}
	}
}

func initNaturalPerson(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init natural_person table", `
	CREATE TABLE IF NOT EXISTS natural_person (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		name TEXT NOT NULL,
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
		companyName TEXT NOT NULL,
		fantasyName TEXT,
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

func initDepartureOrigin(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure_origin table", `
	CREATE TABLE IF NOT EXISTS departure_origin (
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
		content TEXT NOT NULL,
		at TIMESTAMP WITHOUT TIME ZONE NOT NULL
	);
	`)

	handleStmtExec(c, stmt, err, "create sys_log")
}

func initAddDepartureProcedure(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_departure;
		CREATE OR REPLACE FUNCTION add_get_departure(
			IN crop SMALLINT,
			IN recipient_id INTEGER,
			OUT departureId INTEGER,
			OUT productName TEXT,
			IN in_vehicle INTEGER,
			OUT out_vehicle TEXT,
			INOUT departureDate TIMESTAMP WITHOUT TIME ZONE,
			IN farm INTEGER,
			IN grossWeight NUMERIC,
			IN tare NUMERIC,
			INOUT netWeight NUMERIC,
			IN in_humidity NUMERIC(6, 3),
			IN in_damage NUMERIC(6, 3),
			IN in_impurity NUMERIC(6, 3),
			IN in_origin_id INTEGER,
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		DECLARE departure_id INTEGER;
		BEGIN
			INSERT INTO departure (departureDate, vehicle, crop, farm, tare, grossWeight, netWeight) VALUES (departureDate, in_vehicle, crop, farm, tare, grossWeight, netWeight) RETURNING id INTO departure_id;

			IF in_humidity IS NOT NULL OR in_damage IS NOT NULL OR in_impurity IS NOT NULL THEN
				INSERT INTO departure_analysis (humidity, damage, impurity, departure_id) VALUES (in_humidity, in_damage, in_impurity, departure_id);
			END IF;

			IF recipient_id IS NOT NULL THEN
				INSERT INTO departure_recipient (departure_id, person_id) VALUES (departure_id, recipient_id);
			END IF;

			IF in_origin_id IS NOT NULL THEN
				INSERT INTO departure_origin (departure_id, person_id) VALUES (departure_id, in_origin_id);
			END IF;

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = crop INTO productName;
			departureId := departure_id;

			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle;

			SELECT COALESCE(name, 'Própria') FROM
			(SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT lp.companyname AS name, lp.personid FROM legal_person lp)
			WHERE personid = in_origin_id INTO out_origin;

			IF out_origin IS NULL THEN
				out_origin := 'Própria';
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_departure:\n%v", err.Error())
	}
}

func initAddLegalPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_legal_person;
		CREATE OR REPLACE FUNCTION add_get_legal_person(
			OUT person_type INTEGER,
			INOUT companyName TEXT,
			INOUT cnpj TEXT,
			INOUT ie TEXT,
			IN fantasyName TEXT,
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

func initUpdateNaturalPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_natural_person;
		CREATE OR REPLACE FUNCTION update_get_natural_person(
			OUT person_type INTEGER,
			INOUT name TEXT,
			INOUT cpf TEXT,
			INOUT ie TEXT,
			INOUT p_id INTEGER,
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
		DECLARE addressId INTEGER;
		DECLARE config_exists BOOLEAN;
		DECLARE address_exists BOOLEAN;
		DECLARE address_complement_exists BOOLEAN;
		DECLARE contact_exists BOOLEAN;
		BEGIN
			UPDATE person SET ie = update_get_natural_person.ie WHERE id = p_id;
			UPDATE natural_person SET name = update_get_natural_person.name, cpf = update_get_natural_person.cpf WHERE personId = p_id;

			person_type := 0;

			IF humidityDiscount IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM person_config pc WHERE pc.person_id = p_id) INTO config_exists;
				IF config_exists THEN
					UPDATE person_config SET humidity_discount = humidityDiscount WHERE person_id = p_id;
				ELSE
					INSERT INTO person_config (person_id, ie, farm, humidity_discount) VALUES (p_id, update_get_natural_person.ie, farm, humidityDiscount);
				END IF;
			END IF;

			IF street IS NOT NULL AND cep IS NOT NULL AND neighborhood IS NOT NULL AND city IS NOT NULL AND state IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM address a WHERE a.person_id = p_id) INTO address_exists;
				IF address_exists THEN
					UPDATE address SET street = update_get_natural_person.street, cep = update_get_natural_person.cep, number = update_get_natural_person.number, neighborhood = update_get_natural_person.neighborhood, city = update_get_natural_person.city, state = update_get_natural_person.state WHERE person_id = p_id RETURNING id INTO addressId;
				ELSE
					INSERT INTO address (street, cep, number, neighborhood, city, state, person_id) VALUES (street, cep, number, neighborhood, city, state, p_id) RETURNING id INTO addressId;
				END IF;

				IF complement IS NOT NULL AND addressId IS NOT NULL THEN
					SELECT EXISTS (SELECT 1 FROM address_complement ac WHERE ac.address_id = addressId) INTO address_complement_exists;
					IF address_complement_exists THEN
						UPDATE address_complement SET complement = update_get_natural_person.complement WHERE address_id = addressId;
					ELSE
						INSERT INTO address_complement (complement, address_id) VALUES (complement, addressId);
					END IF;
				END IF;
			END IF;

			IF email IS NOT NULL OR phoneNumber IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM contact c WHERE c.person_id = p_id) INTO contact_exists;
				IF contact_exists THEN
					UPDATE contact SET email = update_get_natural_person.email, phone_number = phoneNumber WHERE person_id = p_id;
				ELSE
					INSERT INTO contact (email, phone_number, person_id) VALUES (email, phoneNumber, p_id);
				END IF;
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_natural_person:\n%v", err.Error())
	}
}

func initUpdateLegalPerson(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_legal_person;
		CREATE OR REPLACE FUNCTION update_get_legal_person(
			OUT person_type INTEGER,
			INOUT p_companyName TEXT,
			INOUT cnpj TEXT,
			INOUT ie TEXT,
			INOUT p_id INTEGER,
			IN p_fantasyName TEXT,
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
		DECLARE addressId INTEGER;
		DECLARE config_exists BOOLEAN;
		DECLARE address_exists BOOLEAN;
		DECLARE address_complement_exists BOOLEAN;
		DECLARE contact_exists BOOLEAN;
		BEGIN
			UPDATE person SET ie = update_get_legal_person.ie WHERE id = p_id;
			UPDATE legal_person SET cnpj = update_get_legal_person.cnpj, companyname = p_companyName, fantasyname = p_fantasyName WHERE personId = p_id;

			person_type := 1;

			IF humidityDiscount IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM person_config pc WHERE pc.person_id = p_id) INTO config_exists;
				IF config_exists THEN
					UPDATE person_config SET humidity_discount = humidityDiscount WHERE person_id = p_id;
				ELSE
					INSERT INTO person_config (person_id, ie, farm, humidity_discount) VALUES (p_id, update_get_legal_person.ie, farm, humidityDiscount);
				END IF;
			END IF;

			IF street IS NOT NULL AND cep IS NOT NULL AND neighborhood IS NOT NULL AND city IS NOT NULL AND state IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM address a WHERE a.person_id = p_id) INTO address_exists;
				IF address_exists THEN
					UPDATE address SET street = update_get_legal_person.street, cep = update_get_legal_person.cep, number = update_get_legal_person.number, neighborhood = update_get_legal_person.neighborhood, city = update_get_legal_person.city, state = update_get_legal_person.state WHERE person_id = p_id RETURNING id INTO addressId;
				ELSE
					INSERT INTO address (street, cep, number, neighborhood, city, state, person_id) VALUES (street, cep, number, neighborhood, city, state, p_id) RETURNING id INTO addressId;
				END IF;

				IF complement IS NOT NULL AND addressId IS NOT NULL THEN
					SELECT EXISTS (SELECT 1 FROM address_complement ac WHERE ac.address_id = addressId) INTO address_complement_exists;
					IF address_complement_exists THEN
						UPDATE address_complement SET complement = update_get_legal_person.complement WHERE address_id = addressId;
					ELSE
						INSERT INTO address_complement (complement, address_id) VALUES (complement, addressId);
					END IF;
				END IF;
			END IF;

			IF email IS NOT NULL OR phoneNumber IS NOT NULL THEN
				SELECT EXISTS (SELECT 1 FROM contact c WHERE c.person_id = p_id) INTO contact_exists;
				IF contact_exists THEN
					UPDATE contact SET email = update_get_legal_person.email, phone_number = phoneNumber WHERE person_id = p_id;
				ELSE
					INSERT INTO contact (email, phone_number, person_id) VALUES (email, phoneNumber, p_id);
				END IF;
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_legal_person:\n%v", err.Error())
	}
}

func initAddEntry(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_entry;
		CREATE OR REPLACE FUNCTION add_get_entry(
			IN field SMALLINT,
			IN crop SMALLINT,
			IN grossWeight NUMERIC(10, 3),
			IN tare NUMERIC(10, 3),
		 	IN humidity NUMERIC(5, 2),
	
			OUT entryId INTEGER,
			OUT productName TEXT,
			OUT fieldName TEXT,

			IN in_vehicle INTEGER,
			OUT out_vehicle TEXT,
			INOUT netWeight NUMERIC(10, 3),
			INOUT arrivalDate TIMESTAMP WITHOUT TIME ZONE,
			INOUT farm INTEGER,
			IN damage NUMERIC(5, 2),
			IN impurity NUMERIC(5, 2),
			IN in_humidity_discount_modifier NUMERIC(5, 2),
			IN origin INTEGER,
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		DECLARE entry_id INTEGER;
		BEGIN
			INSERT INTO entry (field, crop, vehicle, grossweight, tare, netweight, arrivalDate, farm) VALUES (field, crop, in_vehicle, grossWeight, tare, netWeight, arrivalDate, farm) RETURNING id INTO entry_id;

			IF humidity IS NOT NULL OR damage IS NOT NULL OR impurity IS NOT NULL THEN
				INSERT INTO entry_analysis (humidity, damage, impurity, entryid, humidity_discount_modifier) VALUES (humidity, damage, impurity, entry_id, in_humidity_discount_modifier);
			END IF;
			
			IF origin IS NOT NULL THEN
				INSERT INTO entry_origin (entry_id, person_id) VALUES (entry_id, origin);
			END IF;

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = crop INTO productName;
			SELECT f.name FROM field f WHERE f.id = field INTO fieldName;
			entryId := entry_id;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle;

			SELECT COALESCE(name, 'Própria') FROM
			(SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT lp.companyname AS name, lp.personid FROM legal_person lp)
			WHERE personid = origin INTO out_origin;

			IF out_origin IS NULL THEN
				out_origin := 'Própria';
			END IF;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_entry:\n%v", err.Error())
	}
}

func initUpdateEntry(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_display_entry;
		CREATE OR REPLACE FUNCTION update_get_display_entry(
			in e_field SMALLINT,
			IN e_crop SMALLINT,
			in e_grossWeight DOUBLE PRECISION,
			in e_tare DOUBLE PRECISION,
		 	in e_humidity NUMERIC(6, 3),

			INOUT e_id INTEGER,
			OUT productName TEXT,
			OUT fieldName TEXT,
			IN in_vehicle INTEGER,
			OUT out_vehicle TEXT,
			INOUT e_netWeight DOUBLE PRECISION,
			INOUT e_arrivalDate TIMESTAMP WITHOUT TIME ZONE,
			OUT e_farm INTEGER,
			IN e_damage NUMERIC(5, 2),
			IN e_impurity NUMERIC(5, 2),
			IN in_humidity_discount_modifier NUMERIC(5, 2),
			IN origin_id INTEGER,
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		DECLARE analysis_exists BOOLEAN;
		DECLARE origin_exists BOOLEAN;
		BEGIN
			UPDATE entry e SET
				field = e_field,
				crop = e_crop,
				vehicle = in_vehicle,
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
					impurity = e_impurity,
					humidity_discount_modifier = in_humidity_discount_modifier
				WHERE ea.entryid = e_id;
			ELSIF e_humidity IS NOT NULL OR e_damage IS NOT NULL OR e_impurity IS NOT NULL THEN
				INSERT INTO entry_analysis (humidity, damage, impurity, entryid, humidity_discount_modifier) VALUES (e_humidity, e_damage, e_impurity, e_id, in_humidity_discount_modifier);
			END IF;

			SELECT EXISTS (SELECT 1 FROM entry_origin eo WHERE eo.entry_id = e_id) INTO origin_exists;
			IF origin_exists THEN
				IF origin_id IS NULL THEN
					DELETE FROM entry_origin WHERE entry_id = e_id;
				ELSE
					UPDATE entry_origin SET person_id = origin_id WHERE entry_id = e_id;
				END IF;
			ELSIF origin_id IS NOT NULL THEN
				INSERT INTO entry_origin (entry_id, person_id) VALUES (e_id, origin_id);
			END IF;

			SELECT COALESCE(name, 'Própria') FROM
			(SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT COALESCE(lp.fantasyname, lp.companyname) AS name, lp.personid FROM legal_person lp)
			WHERE personid = origin_id INTO out_origin;

			IF out_origin IS NULL THEN
				out_origin := 'Própria';
			END IF;

			SELECT p.name FROM product p JOIN crop c ON c.product = p.id WHERE c.id = e_crop INTO productName;
			SELECT f.name FROM field f WHERE f.id = e_field INTO fieldName;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle;
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
			role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
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
			storage_name TEXT NOT NULL,
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
			INOUT f_storage_name TEXT,
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
					UPDATE farm_config SET name = f_name, humidity_discount = f_humidity_discount, storage_name = f_storage_name WHERE farm_id = f_id;
				ELSE
					INSERT INTO farm_config (farm_id, name, humidity_discount, storage_name) VALUES (f_id, f_name, f_humidity_discount, f_storage_name);
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
	c.Exec(context.Background(), "DROP FUNCTION IF EXISTS update_get_departure;")
	stmt, err := c.Prepare(context.Background(), "init update departure stmt", `
		CREATE OR REPLACE FUNCTION update_get_departure(
			IN d_crop SMALLINT,
			IN in_recipient_id INTEGER,
			INOUT departureId INTEGER,
			OUT productName TEXT,
			IN in_vehicle INTEGER,
			OUT out_vehicle TEXT,
			INOUT departure_Date TIMESTAMP WITHOUT TIME ZONE,
			IN d_grossWeight NUMERIC,
			IN d_tare NUMERIC,
			INOUT d_netWeight NUMERIC,
			IN in_humidity NUMERIC(6, 3),
			IN in_damage NUMERIC(6, 3),
			IN in_impurity NUMERIC(6, 3),
			IN in_origin_id INTEGER,
			OUT out_origin TEXT
		)
		LANGUAGE plpgsql AS $$
		DECLARE analysis_exists BOOLEAN;
		DECLARE recipient_exists BOOLEAN;
		DECLARE origin_exists BOOLEAN;
		BEGIN
			UPDATE departure d SET
				 departureDate = departure_Date,
				 vehicle = in_vehicle,
				 crop = d_crop,
				 grossweight = d_grossWeight,
				 tare = d_tare,
				 netweight = d_netWeight
			WHERE d.id = departureId;

			SELECT EXISTS (SELECT 1 FROM departure_analysis da WHERE da.departure_id = departureId) INTO analysis_exists;

			IF analysis_exists THEN
				UPDATE departure_analysis da SET
					humidity = in_humidity,
					damage = in_damage,
					impurity = in_impurity
				WHERE da.departure_id = departureId;
			ELSIF in_humidity IS NOT NULL OR in_damage IS NOT NULL OR in_impurity IS NOT NULL THEN
				INSERT INTO departure_analysis (humidity, damage, impurity, departure_id) VALUES (in_humidity, in_damage, in_impurity, departureId);
			END IF;

			SELECT EXISTS (SELECT 1 FROM departure_recipient dor WHERE dor.departure_id = departureId) INTO recipient_exists;
			IF recipient_exists THEN
				IF in_recipient_id IS NULL THEN
					DELETE FROM departure_recipient WHERE departure_id = departureId;
				ELSE
					UPDATE departure_recipient SET person_id = in_recipient_id WHERE departure_id = departureId;
				END IF;
			ELSIF in_recipient_id IS NOT NULL THEN
				INSERT INTO departure_recipient (departure_id, person_id) VALUES (departureId, in_recipient_id);
			END IF;

			SELECT EXISTS (SELECT 1 FROM departure_origin dor WHERE dor.departure_id = departureId) INTO origin_exists;
			IF origin_exists THEN
				IF in_origin_id IS NULL THEN
					DELETE FROM departure_origin WHERE departure_id = departureId;
				ELSE
					UPDATE departure_origin SET person_id = in_origin_id WHERE departure_id = departureId;
				END IF;
			ELSIF in_origin_id IS NOT NULL THEN
				INSERT INTO departure_origin (departure_id, person_id) VALUES (departureId, in_origin_id);
			END IF;

			SELECT COALESCE(name, 'Própria') FROM
			(SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT COALESCE(lp.fantasyname, lp.companyname) AS name, lp.personid FROM legal_person lp)
			WHERE personid = in_origin_id INTO out_origin;

			IF out_origin IS NULL THEN
				out_origin := 'Própria';
			END IF;

			SELECT p.name INTO productName FROM departure d JOIN crop c ON d.crop = c.id JOIN product p ON c.product = p.id WHERE d.id = departureId;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle;
		END;
		$$;
	`)
	handleStmtExec(c, stmt, err, "init update departure proc")
}

func initDepartureDraft(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure draft table", `
		CREATE TABLE IF NOT EXISTS departure_draft (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name TEXT NOT NULL,
			crop SMALLINT,
			vehicle INTEGER,
			tare NUMERIC(10, 3),
			farm INTEGER NOT NULL,
			FOREIGN KEY (vehicle) REFERENCES vehicle(id),
			FOREIGN KEY (crop) REFERENCES crop(id),
			FOREIGN KEY (farm) REFERENCES farm(id)
		)
	`)
	handleStmtExec(c, stmt, err, "create departure_draft")
}

func initDepartureDraftOrigin(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure draft origin table", `
		CREATE TABLE IF NOT EXISTS departure_draft_origin (
			departure_draft_id INTEGER UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			FOREIGN KEY (departure_draft_id) REFERENCES departure_draft(id) ON DELETE CASCADE,
			FOREIGN KEY (person_id) REFERENCES person(id)
		);
		`)
	handleStmtExec(c, stmt, err, "create departure draft origin")
}

func initDepartureDraftRecipient(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init departure draft recipient table", `
		CREATE TABLE IF NOT EXISTS departure_draft_recipient (
			departure_draft_id INTEGER UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			FOREIGN KEY (departure_draft_id) REFERENCES departure_draft(id) ON DELETE CASCADE,
			FOREIGN KEY (person_id) REFERENCES person(id)
		);
		`)
	handleStmtExec(c, stmt, err, "create departure draft recipient")
}

func initAddGetDepartureDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS add_get_departure_draft;
		CREATE OR REPLACE FUNCTION add_get_departure_draft(
			in_name TEXT,
			in_recipient INTEGER,
			in_crop SMALLINT,
			in_vehicle INTEGER,
			in_tare NUMERIC(10, 3),
			in_farm INTEGER,
			in_origin INTEGER,
			OUT out_id INTEGER,
			OUT out_name TEXT,
			OUT out_origin_name TEXT,
			OUT out_crop_name TEXT,
			OUT out_vehicle_plate TEXT,
			OUT out_tare NUMERIC(10, 3)
		)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO departure_draft (name, crop, vehicle, tare, farm) VALUES (in_name, in_crop, in_vehicle, in_tare, in_farm) RETURNING departure_draft.id INTO out_id;

			IF in_origin IS NOT NULL THEN
				INSERT INTO departure_draft_origin (departure_draft_id, person_id) VALUES (out_id, in_origin);
				SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
				LEFT JOIN natural_person np ON p.id = np.personid
				LEFT JOIN legal_person lp ON p.id = lp.personid
				WHERE p.id = in_origin INTO out_origin_name;
			ELSE
				out_origin_name := 'Pŕopria';
			END IF;

			IF in_recipient IS NOT NULL THEN
				INSERT INTO departure_draft_recipient (departure_draft_id, person_id) VALUES (out_id, in_recipient);
			END IF;

			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;

			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
			
			out_name := in_name;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function add_get_departure_draft:\n%v", err.Error())
	}
}

func initUpdateDepartureDraft(c *pgx.Conn) {
	_, err := c.Exec(context.Background(), `
		DROP FUNCTION IF EXISTS update_get_departure_draft;
		CREATE OR REPLACE FUNCTION update_get_departure_draft(
			INOUT draft_id INTEGER,
			IN in_name TEXT,
			IN in_recipient INTEGER,
			IN in_crop SMALLINT,
			IN in_vehicle INTEGER,
			IN in_tare NUMERIC(10, 3),
			IN in_farm INTEGER,
			IN in_origin INTEGER,
			OUT out_name TEXT,
			OUT out_origin_name TEXT,
			OUT out_crop_name TEXT,
			OUT out_vehicle_plate TEXT,
			OUT out_tare NUMERIC(10, 3)
		)
		LANGUAGE plpgsql AS $$
		DECLARE origin_exists BOOLEAN;
		DECLARE recipient_exists BOOLEAN;
		BEGIN
			UPDATE departure_draft SET
				name = in_name,
				crop = in_crop,
				vehicle = in_vehicle,
				tare = in_tare
			WHERE id = draft_id;

			-- Handle origin relationship
			SELECT EXISTS (SELECT 1 FROM departure_draft_origin ddo WHERE ddo.departure_draft_id = draft_id) INTO origin_exists;
			
			IF in_origin IS NOT NULL THEN
				IF origin_exists THEN
					UPDATE departure_draft_origin SET person_id = in_origin WHERE departure_draft_id = draft_id;
					-- Get origin name
					SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
					LEFT JOIN natural_person np ON p.id = np.personid
					LEFT JOIN legal_person lp ON p.id = lp.personid
					WHERE p.id = in_origin INTO out_origin_name;
				ELSE
					INSERT INTO departure_draft_origin (departure_draft_id, person_id) VALUES (draft_id, in_origin);
					-- Get origin name
					SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
					LEFT JOIN natural_person np ON p.id = np.personid
					LEFT JOIN legal_person lp ON p.id = lp.personid
					WHERE p.id = in_origin INTO out_origin_name;
				END IF;
			ELSE
				-- Remove origin if it exists and new origin is null
				IF origin_exists THEN
					DELETE FROM departure_draft_origin WHERE departure_draft_id = draft_id;
				END IF;
				out_origin_name := 'Pŕopria';
			END IF;

			-- Handle recipient relationship
			SELECT EXISTS (SELECT 1 FROM departure_draft_recipient ddr WHERE ddr.departure_draft_id = draft_id) INTO recipient_exists;
			
			IF in_recipient IS NOT NULL THEN
				IF recipient_exists THEN
					UPDATE departure_draft_recipient SET person_id = in_recipient WHERE departure_draft_id = draft_id;
				ELSE
					INSERT INTO departure_draft_recipient (departure_draft_id, person_id) VALUES (draft_id, in_recipient);
				END IF;
			ELSE
				-- Remove recipient if it exists and new recipient is null
				IF recipient_exists THEN
					DELETE FROM departure_draft_recipient WHERE departure_draft_id = draft_id;
				END IF;
			END IF;

			-- Get related names
			SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
			SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
			
			out_name := in_name;
			out_tare := in_tare;
		END;
		$$;
	`)

	if err != nil {
		fmt.Printf("\n error at function update_get_departure_draft:\n%v", err.Error())
	}
}

func initEntryDraftOrigin(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init entry draft origin table", `
		CREATE TABLE IF NOT EXISTS entry_draft_origin (
			entry_draft_id INTEGER UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			FOREIGN KEY (entry_draft_id) REFERENCES entry_draft(id) ON DELETE CASCADE,
			FOREIGN KEY (person_id) REFERENCES person(id)
		);
		`)
	handleStmtExec(c, stmt, err, "create entry draft origin")
}

func initModifiedAtTriggers(c *pgx.Conn) {
	// Create the trigger function
	_, err := c.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION update_modified_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.modified_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		fmt.Printf("\n error creating update_modified_at function: %v\n", err.Error())
		return
	}

	// Create triggers for each table
	tables := []string{"entry", "departure", "person"}
	for _, table := range tables {
		triggerName := fmt.Sprintf("%s_modified_at_trigger", table)
		_, err := c.Exec(context.Background(), fmt.Sprintf(`
			DROP TRIGGER IF EXISTS %s ON %s;
			CREATE TRIGGER %s
				BEFORE UPDATE ON %s
				FOR EACH ROW
				EXECUTE FUNCTION update_modified_at();
		`, triggerName, table, triggerName, table))
		if err != nil {
			fmt.Printf("\n error creating trigger for %s: %v\n", table, err.Error())
		}
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
	initInactiveUser(c)
	initPerson(c)
	initProduct(c)
	initCrop(c)
	initVehicle(c)
	initField(c)
	initEntry(c)
	initEntryTax(c)
	initEntryOrigin(c)
	initPreEntry(c)
	initEntryDraftOrigin(c)
	initEntryAnalysis(c)
	initDeparture(c)
	initDepartureDraft(c)
	initDepartureDraftOrigin(c)
	initDepartureDraftRecipient(c)
	initUpdateDepartureDraft(c)
	initPersonConfig(c)
	initDefaultPersonConfig(c)
	initDepartureRecipient(c)
	initDepartureOrigin(c)
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
	initAddLegalPerson(c)
	initUpdateNaturalPerson(c)
	initUpdateLegalPerson(c)
	initUpdateDepartureProc(c)
	initAddGetEntryDraft(c)
	initUpdateEntryDraft(c)
	initAddGetDepartureDraft(c)
	initDepartureAnalysis(c)
	initModifiedAtTriggers(c)
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
			role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
			status TEXT NOT NULL DEFAULT 'pending',
			FOREIGN KEY (farm_id) REFERENCES farm(id)
		);
	`)
	handleStmtExec(c, stmt, err, "init user approval table")
}

func initInactiveUser(c *pgx.Conn) {
	stmt, err := c.Prepare(context.Background(), "init inactive user stmt", `
		CREATE TABLE IF NOT EXISTS inactive_user (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			user_id INTEGER NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES app_user(id)
		);
	`)
	handleStmtExec(c, stmt, err, "init inactive user table")
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
			// Handle error
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
