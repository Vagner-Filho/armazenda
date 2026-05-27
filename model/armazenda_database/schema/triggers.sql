CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS entry_modified_at_trigger ON entry;
CREATE TRIGGER entry_modified_at_trigger
    BEFORE UPDATE ON entry
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

DROP TRIGGER IF EXISTS departure_modified_at_trigger ON departure;
CREATE TRIGGER departure_modified_at_trigger
    BEFORE UPDATE ON departure
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

DROP TRIGGER IF EXISTS person_modified_at_trigger ON person;
CREATE TRIGGER person_modified_at_trigger
    BEFORE UPDATE ON person
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();
