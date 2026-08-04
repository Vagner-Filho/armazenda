--
-- Test Farm and Admin User Seed File
--
-- This file creates the test infrastructure needed for e2e tests.
-- Run this SQL after starting the application for the first time.
--
-- Test Admin Credentials:
--   CPF: 52998224725
--   Password: TestAdmin123!
--   Role: admin
--   Email: test-admin@armazenda.test
--

-- Create test farm (required for the user)
INSERT INTO farm (inscricao_estadual, uf)
VALUES ('123456789', 'MT')
ON CONFLICT (inscricao_estadual) DO NOTHING;

INSERT INTO owner (owner_document, owner_document_type)
VALUES ('11111111111', 2);

INSERT INTO farm_owner_subscription
  (farm_id, owner_id, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, tier_key)
VALUES (1, 1, 'cus_Uv8dHw5pTqc0GW', 'sub_1TvIM4KAorCSDWaVL0Rvjwe1', 'active', '2037-07-21 09:41:46', 'fiscal');

-- Create farm config for test farm
INSERT INTO farm_config (farm_id, name, humidity_progression_id, storage_name)
SELECT
  id,
  'Test Farm',
  (SELECT id FROM humidity_progression WHERE is_system_default = TRUE),
  'Main Storage'
FROM farm WHERE inscricao_estadual = '123456789'
ON CONFLICT (farm_id) DO NOTHING;

-- Create test admin user
-- Password: TestAdmin123! (bcrypt hash below)
INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf, role)
SELECT 
  'test-admin@armazenda.test',
  'Test Admin',
  '$2a$10$AXTlVhKljI.rfz/TENyHfuc9wpikT15FNPM1iJOHPS7ON9FE7u0iq',
  '123456789',
  id,
  '52998224725',
  'admin'
FROM farm WHERE inscricao_estadual = '123456789'
ON CONFLICT (farm, cpf, email) DO NOTHING;

INSERT INTO crop (name, product, startDate, farm)
  SELECT
    'Milho Default',
    1,
    '2026-02-19',
    id
  FROM farm WHERE inscricao_estadual = '123456789';

INSERT INTO crop (name, product, startDate, farm)
  SELECT
    'Soja Default',
    2,
    '2026-02-19',
    id
  FROM farm WHERE inscricao_estadual = '123456789';

INSERT INTO field (name, farm, hectares)
  SELECT
    'Talhão Default',
    id,
    10.0
  FROM farm WHERE inscricao_estadual = '123456789';

INSERT INTO vehicle (plate, name, farm)
  SELECT
    'ABC123',
    '',
    id
  FROM farm WHERE inscricao_estadual = '123456789';
