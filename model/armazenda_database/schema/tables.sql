-- Tables are ordered by dependency: no foreign key references first,
-- then tables that reference them.

CREATE TABLE IF NOT EXISTS farm (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    inscricao_estadual TEXT UNIQUE NOT NULL
);

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

CREATE TABLE IF NOT EXISTS inactive_user (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES app_user(id)
);

CREATE TABLE IF NOT EXISTS user_session (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    session_id TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES app_user(id)
);

CREATE TABLE IF NOT EXISTS humidity_progression (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    farm_id INTEGER,
    is_system_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    CONSTRAINT single_system_default CHECK (
        (is_system_default = TRUE AND farm_id IS NULL) OR
        (is_system_default = FALSE)
    )
);

CREATE TABLE IF NOT EXISTS humidity_progression_tier (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    progression_id INTEGER NOT NULL,
    threshold_humidity NUMERIC(5, 2) NOT NULL,
    discount_value NUMERIC(5, 2) NOT NULL,
    FOREIGN KEY (progression_id) REFERENCES humidity_progression(id) ON DELETE CASCADE,
    UNIQUE(progression_id, threshold_humidity)
);

CREATE TABLE IF NOT EXISTS farm_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    humidity_progression_id INTEGER,
    storage_name TEXT NOT NULL,
    farm_used_humidity_progression_id INTEGER,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id),
    FOREIGN KEY (farm_used_humidity_progression_id) REFERENCES humidity_progression(id)
);

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

CREATE TABLE IF NOT EXISTS farm_address_complement (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    complement TEXT NOT NULL,
    farm_address_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (farm_address_id) REFERENCES farm_address(id)
);

CREATE TABLE IF NOT EXISTS farm_contact (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email TEXT,
    phone_number TEXT,
    farm_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (farm_id) REFERENCES farm(id)
);

CREATE TABLE IF NOT EXISTS person (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ie TEXT,
    farm INTEGER NOT NULL,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (farm) REFERENCES farm(id),
    CONSTRAINT unique_person_in_farm UNIQUE (farm, ie)
);

CREATE TABLE IF NOT EXISTS natural_person (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    cpf VARCHAR(14) NOT NULL,
    personId INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (personId) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS legal_person (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cnpj VARCHAR(18) NOT NULL,
    personId INTEGER UNIQUE NOT NULL,
    companyName TEXT NOT NULL,
    fantasyName TEXT,
    FOREIGN KEY (personId) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS person_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    person_id INTEGER UNIQUE NOT NULL,
    ie TEXT NOT NULL,
    farm INTEGER NOT NULL,
    humidity_progression_id INTEGER,
    entry_soy_discount NUMERIC (5, 2),
    entry_corn_discount NUMERIC (5, 2),
    FOREIGN KEY (farm) REFERENCES farm(id),
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id)
);

CREATE TABLE IF NOT EXISTS default_person_config (
    id INTEGER PRIMARY KEY DEFAULT 1,
    humidity_progression_id INTEGER,
    entry_soy_discount NUMERIC (5, 2) NOT NULL DEFAULT 3.5,
    entry_corn_discount NUMERIC (5, 2) NOT NULL DEFAULT 5.5,
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id),
    CONSTRAINT single_row CHECK (id = 1)
);

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

CREATE TABLE IF NOT EXISTS address_complement (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    complement TEXT NOT NULL,
    address_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (address_id) REFERENCES address(id)
);

CREATE TABLE IF NOT EXISTS contact (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email TEXT,
    phone_number TEXT,
    person_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (person_id) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS product (
    id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT UNIQUE NOT NULL
);

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

CREATE TABLE IF NOT EXISTS vehicle (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    plate TEXT NOT NULL,
    name TEXT,
    farm INTEGER NOT NULL,
    FOREIGN KEY (farm) REFERENCES farm(id),
    CONSTRAINT unique_vehicle_in_farm UNIQUE (plate, farm)
);

CREATE TABLE IF NOT EXISTS field (
    id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    farm INTEGER NOT NULL,
    hectares NUMERIC(10, 4) NOT NULL,
    FOREIGN KEY (farm) REFERENCES farm(id),
    CONSTRAINT unique_field_in_farm UNIQUE (name, farm)
);

CREATE TABLE IF NOT EXISTS entry (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    field SMALLINT NOT NULL,
    crop SMALLINT NOT NULL,
    vehicle INTEGER NOT NULL,
    grossWeight NUMERIC(12, 6) NOT NULL,
    tare NUMERIC(12, 6) NOT NULL,
    netWeight NUMERIC(12, 6) NOT NULL,
    arrivalDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    farm INTEGER NOT NULL,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle) REFERENCES vehicle(id),
    FOREIGN KEY (field) REFERENCES field(id),
    FOREIGN KEY (crop) REFERENCES crop(id),
    FOREIGN KEY (farm) REFERENCES farm(id)
);

