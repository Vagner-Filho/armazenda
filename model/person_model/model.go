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
)

type personModel struct {
	conn *pgx.Conn
}

var personModelImpl *personModel

func InitPersonModel(conn *pgx.Conn) (*personModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if personModelImpl == nil {
		personModelImpl = &personModel{
			conn: conn,
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

func (bm *personModel) AddLegalPerson(bc entity_public.LegalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	row, queryErr := bm.conn.Query(
		context.Background(),
		`SELECT * FROM add_get_legal_person(@companyName, @cnpj, @ie, @fantasyName, @farm, @humidityDiscount, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone)`,
		pgx.NamedArgs{
			"ie":               bc.InscricaoEstadual,
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
		model_error.Logger(bm.conn, queryErr.Error())
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	person, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.PersonDisplay])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return person, nil
}

func (bm *personModel) AddNaturalPerson(bp entity_public.NaturalPerson) (entity_public.PersonDisplay, *model_error.ModelError) {
	row, queryErr := bm.conn.Query(context.Background(), `
			SELECT * FROM add_get_natural_person(@name, @cpf, @ie, @farm, @humidityDiscount, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone)
		`,
		pgx.NamedArgs{
			"ie":               bp.InscricaoEstadual,
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
		model_error.Logger(bm.conn, queryErr.Error())
		return entity_public.PersonDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	person, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.PersonDisplay])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
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

func (bm *personModel) GetPeopleByFarm(farm uint32) ([]entity_public.PersonOption, *model_error.ModelError) {
	rows, queryErr := bm.conn.Query(context.Background(), `
		SELECT id, name, humidity_discount FROM (
			SELECT p.id, lp.companyname AS name, COALESCE(pc.humidity_discount, 1.7) as humidity_discount
			FROM person p
			JOIN legal_person lp ON p.id = lp.personid
			LEFT JOIN person_config pc ON p.id = pc.person_id
			WHERE p.farm = @userFarm
			UNION
			SELECT p.id, np.name, COALESCE(pc.humidity_discount, 1.7) as humidity_discount
			FROM person p
			JOIN natural_person np ON p.id = np.personid
			LEFT JOIN person_config pc ON p.id = pc.person_id
			WHERE p.farm = @userFarm
			UNION
			SELECT NULL, 'Própria', COALESCE((SELECT humidity_discount FROM farm_config WHERE farm_id = @userFarm), 1.15) as humidity_discount
		) AS result
		ORDER BY CASE WHEN id IS NULL THEN 0 ELSE 1 END, name
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		model_error.Logger(bm.conn, queryErr.Error())
		return []entity_public.PersonOption{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	people, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.PersonOption])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
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
	err := bm.conn.QueryRow(context.Background(), countStmt, pgx.NamedArgs{"userFarm": farm}).Scan(&total)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []entity_public.PersonDisplay{}, 0, nil
		}
		model_error.Logger(bm.conn, err.Error())
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

	rows, queryErr := bm.conn.Query(context.Background(), stmt, pgx.NamedArgs{"userFarm": farm, "limit": limit, "offset": offset})
	if queryErr != nil {
		model_error.Logger(bm.conn, queryErr.Error())
		return []entity_public.PersonDisplay{}, 0, &model_error.ModelError{Message: queryErr.Error(), IsServerErr: true}
	}

	people, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.PersonDisplay])

	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return []entity_public.PersonDisplay{}, total, nil
		}
		model_error.Logger(bm.conn, collectErr.Error())
		return []entity_public.PersonDisplay{}, 0, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return people, total, nil
}

func (bm *personModel) CnpjExistsInFarm(cnpj string, farmId uint32) (bool, *model_error.ModelError) {
	var exists bool
	err := bm.conn.QueryRow(context.Background(), `
		SELECT true FROM legal_person lp
		JOIN person p ON lp.personId = p.id
		WHERE lp.cnpj = $1 AND p.farm = $2
	`, cnpj, farmId).Scan(&exists)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		model_error.Logger(bm.conn, err.Error())
		return false, &model_error.ModelError{Message: "Error checking for CNPJ", IsServerErr: true}
	}

	return exists, nil
}

func (bm *personModel) CpfExistsInFarm(cpf string, farmId uint32) (bool, *model_error.ModelError) {
	var exists bool
	err := bm.conn.QueryRow(context.Background(), `
		SELECT true FROM natural_person np
		JOIN person p ON np.personId = p.id
		WHERE np.cpf = $1 AND p.farm = $2
	`, cpf, farmId).Scan(&exists)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		model_error.Logger(bm.conn, err.Error())
		return false, &model_error.ModelError{Message: "Error checking for CPF", IsServerErr: true}
	}

	return exists, nil
}