-- farm_product: farm-scoped products created by the user, used by the
-- detached NF-e feature. Mirrors the definition of `product` (id, name,
-- ncm) plus a farm_id column; independent from the global catalog.
CREATE TABLE IF NOT EXISTS farm_product (
    id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    ncm TEXT NOT NULL DEFAULT '00000000',
    farm_id INTEGER NOT NULL,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    CONSTRAINT unique_product_name_in_farm UNIQUE (farm_id, name)
);

-- Guard against user mistakes: a farm product cannot have the same name nor
-- the same NCM as any global catalog product, so the two product lists stay
-- unambiguous side by side.
CREATE OR REPLACE FUNCTION forbid_farm_product_global_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM product WHERE name = NEW.name) THEN
        RAISE EXCEPTION 'Ja existe um produto global com o nome "%" — selecione o produto global na lista', NEW.name;
    END IF;
    IF EXISTS (SELECT 1 FROM product WHERE ncm = NEW.ncm) THEN
        RAISE EXCEPTION 'Ja existe um produto global com o NCM %', NEW.ncm;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS farm_product_no_global_overlap ON farm_product;
CREATE TRIGGER farm_product_no_global_overlap
    BEFORE INSERT OR UPDATE ON farm_product
    FOR EACH ROW
    EXECUTE FUNCTION forbid_farm_product_global_overlap();
