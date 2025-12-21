package person_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type personModel struct {
	pool *pgxpool.Pool
}

var personModelImpl *personModel

func InitPersonModel(pool *pgxpool.Pool) (*personModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if personModelImpl == nil {
		personModelImpl = &personModel{
			pool: pool,
		}
	}

	return personModelImpl, nil
}

func GetpersonModel() *personModel {
	if personModelImpl == nil {
		panic("\nperson model hasnt been initialized\n")
	}
	return personModelImpl
}

type basePerson struct {
	entity_public.Address
	entity_public.Person
}

const basePersonQuery = `
	SELECT ad.id, ad.street, ad.cep, ad.number, adc.complement, ad.neighborhood, ad.city, ad.state, c.email, c.phone_number, p.ie, p.id, p.farm, pc.humidity_discount FROM person p
		LEFT JOIN address ad ON ad.person_id = p.id
		LEFT JOIN address_complement adc ON adc.address_id = ad.id
		LEFT JOIN contact c ON c.person_id = p.id
		LEFT JOIN person_config pc ON pc.person_id = p.id
		WHERE p.id = @id
	`

func (pm *personModel) getBasePerson(id uint32) (basePerson, *model_error.ModelError) {
	var base basePerson
	row, queryErr := pm.pool.Query(context.Background(), basePersonQuery,
		pgx.NamedArgs{"id": id})

	if queryErr != nil {
		return base, &model_error.ModelError{Message: queryErr.Error(), IsServerErr: true}
	}
	base, collectErr := pgx.CollectExactlyOneRow(row, pgx.RowToStructByPos[basePerson])

	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return base, &model_error.ModelError{Message: "Pessoa não encontrada"}
		}
		return base, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}
	return base, nil
}

func (bm *personModel) AddLegalPerson(bc entity_public.LegalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	row, queryErr := bm.pool.Query(
		context.Background(),
		`SELECT * FROM add_get_legal_person(@companyName, @cnpj, @ie, @fantasyName, @farm, @humidityDiscount, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone)`,
		pgx.NamedArgs{
			"ie":               bc.Person.Ie,
			"cnpj":             bc.Cnpj,
			"fantasyName":      bc.FantasyName,
			"farm":             bc.Person.Farm,
			"companyName":      bc.CompanyName,
			"humidityDiscount": bc.Person.HumidityDiscount,
			"street":           bc.Street,
			"cep":              bc.Cep,
			"number":           bc.Number,
			"neighborhood":     bc.Neighborhood,
			"city":             bc.City,
			"state":            bc.State,
			"complement":       bc.Complement,
			"email":            bc.Email,
			"phone":            bc.PhoneNumber,
		})

	if queryErr != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	person, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.PersonDisplay])
	if collectErr != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return person, nil
}

