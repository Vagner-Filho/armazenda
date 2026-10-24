CREATE TABLE IF NOT EXISTS farm_logo (
    farm_id    INTEGER PRIMARY KEY REFERENCES farm(id) ON DELETE CASCADE,
    logo_path  TEXT,
    logo_data  BYTEA,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);