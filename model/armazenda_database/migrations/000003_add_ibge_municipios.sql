CREATE TABLE IF NOT EXISTS ibge_municipio (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    uf TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ibge_municipio_uf ON ibge_municipio(uf);
