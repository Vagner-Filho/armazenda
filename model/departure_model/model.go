package departure_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type departureModel struct {
	pool *pgxpool.Pool
}

var departureModelImpl *departureModel

func InitDepartureModel(pool *pgxpool.Pool) (*departureModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if departureModelImpl == nil {
		departureModelImpl = &departureModel{
			pool: pool,
		}
	}

	return departureModelImpl, nil
}

func GetDepartureModel() *departureModel {
	if departureModelImpl == nil {
		panic("\ndeparture model hasnt been initialized\n")
	}
	return departureModelImpl
}

var availableDepartureFilters = map[string]func(df entity_public.DepartureFilter) string{
	"DepartureDateMin": func(df entity_public.DepartureFilter) string {
		return fmt.Sprintf("d.departureDate >= '%v'", df.DepartureDateMin.Format(utils.DBTimeWithoutTimeZone))
	},
	"DepartureDateMax": func(df entity_public.DepartureFilter) string {
		return fmt.Sprintf("d.departureDate <= '%v'", df.DepartureDateMax.Format(utils.DBTimeWithoutTimeZone))
	},
	"VehiclePlate": func(df entity_public.DepartureFilter) string {
		return fmt.Sprintf("d.vehicle = '%s'", df.VehiclePlate)
	},
	"Product": func(df entity_public.DepartureFilter) string {
		return "p.id = " + strconv.FormatInt(int64(df.Product), 10)
	},
	"WeightMin": func(df entity_public.DepartureFilter) string {
		return "d.netWeight >= " + strconv.FormatFloat(df.NetWeightMin, 'f', -1, 64)
	},
	"WeightMax": func(df entity_public.DepartureFilter) string {
		return "d.netWeight <= " + strconv.FormatFloat(df.NetWeightMax, 'f', -1, 64)
	},
	"Person": func(df entity_public.DepartureFilter) string {
		if df.Person == "NULL" {
			return "dr.person_id IS NULL"
		}
		return "dr.person_id = " + df.Person
	},
}

func (dm *departureModel) FilterDepartures(df entity_public.DepartureFilter, page int, farm uint32) ([]entity_public.DisplayDeparture, int, error) {
	filters := df.GetFilters(availableDepartureFilters)
	pageSize := 10
	offset := (page - 1) * pageSize

	whereClauses := []string{"id.departure_id IS NULL"}
	for _, filter := range filters {
		whereClauses = append(whereClauses, filter(df))
	}
	whereCondition := strings.Join(whereClauses, " AND ")

	countQuery := `SELECT COUNT(*)
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN inactive_departure id ON id.departure_id = d.id
			LEFT JOIN departure_recipient dr ON dr.departure_id = d.id
			JOIN vehicle v ON v.id = d.id
			WHERE ` + whereCondition + " AND d.farm = @farm"

	var totalDepartures int
	countRow := dm.pool.QueryRow(context.Background(), countQuery, pgx.NamedArgs{"farm": farm})
	if err := countRow.Scan(&totalDepartures); err != nil {
		return nil, 0, err
	}

	if totalDepartures == 0 {
		return []entity_public.DisplayDeparture{}, 0, nil
	}

	query := `SELECT d.id, p.name, v.plate, d.departureDate, d.netWeight 
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN inactive_departure id ON id.departure_id = d.id
			LEFT JOIN departure_recipient dr ON dr.departure_id = d.id
			JOIN vehicle v ON v.id = d.id
			WHERE ` + whereCondition + `
			AND d.farm = @farm
			ORDER BY d.departureDate DESC
			LIMIT @pageSize OFFSET @offset`

	rows, queryErr := dm.pool.Query(context.Background(), query, pgx.NamedArgs{"pageSize": pageSize, "offset": offset, "farm": farm})
	if queryErr != nil {
		return []entity_public.DisplayDeparture{}, 0, queryErr
	}

	departures, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		return []entity_public.DisplayDeparture{}, 0, collectErr
	}

	return departures, totalDepartures, nil
}