func (bm *personModel) AddNaturalPerson(bp entity_public.NaturalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	ctx := context.Background()
	tx, err := bm.pool.Begin(ctx)
	if err != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	defer tx.Rollback(ctx)

	// 1. Insert Person
	var personID uint32
	err = tx.QueryRow(ctx, "INSERT INTO person (ie, farm) VALUES ($1, $2) RETURNING id", bp.Person.Ie, bp.Person.Farm).Scan(&personID)
	if err != nil {
		return handleInsertError(err)
	}

	// 2. Insert Natural Person
	_, err = tx.Exec(ctx, "INSERT INTO natural_person (name, cpf, personId) VALUES ($1, $2, $3)", bp.Name, bp.Cpf, personID)
	if err != nil {
		return handleInsertError(err)
	}

	// 3. Insert Config (if any discount is provided)
	if !bp.Person.HumidityDiscount.IsZero() || !bp.Person.EntrySoyDiscount.IsZero() || !bp.Person.EntryCornDiscount.IsZero() {
		_, err = tx.Exec(ctx, `INSERT INTO person_config 
			(person_id, ie, farm, humidity_discount, entry_soy_discount, entry_corn_discount) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			personID, bp.Person.Ie, bp.Person.Farm,
			bp.Person.HumidityDiscount, bp.Person.EntrySoyDiscount, bp.Person.EntryCornDiscount)
		if err != nil {
			return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
		}
	}

	// 4. Insert Address (if required fields are present)
	if bp.Street != nil && bp.Cep != nil && bp.Neighborhood != nil && bp.City != nil && bp.State != nil {
		var addressID uint32
		err = tx.QueryRow(ctx, `INSERT INTO address 
			(street, cep, number, neighborhood, city, state, person_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			bp.Street, bp.Cep, bp.Number, bp.Neighborhood, bp.City, bp.State, personID).Scan(&addressID)
		if err != nil {
			return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
		}

		if bp.Complement != nil {
			_, err = tx.Exec(ctx, "INSERT INTO address_complement (complement, address_id) VALUES ($1, $2)", bp.Complement, addressID)
			if err != nil {
				return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
			}
		}
	}

	// 5. Insert Contact
	if bp.Email != nil || bp.PhoneNumber != nil {
		_, err = tx.Exec(ctx, "INSERT INTO contact (email, phone_number, person_id) VALUES ($1, $2, $3)", bp.Email, bp.PhoneNumber, personID)
		if err != nil {
			return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return entity_public.PersonDisplay{
		Type:     0, // Natural Person
		Name:     bp.Name,
		Document: bp.Cpf,
		IE:       bp.Person.Ie,
		Id:       personID,
	}, nil
}

func handleInsertError(err error) (entity_public.PersonDisplay, *model_error.ModelError) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			var message string = "Esta pessoa já existe"
			if strings.Contains(err.Error(), "person_ie") {
				message = "Inscrição Estadual em uso"
			}
			if strings.Contains(err.Error(), "natural_person_cpf") {
				message = "CPF em uso"
			}
			return entity_public.PersonDisplay{}, &model_error.ModelError{Message: message}
		}
	}
	return entity_public.PersonDisplay{}, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
}