CREATE TABLE IF NOT EXISTS entry_tax (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    entry_id INTEGER UNIQUE NOT NULL,
    weight NUMERIC(12, 6) NOT NULL,
    applied_tax NUMERIC(5, 2) NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES entry(id)
);

CREATE TABLE IF NOT EXISTS entry_origin (
    entry_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES entry(id),
    FOREIGN KEY (person_id) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS entry_draft (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    field SMALLINT,
    crop SMALLINT,
    vehicle INTEGER,
    tare NUMERIC(12, 6),
    farm INTEGER NOT NULL,
    FOREIGN KEY (vehicle) REFERENCES vehicle(id),
    FOREIGN KEY (field) REFERENCES field(id),
    FOREIGN KEY (crop) REFERENCES crop(id),
    FOREIGN KEY (farm) REFERENCES farm(id)
);

CREATE TABLE IF NOT EXISTS entry_draft_origin (
    entry_draft_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (entry_draft_id) REFERENCES entry_draft(id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS entry_analysis (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    humidity NUMERIC(5, 2),
    damage NUMERIC(5, 2),
    impurity NUMERIC(5, 2),
    entryId INTEGER UNIQUE NOT NULL,
    humidity_discount_modifier NUMERIC(5, 2),
    FOREIGN KEY (entryId) REFERENCES entry(id)
);

CREATE TABLE IF NOT EXISTS departure (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    departureDate TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    vehicle INTEGER NOT NULL,
    crop SMALLINT NOT NULL,
    grossWeight NUMERIC(12, 6) NOT NULL,
    tare NUMERIC(12, 6) NOT NULL,
    netWeight NUMERIC(12, 6) NOT NULL,
    farm INTEGER NOT NULL,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle) REFERENCES vehicle(id),
    FOREIGN KEY (crop) REFERENCES crop(id),
    FOREIGN KEY (farm) REFERENCES farm(id)
);

CREATE TABLE IF NOT EXISTS departure_analysis (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    humidity NUMERIC(6, 3),
    damage NUMERIC(6, 3),
    impurity NUMERIC(6, 3),
    departure_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (departure_id) REFERENCES departure(id)
);

CREATE TABLE IF NOT EXISTS departure_recipient (
    departure_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (person_id) REFERENCES person(id),
    FOREIGN KEY (departure_id) REFERENCES departure(id)
);

CREATE TABLE IF NOT EXISTS departure_origin (
    departure_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (person_id) REFERENCES person(id),
    FOREIGN KEY (departure_id) REFERENCES departure(id)
);

CREATE TABLE IF NOT EXISTS departure_draft (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    crop SMALLINT,
    vehicle INTEGER,
    tare NUMERIC(12, 6),
    farm INTEGER NOT NULL,
    FOREIGN KEY (vehicle) REFERENCES vehicle(id),
    FOREIGN KEY (crop) REFERENCES crop(id),
    FOREIGN KEY (farm) REFERENCES farm(id)
);

CREATE TABLE IF NOT EXISTS departure_draft_origin (
    departure_draft_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (departure_draft_id) REFERENCES departure_draft(id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS departure_draft_recipient (
    departure_draft_id INTEGER UNIQUE NOT NULL,
    person_id INTEGER NOT NULL,
    FOREIGN KEY (departure_draft_id) REFERENCES departure_draft(id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES person(id)
);

CREATE TABLE IF NOT EXISTS inactive_departure (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    departure_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (departure_id) REFERENCES departure(id)
);

CREATE TABLE IF NOT EXISTS inactive_entry (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    entry_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES entry(id)
);

CREATE TABLE IF NOT EXISTS sys_log (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    content TEXT NOT NULL,
    at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