func (dm *departureModel) GetDisplayDepartures(farm uint32, page int) ([]entity_public.DisplayDeparture, int, *model_error.ModelError) {
	pageSize := 10
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM departure d
		WHERE d.id NOT IN (SELECT departure_id FROM inactive_departure)
		AND d.farm = @userFarm
	`
	var totalDepartures int
	countRow := dm.pool.QueryRow(context.Background(), countQuery, pgx.NamedArgs{"userFarm": farm})
	if err := countRow.Scan(&totalDepartures); err != nil {
		fmt.Printf("\n countErr: %v\n", err.Error())
		return nil, 0, &model_error.ModelError{Message: err.Error()}
	}

	rows, queryErr := dm.pool.Query(context.Background(), `
		SELECT d.id, p.name, v.plate, d.departureDate, d.netWeight, COALESCE(origin_union.name, 'Própria')
		FROM departure d
		JOIN crop c ON d.crop = c.id
		JOIN product p ON c.product = p.id
		JOIN vehicle v ON v.id = d.vehicle
		LEFT JOIN departure_origin dor ON dor.departure_id = d.id
		LEFT JOIN (
			SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name FROM legal_person lp
			UNION ALL
			SELECT np.personid, np.name FROM natural_person np
		) origin_union ON origin_union.personid = dor.person_id
		WHERE d.id NOT IN (SELECT departure_id FROM inactive_departure)
		AND d.farm = @userFarm
		ORDER BY d.departureDate DESC
		LIMIT @pageSize OFFSET @offset
	`, pgx.NamedArgs{"userFarm": farm, "pageSize": pageSize, "offset": offset})

	if queryErr != nil {
		return []entity_public.DisplayDeparture{}, 0, &model_error.ModelError{Message: queryErr.Error()}
	}

	departures, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		return []entity_public.DisplayDeparture{}, 0, &model_error.ModelError{Message: collectErr.Error()}
	}

	return departures, totalDepartures, nil
}

func (dm *departureModel) GetDeparture(id uint32) (entity_public.Departure, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(context.Background(), `
		SELECT d.*, dr.person_id, da.humidity, da.damage, da.impurity, dor.person_id FROM departure d
		LEFT JOIN departure_recipient dr ON dr.departure_id = d.id
		LEFT JOIN departure_analysis da ON da.departure_id = d.id
		LEFT JOIN departure_origin dor ON dor.departure_id = d.id
		WHERE d.id = @id
	`, pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.Departure{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	departure, collectErr := pgx.CollectExactlyOneRow(row, pgx.RowToStructByPos[entity_public.Departure])
	if collectErr != nil {
		return entity_public.Departure{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return departure, nil
}

func (dm *departureModel) AddDeparture(d entity_public.Departure) (entity_public.DisplayDeparture, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(context.Background(), `
		SELECT * FROM add_get_departure(@crop, @recipientId, @vehicle, @departureDate, @farm, @tare, @grossWeight, @netWeight, @humidity, @damage, @impurity, @originId)
		`, pgx.NamedArgs{
		"crop":          d.Crop,
		"recipientId":   d.Recipient,
		"vehicle":       d.Vehicle,
		"departureDate": d.DepartureDate,
		"farm":          d.Farm,
		"tare":          d.Tare,
		"grossWeight":   d.GrossWeight,
		"netWeight":     d.NetWeight,
		"humidity":      d.Humidity,
		"damage":        d.Damage,
		"impurity":      d.Impurity,
		"originId":      d.Origin,
	})
	if queryErr != nil {
		fmt.Printf("\nadd departure query err:\n%v", queryErr.Error())
	}

	departure, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		fmt.Printf("\nadd departure collect err:\n%v", collectErr.Error())
	}

	return departure, nil
}

func (dm *departureModel) PutDeparture(d entity_public.Departure) (entity_public.DisplayDeparture, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(
		context.Background(),
		"SELECT * FROM update_get_departure(@crop, @recipientId, @departureId, @vehicle, @departureDate, @grossWeight, @tare, @netWeight, @humidity, @damage, @impurity, @originId)",
		pgx.NamedArgs{
			"departureId":   d.Id,
			"crop":          d.Crop,
			"vehicle":       d.Vehicle,
			"grossWeight":   d.GrossWeight,
			"tare":          d.Tare,
			"netWeight":     d.NetWeight,
			"recipientId":   d.Recipient,
			"departureDate": d.DepartureDate,
			"humidity":      d.Humidity,
			"damage":        d.Damage,
			"impurity":      d.Impurity,
			"originId":      d.Origin,
		})
	if queryErr != nil {
		return entity_public.DisplayDeparture{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	departure, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		return entity_public.DisplayDeparture{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return departure, nil
}

func (dm *departureModel) DeleteDeparture(id uint32) *model_error.ModelError {
	_, err := dm.pool.Exec(context.Background(),
		"INSERT INTO inactive_departure (departure_id) VALUES (@departureId)",
		pgx.NamedArgs{"departureId": id},
	)

	if err != nil {
	}

	return nil
}

func (dm *departureModel) GetDeparturePdf(id uint32) (entity_public.DeparturePdf, *model_error.ModelError) {

	rows, queryErr := dm.pool.Query(
		context.Background(),
		`SELECT
			d.id,
			c.name AS safra,
			v.plate,
			d.grossweight,
			d.tare,
			d.netweight,
			d.departuredate,
			f.inscricao_estadual,
			prod.name AS produto,
			COALESCE(recipient_union.name, 'Próprio') AS person_name,
			COALESCE(recipient_union.document, f.inscricao_estadual) AS document,
			fc.name as farm_name,
			fa.street as farm_street,
			fa.cep as farm_cep,
			fa.number as farm_number,
			fa.neighborhood as farm_neighborhood,
			fa.city as farm_city,
			fa.state as farm_state,
			da.humidity,
			da.damage,
			da.impurity,
			fc.storage_name,
			COALESCE(origin_union.name, 'Própria') AS origin_name,
			COALESCE(origin_union.document, f.inscricao_estadual) AS origin_document
		FROM departure d
		JOIN crop c ON c.id = d.crop
		JOIN product prod ON prod.id = c.product
		JOIN farm f ON f.id = d.farm
		JOIN vehicle v ON v.id = d.vehicle
		LEFT JOIN farm_config fc ON fc.farm_id = f.id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN departure_recipient dr ON dr.departure_id = d.id
		LEFT JOIN (
			SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name, lp.cnpj AS document FROM legal_person lp
			UNION ALL
			SELECT np.personid, np.name, np.cpf AS document FROM natural_person np
		) recipient_union ON recipient_union.personid = dr.person_id
		LEFT JOIN departure_analysis da ON da.departure_id = d.id
		LEFT JOIN departure_origin dor ON dor.departure_id = d.id
		LEFT JOIN (
			SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name, lp.cnpj AS document FROM legal_person lp
			UNION ALL
			SELECT np.personid, np.name, np.cpf AS document FROM natural_person np
		) origin_union ON origin_union.personid = dor.person_id
		WHERE d.id = @id;`,
		pgx.NamedArgs{"id": id},
	)

	if queryErr != nil {
		return entity_public.DeparturePdf{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	departure, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.DeparturePdf])
	if collectErr != nil {
		return entity_public.DeparturePdf{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return departure, nil
}

func (dm *departureModel) CreateDepartureDraft(d entity_public.DepartureDraft) (entity_public.DisplayDepartureDraft, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(context.Background(), `
		SELECT * FROM add_get_departure_draft(@name, @recipient, @crop, @vehicle, @tare, @farm, @origin)
		`, pgx.NamedArgs{
		"name":      d.Name,
		"recipient": d.Recipient,
		"crop":      d.Crop,
		"vehicle":   d.Vehicle,
		"tare":      d.Tare,
		"farm":      d.Farm,
		"origin":    d.Origin,
	})
	if queryErr != nil {
		return entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	draft, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayDepartureDraft])
	if collectErr != nil {
		return entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return draft, nil
}

func (dm *departureModel) GetDepartureDraft(id uint32) (entity_public.DepartureDraft, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(context.Background(), `
		SELECT dd.id, dd.name, ddo.person_id, dd.crop, dd.vehicle, dd.tare, dd.farm, ddr.person_id FROM departure_draft dd
		LEFT JOIN departure_draft_origin ddo ON ddo.departure_draft_id = dd.id
		LEFT JOIN person p1 ON p1.id = ddo.person_id
		LEFT JOIN departure_draft_recipient ddr ON ddr.departure_draft_id = dd.id
		LEFT JOIN person p2 ON p2.id = ddr.person_id
		WHERE dd.id = @id
	`, pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.DepartureDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	draft, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DepartureDraft])
	if collectErr != nil {
		return entity_public.DepartureDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return draft, nil
}

func (dm *departureModel) GetAllDepartureDrafts(farmId uint32) ([]entity_public.DisplayDepartureDraft, *model_error.ModelError) {

	rows, queryErr := dm.pool.Query(context.Background(), `
		SELECT
			dd.id,
			dd.name,
			COALESCE(np.name, lp.fantasyname, lp.companyname, 'Própria') as person,
			c.name as crop,
			v.plate as vehicle,
			dd.tare
		FROM departure_draft dd
		LEFT JOIN departure_draft_origin ddo ON ddo.departure_draft_id = dd.id
		LEFT JOIN person p ON p.id = ddo.person_id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		LEFT JOIN crop c ON c.id = dd.crop
		LEFT JOIN vehicle v ON v.id = dd.vehicle
		WHERE dd.farm = @farmId
	`, pgx.NamedArgs{"farmId": farmId})
	if queryErr != nil {
		return []entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	drafts, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayDepartureDraft])
	if collectErr != nil {
		return []entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return drafts, nil
}

func (dm *departureModel) UpdateDepartureDraft(d entity_public.DepartureDraft) (entity_public.DisplayDepartureDraft, *model_error.ModelError) {
	row, queryErr := dm.pool.Query(context.Background(), `
		SELECT * FROM update_get_departure_draft(@id, @name, @recipient, @crop, @vehicle, @tare, @farm, @origin)
		`, pgx.NamedArgs{
		"id":        d.Id,
		"name":      d.Name,
		"recipient": d.Recipient,
		"crop":      d.Crop,
		"vehicle":   d.Vehicle,
		"tare":      d.Tare,
		"farm":      d.Farm,
		"origin":    d.Origin,
	})
	if queryErr != nil {
		return entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	draft, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayDepartureDraft])
	if collectErr != nil {
		return entity_public.DisplayDepartureDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return draft, nil
}

func (dm *departureModel) DeleteDepartureDraft(id uint32) *model_error.ModelError {
	_, err := dm.pool.Exec(context.Background(),
		"DELETE FROM departure_draft WHERE id = @id",
		pgx.NamedArgs{"id": id},
	)

	if err != nil {
		return &model_error.ModelError{Message: err.Error()}
	}

	return nil
}
