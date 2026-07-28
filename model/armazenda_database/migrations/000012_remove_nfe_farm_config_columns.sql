ALTER TABLE IF EXISTS nfe_farm_config
	DROP COLUMN IF EXISTS emitter_type,
	DROP COLUMN IF EXISTS cnpj_emitter,
	DROP COLUMN IF EXISTS cpf_emitter,
	DROP COLUMN IF EXISTS emitter_uf,
	DROP COLUMN IF EXISTS ie_emitter;
