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
    INOUT f_humidity_progression_id INTEGER,
    INOUT f_farm_used_humidity_progression_id INTEGER
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
            IF f_humidity_progression_id IS NOT NULL OR f_farm_used_humidity_progression_id IS NOT NULL THEN
                UPDATE farm_config SET name = f_name, humidity_progression_id = f_humidity_progression_id, storage_name = f_storage_name, farm_used_humidity_progression_id = f_farm_used_humidity_progression_id WHERE farm_id = f_id;
            ELSE
                UPDATE farm_config SET name = f_name, storage_name = f_storage_name WHERE farm_id = f_id;
            END IF;
        ELSE
            INSERT INTO farm_config (farm_id, name, humidity_progression_id, storage_name, farm_used_humidity_progression_id) VALUES (f_id, f_name, f_humidity_progression_id, f_storage_name, f_farm_used_humidity_progression_id);
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

CREATE OR REPLACE FUNCTION add_get_entry(
    IN field SMALLINT,
    IN crop SMALLINT,
    IN grossWeight NUMERIC(12, 6),
    IN tare NUMERIC(12, 6),
    IN humidity NUMERIC(5, 2),
    OUT entryId INTEGER,
    OUT productName TEXT,
    OUT fieldName TEXT,
    IN in_vehicle INTEGER,
    OUT out_vehicle TEXT,
    INOUT netWeight NUMERIC(12, 6),
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

CREATE OR REPLACE FUNCTION add_get_entry_draft(
    in_name TEXT,
    in_field SMALLINT,
    in_crop SMALLINT,
    in_vehicle INTEGER,
    in_tare NUMERIC(12, 6),
    in_farm INTEGER,
    in_origin INTEGER,
    OUT out_id INTEGER,
    OUT out_name TEXT,
    OUT out_field_name TEXT,
    OUT out_crop_name TEXT,
    OUT out_vehicle_plate TEXT,
    OUT out_tare NUMERIC(12, 6),
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

CREATE OR REPLACE FUNCTION update_get_entry_draft(
    INOUT draft_id INTEGER,
    IN in_name TEXT,
    IN in_field SMALLINT,
    IN in_crop SMALLINT,
    IN in_vehicle INTEGER,
    IN in_tare NUMERIC(12, 6),
    IN in_farm INTEGER,
    IN in_origin INTEGER,
    OUT out_name TEXT,
    OUT out_field_name TEXT,
    OUT out_crop_name TEXT,
    OUT out_vehicle_plate TEXT,
    OUT out_tare NUMERIC(12, 6),
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

    SELECT EXISTS (SELECT 1 FROM entry_draft_origin edo WHERE edo.entry_draft_id = draft_id) INTO origin_exists;
    
    IF in_origin IS NOT NULL THEN
        IF origin_exists THEN
            UPDATE entry_draft_origin SET person_id = in_origin WHERE entry_draft_id = draft_id;
            SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
            LEFT JOIN natural_person np ON p.id = np.personid
            LEFT JOIN legal_person lp ON p.id = lp.personid
            WHERE p.id = in_origin INTO out_origin;
        ELSE
            INSERT INTO entry_draft_origin (entry_draft_id, person_id) VALUES (draft_id, in_origin);
            SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
            LEFT JOIN natural_person np ON p.id = np.personid
            LEFT JOIN legal_person lp ON p.id = lp.personid
            WHERE p.id = in_origin INTO out_origin;
        END IF;
    ELSE
        IF origin_exists THEN
            DELETE FROM entry_draft_origin WHERE entry_draft_id = draft_id;
        END IF;
        out_origin := 'Própria';
    END IF;

    SELECT f.name FROM field f WHERE f.id = in_field INTO out_field_name;
    SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
    SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
    
    out_name := in_name;
    out_tare := in_tare;
END;
$$;

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
    OUT out_origin TEXT,
    OUT out_recipient_id INTEGER,
    OUT out_origin_id INTEGER
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

    out_recipient_id := recipient_id;
    out_origin_id := in_origin_id;
END;
$$;

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
    OUT out_origin TEXT,
    OUT out_recipient_id INTEGER,
    OUT out_origin_id INTEGER
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

    out_recipient_id := in_recipient_id;
    out_origin_id := in_origin_id;
END;
$$;

CREATE OR REPLACE FUNCTION add_get_departure_draft(
    in_name TEXT,
    in_recipient INTEGER,
    in_crop SMALLINT,
    in_vehicle INTEGER,
    in_tare NUMERIC(12, 6),
    in_farm INTEGER,
    in_origin INTEGER,
    OUT out_id INTEGER,
    OUT out_name TEXT,
    OUT out_origin_name TEXT,
    OUT out_crop_name TEXT,
    OUT out_vehicle_plate TEXT,
    OUT out_tare NUMERIC(12, 6)
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

CREATE OR REPLACE FUNCTION update_get_departure_draft(
    INOUT draft_id INTEGER,
    IN in_name TEXT,
    IN in_recipient INTEGER,
    IN in_crop SMALLINT,
    IN in_vehicle INTEGER,
    IN in_tare NUMERIC(12, 6),
    IN in_farm INTEGER,
    IN in_origin INTEGER,
    OUT out_name TEXT,
    OUT out_origin_name TEXT,
    OUT out_crop_name TEXT,
    OUT out_vehicle_plate TEXT,
    OUT out_tare NUMERIC(12, 6)
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

    SELECT EXISTS (SELECT 1 FROM departure_draft_origin ddo WHERE ddo.departure_draft_id = draft_id) INTO origin_exists;
    
    IF in_origin IS NOT NULL THEN
        IF origin_exists THEN
            UPDATE departure_draft_origin SET person_id = in_origin WHERE departure_draft_id = draft_id;
            SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
            LEFT JOIN natural_person np ON p.id = np.personid
            LEFT JOIN legal_person lp ON p.id = lp.personid
            WHERE p.id = in_origin INTO out_origin_name;
        ELSE
            INSERT INTO departure_draft_origin (departure_draft_id, person_id) VALUES (draft_id, in_origin);
            SELECT COALESCE(np.name, lp.fantasyname, lp.companyname) FROM person p
            LEFT JOIN natural_person np ON p.id = np.personid
            LEFT JOIN legal_person lp ON p.id = lp.personid
            WHERE p.id = in_origin INTO out_origin_name;
        END IF;
    ELSE
        IF origin_exists THEN
            DELETE FROM departure_draft_origin WHERE departure_draft_id = draft_id;
        END IF;
        out_origin_name := 'Pŕopria';
    END IF;

    SELECT EXISTS (SELECT 1 FROM departure_draft_recipient ddr WHERE ddr.departure_draft_id = draft_id) INTO recipient_exists;
    
    IF in_recipient IS NOT NULL THEN
        IF recipient_exists THEN
            UPDATE departure_draft_recipient SET person_id = in_recipient WHERE departure_draft_id = draft_id;
        ELSE
            INSERT INTO departure_draft_recipient (departure_draft_id, person_id) VALUES (draft_id, in_recipient);
        END IF;
    ELSE
        IF recipient_exists THEN
            DELETE FROM departure_draft_recipient WHERE departure_draft_id = draft_id;
        END IF;
    END IF;

    SELECT c.name FROM crop c WHERE c.id = in_crop INTO out_crop_name;
    SELECT v.plate FROM vehicle v WHERE v.id = in_vehicle INTO out_vehicle_plate;
    
    out_name := in_name;
    out_tare := in_tare;
END;
$$;

CREATE OR REPLACE FUNCTION add_get_legal_person(
    OUT person_type INTEGER,
    INOUT companyName TEXT,
    INOUT cnpj TEXT,
    INOUT ie TEXT,
    IN fantasyName TEXT,
    IN farm INTEGER,
    OUT personId INTEGER,
    IN humidityProgressionId INTEGER DEFAULT NULL,
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
    
    IF humidityProgressionId IS NOT NULL THEN
        INSERT INTO person_config (person_id, ie, farm, humidity_progression_id) VALUES (person_id, ie, farm, humidityProgressionId);
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

CREATE OR REPLACE FUNCTION update_get_natural_person(
    OUT person_type INTEGER,
    INOUT name TEXT,
    INOUT cpf TEXT,
    INOUT ie TEXT,
    INOUT p_id INTEGER,
    IN farm INTEGER,
    IN humidityProgressionId INTEGER DEFAULT NULL,
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

    IF humidityProgressionId IS NOT NULL THEN
        SELECT EXISTS (SELECT 1 FROM person_config pc WHERE pc.person_id = p_id) INTO config_exists;
        IF config_exists THEN
            UPDATE person_config SET humidity_progression_id = humidityProgressionId WHERE person_id = p_id;
        ELSE
            INSERT INTO person_config (person_id, ie, farm, humidity_progression_id) VALUES (p_id, update_get_natural_person.ie, farm, humidityProgressionId);
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

CREATE OR REPLACE FUNCTION update_get_legal_person(
    OUT person_type INTEGER,
    INOUT p_companyName TEXT,
    INOUT cnpj TEXT,
    INOUT ie TEXT,
    INOUT p_id INTEGER,
    IN p_fantasyName TEXT,
    IN farm INTEGER,
    IN humidityProgressionId INTEGER DEFAULT NULL,
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

    IF humidityProgressionId IS NOT NULL THEN
        SELECT EXISTS (SELECT 1 FROM person_config pc WHERE pc.person_id = p_id) INTO config_exists;
        IF config_exists THEN
            UPDATE person_config SET humidity_progression_id = humidityProgressionId WHERE person_id = p_id;
        ELSE
            INSERT INTO person_config (person_id, ie, farm, humidity_progression_id) VALUES (p_id, update_get_legal_person.ie, farm, humidityProgressionId);
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
            SELECT EXISTS (SELECT 1 FROM address_address_complement ac WHERE ac.address_id = addressId) INTO address_complement_exists;
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