func (bm *personModel) GetPeopleByFarm(farm uint32) ([]entity_public.PersonOption, *model_error.ModelError) {
	rows, queryErr := bm.pool.Query(context.Background(), `
		SELECT result.id, result.name, COALESCE(result.humidity_discount, dpc.humidity_discount), COALESCE(result.entry_soy_discount, dpc.entry_soy_discount), COALESCE(result.entry_corn_discount, dpc.entry_corn_discount) FROM (
			SELECT p.id, COALESCE(lp.fantasyname, lp.companyname) AS name, pc.humidity_discount as humidity_discount, pc.entry_soy_discount, pc.entry_corn_discount
			FROM person p
			JOIN legal_person lp ON p.id = lp.personid
			LEFT JOIN person_config pc ON p.id = pc.person_id
			WHERE p.farm = @userFarm
			UNION
			SELECT p.id, np.name, pc.humidity_discount as humidity_discount, pc.entry_soy_discount, pc.entry_corn_discount
			FROM person p
			JOIN natural_person np ON p.id = np.personid
			LEFT JOIN person_config pc ON p.id = pc.person_id
			WHERE p.farm = @userFarm
			UNION
			SELECT NULL, 'Própria', COALESCE((SELECT humidity_discount FROM farm_config WHERE farm_id = @userFarm), 1.15) as humidity_discount, 0.0, 0.0
		) AS result, default_person_config dpc
		ORDER BY CASE WHEN result.id IS NULL THEN 0 ELSE 1 END, name
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		return []entity_public.PersonOption{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	people, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.PersonOption])
	if collectErr != nil {
		return []entity_public.PersonOption{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return people, nil
}

var availablePersonFilters = map[string]func(pf entity_public.PersonFilter) string{
	"Vehicle": func(pf entity_public.PersonFilter) string {
		return fmt.Sprintf("e.vehicle = '%s'", pf.Vehicle)
	},
	"Product": func(pf entity_public.PersonFilter) string {
		return "p.id = " + strconv.FormatInt(int64(pf.Product), 10)
	},
	"Field": func(pf entity_public.PersonFilter) string {
		return "e.field = " + strconv.FormatInt(int64(pf.Field), 10)
	},
	"NetWeightMin": func(pf entity_public.PersonFilter) string {
		return "e.netweight >= " + strconv.FormatFloat(pf.NetWeightMin, 'f', -1, 64)
	},
	"NetWeightMax": func(pf entity_public.PersonFilter) string {
		return "e.netweight <= " + strconv.FormatFloat(pf.NetWeightMax, 'f', -1, 64)
	},
	"Crop": func(pf entity_public.PersonFilter) string {
		return fmt.Sprintf("c.id = %v", pf.Crop)
	},
}

func (bm *personModel) FilterPerson(pf entity_public.PersonFilter, farm uint32, page, limit int) ([]entity_public.PersonDisplay, int, *model_error.ModelError) {
	countStmt := `SELECT COUNT(*) FROM person WHERE farm = @userFarm`

	var total int
	err := bm.pool.QueryRow(context.Background(), countStmt, pgx.NamedArgs{"userFarm": farm}).Scan(&total)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []entity_public.PersonDisplay{}, 0, nil
		}
		return nil, 0, &model_error.ModelError{Message: "Error counting people", IsServerErr: true}
	}

	if total == 0 {
		return []entity_public.PersonDisplay{}, 0, nil
	}

	offset := (page - 1) * limit
	stmt := `
        SELECT * FROM (
                SELECT 0 AS TYPE, np.name, np.cpf AS document, p.ie, np.personid AS id FROM natural_person np
                JOIN person p ON p.id = np.personid
                WHERE p.farm = @userFarm
                UNION ALL
                SELECT 1 AS TYPE, lp.companyname AS name, lp.cnpj AS document, p.ie, lp.personid AS id FROM legal_person lp
                JOIN person p ON p.id = lp.personid
                WHERE p.farm = @userFarm
        ) AS p
		ORDER BY p.name
		LIMIT @limit OFFSET @offset
	`

	rows, queryErr := bm.pool.Query(context.Background(), stmt, pgx.NamedArgs{"userFarm": farm, "limit": limit, "offset": offset})
	if queryErr != nil {
		return []entity_public.PersonDisplay{}, 0, &model_error.ModelError{Message: queryErr.Error(), IsServerErr: true}
	}

	people, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.PersonDisplay])

	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return []entity_public.PersonDisplay{}, total, nil
		}
		return []entity_public.PersonDisplay{}, 0, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return people, total, nil
}

func (bm *personModel) CnpjExistsInFarm(cnpj string, farmId uint32) (bool, *model_error.ModelError) {
	var exists bool
	err := bm.pool.QueryRow(context.Background(), `
		SELECT true FROM legal_person lp
		JOIN person p ON lp.personId = p.id
		WHERE lp.cnpj = $1 AND p.farm = $2
	`, cnpj, farmId).Scan(&exists)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, &model_error.ModelError{Message: "Error checking for CNPJ", IsServerErr: true}
	}

	return exists, nil
}

func (bm *personModel) CpfExistsInFarm(cpf string, farmId uint32) (bool, *model_error.ModelError) {
	var exists bool
	err := bm.pool.QueryRow(context.Background(), `
		SELECT true FROM natural_person np
		JOIN person p ON np.personId = p.id
		WHERE np.cpf = $1 AND p.farm = $2
	`, cpf, farmId).Scan(&exists)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, &model_error.ModelError{Message: "Error checking for CPF", IsServerErr: true}
	}

	return exists, nil
}

func (bm *personModel) GetHumidityDiscount(person *uint32, farm uint32) (decimal.Decimal, *model_error.ModelError) {
	var discountModifier decimal.Decimal
	var err error
	if person != nil {
		err = bm.pool.QueryRow(context.Background(), `
		SELECT pc.humidity_discount FROM person_config pc WHERE pc.person_id = @person
		`, pgx.NamedArgs{"person": person}).Scan(&discountModifier)
	} else {
		err = bm.pool.QueryRow(context.Background(), `
		SELECT COALESCE(fc.humidity_discount, 1.15) AS humidity_discount FROM farm_config fc WHERE fc.farm_id = @farm
		`, pgx.NamedArgs{"farm": farm}).Scan(&discountModifier)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			defaultDisc := "1.15"
			if person != nil {
				defaultDisc = "1.7"
			}
			discountModifier, newDecErr := decimal.NewFromString(defaultDisc)
			if newDecErr != nil {
				return discountModifier, &model_error.ModelError{Message: newDecErr.Error()}
			}
			return discountModifier, nil

		}
		return discountModifier, &model_error.ModelError{Message: err.Error()}
	}
	return discountModifier, nil
}

func (pm *personModel) GetLegalPersonById(id uint32) (entity_public.LegalPerson, *model_error.ModelError) {
	var legalPerson entity_public.LegalPerson

	scanErr := pm.pool.QueryRow(context.Background(), `
		SELECT lp.id, lp.companyname, lp.fantasyname, lp.cnpj FROM person p
			JOIN legal_person lp ON lp.personid = p.id
			WHERE p.id = @id
		`,
		pgx.NamedArgs{"id": id}).Scan(&legalPerson.Id, &legalPerson.CompanyName, &legalPerson.FantasyName, &legalPerson.Cnpj)

	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return legalPerson, &model_error.ModelError{Message: "Pessoa não encontrada"}
		}
		return legalPerson, &model_error.ModelError{Message: scanErr.Error()}
	}

	base, err := pm.getBasePerson(id)
	if err != nil {
		return legalPerson, err
	}

	legalPerson.Address = base.Address
	legalPerson.Person = base.Person
	return legalPerson, nil
}

func (pm *personModel) GetNaturalPersonById(id uint32) (entity_public.NaturalPerson, *model_error.ModelError) {
	var naturalPerson entity_public.NaturalPerson

	scanErr := pm.pool.QueryRow(context.Background(), `
		SELECT np.id, np.name, np.cpf FROM person p
			JOIN natural_person np ON np.personid = p.id
			WHERE p.id = @id
		`,
		pgx.NamedArgs{"id": id}).Scan(&naturalPerson.Id, &naturalPerson.Name, &naturalPerson.Cpf)

	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return naturalPerson, &model_error.ModelError{Message: "Pessoa não encontrada"}
		}
		return naturalPerson, &model_error.ModelError{Message: scanErr.Error()}
	}

	base, err := pm.getBasePerson(id)
	if err != nil {
		return naturalPerson, err
	}

	naturalPerson.Address = base.Address
	naturalPerson.Person = base.Person
	return naturalPerson, nil
}

func (bm *personModel) UpdateNaturalPerson(bp entity_public.NaturalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	row, queryErr := bm.pool.Query(context.Background(), `
			SELECT * FROM update_get_natural_person(@name, @cpf, @ie, @id, @farm, @humidityDiscount, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone)
		`,
		pgx.NamedArgs{
			"id":               bp.Person.Id,
			"ie":               bp.Person.Ie,
			"cpf":              bp.Cpf,
			"name":             bp.Name,
			"farm":             bp.Person.Farm,
			"humidityDiscount": bp.Person.HumidityDiscount,
			"street":           bp.Street,
			"cep":              bp.Cep,
			"number":           bp.Number,
			"neighborhood":     bp.Neighborhood,
			"city":             bp.City,
			"state":            bp.State,
			"complement":       bp.Complement,
			"email":            bp.Email,
			"phone":            bp.PhoneNumber,
		})
	if queryErr != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	person, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.PersonDisplay])
	if collectErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(collectErr, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var message string = "Esta pessoa já existe"
				if strings.Contains(collectErr.Error(), "person_ie") {
					message = "Inscrição Estadual em uso"
				}
				if strings.Contains(collectErr.Error(), "natural_person_cpf") {
					message = "CPF em uso"
				}
				return entity_public.PersonDisplay{}, &model_error.ModelError{Message: message}
			}
		}
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return person, nil
}

func (bm *personModel) UpdateLegalPerson(bc entity_public.LegalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	row, queryErr := bm.pool.Query(
		context.Background(),
		`SELECT * FROM update_get_legal_person(@companyName, @cnpj, @ie, @id, @fantasyName, @farm, @humidityDiscount, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone)`,
		pgx.NamedArgs{
			"id":               bc.Person.Id,
			"ie":               bc.Person.Ie,
			"cnpj":             bc.Cnpj,
			"fantasyName":      bc.FantasyName,
			"farm":             bc.Person.Farm,
			"companyName":      bc.CompanyName,
			"humidityDiscount": bc.Person.HumidityDiscount,
			"street":           bc.Street,
			"cep":              bc.Cep,
			"number":           bc.Number,
			"neighborhood":     bc.Neighborhood,
			"city":             bc.City,
			"state":            bc.State,
			"complement":       bc.Complement,
			"email":            bc.Email,
			"phone":            bc.PhoneNumber,
		})

	if queryErr != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	person, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.PersonDisplay])
	if collectErr != nil {
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return person, nil
}
