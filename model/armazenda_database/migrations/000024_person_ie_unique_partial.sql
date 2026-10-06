-- 000024_person_ie_unique_partial.sql
--
-- IE (Inscrição Estadual) is optional. The baseline constraint
-- unique_person_in_farm UNIQUE (farm, ie) treats two contacts without an IE
-- (stored as an empty string) as duplicates, which breaks creating more than
-- one inline NF-e recipient without IE in the same farm.
--
-- Replace it with a partial unique index so real IEs remain unique per farm
-- while empty/absent IEs may coexist.

ALTER TABLE person DROP CONSTRAINT IF EXISTS unique_person_in_farm;

CREATE UNIQUE INDEX IF NOT EXISTS unique_person_ie_in_farm
    ON person (farm, ie)
    WHERE ie IS NOT NULL AND ie <> '';
