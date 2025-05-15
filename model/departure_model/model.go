package departure_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type departureModel struct {
	conn *pgx.Conn
}

var departureModelImpl *departureModel

func InitDepartureModel(conn *pgx.Conn) (*departureModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if departureModelImpl == nil {
		departureModelImpl = &departureModel{
			conn: conn,
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
	"Buyer": func(df entity_public.DepartureFilter) string {
		return "d.Buyer = " + df.Buyer
	},
}

func (dm *departureModel) FilterDepartures(df entity_public.DepartureFilter) ([]entity_public.DisplayDeparture, error) {
	filters := df.GetFilters(availableDepartureFilters)

	stmt := `SELECT d.id, p.name, d.vehicle, d.departureDate, d.netWeight 
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT OUTER JOIN inactive_departure id ON id.departure_id = d.id
			WHERE id.departure_id IS NULL`

	for _, filter := range filters {
		stmt += "\nAND " + filter(df)
	}

	rows, queryErr := dm.conn.Query(context.Background(), stmt)
	if queryErr != nil {
		model_error.Logger(dm.conn, queryErr.Error())
		return []entity_public.DisplayDeparture{}, queryErr
	}

	departures, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		model_error.Logger(dm.conn, collectErr.Error())
		return []entity_public.DisplayDeparture{}, collectErr
	}

	return departures, nil
}

func (dm *departureModel) GetDisplayDepartures(farm uint32) ([]entity_public.DisplayDeparture, *model_error.ModelError) {
	rows, queryErr := dm.conn.Query(context.Background(), `
		SELECT d.id, p.name, d.vehicle, d.departureDate, d.netWeight
		FROM departure d
		JOIN crop c ON d.crop = c.id
		JOIN product p ON c.product = p.id
		WHERE d.id NOT IN (SELECT departure_id FROM inactive_departure)
		AND d.farm = @userFarm
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		return []entity_public.DisplayDeparture{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	departures, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		return []entity_public.DisplayDeparture{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return departures, nil
}

func (dm *departureModel) GetDeparture(id uint32) (entity_public.Departure, *model_error.ModelError) {
	row, queryErr := dm.conn.Query(context.Background(), `
		SELECT d.*, db.buyerid FROM departure d
		JOIN departurebuyer db ON db.departureid = d.id
		WHERE d.id = @id
	`, pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.Departure{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	departure, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.Departure])
	if collectErr != nil {
		return entity_public.Departure{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return departure, nil
}

func (dm *departureModel) AddDeparture(d entity_public.Departure) (entity_public.DisplayDeparture, *model_error.ModelError) {
	row, queryErr := dm.conn.Query(context.Background(), `
		SELECT * FROM add_get_departure(@crop, @buyer, @vehicle, @departureDate, @farm, @tare, @grossWeight, @netWeight)
		`, pgx.NamedArgs{
		"crop":          d.Crop,
		"buyer":         d.Buyer,
		"vehicle":       d.VehiclePlate,
		"departureDate": d.DepartureDate,
		"farm":          d.Farm,
		"tare":          d.Tare,
		"grossWeight":   d.GrossWeight,
		"netWeight":     d.NetWeight,
	})
	if queryErr != nil {
		model_error.Logger(dm.conn, queryErr.Error())
		fmt.Printf("\nadd departure query err:\n%v", queryErr.Error())
	}

	departure, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayDeparture])
	if collectErr != nil {
		model_error.Logger(dm.conn, collectErr.Error())
		fmt.Printf("\nadd departure collect err:\n%v", collectErr.Error())
	}

	return departure, nil
}

func PutDeparture(d entity_public.Departure) (entity_public.Departure, bool) {
	return d, false
}

func (dm *departureModel) DeleteDeparture(id uint32) *model_error.ModelError {
	_, err := dm.conn.Exec(context.Background(),
		"INSERT INTO inactive_departure (departure_id) VALUES (@departureId)",
		pgx.NamedArgs{"departureId": id},
	)

	if err != nil {
		model_error.Logger(dm.conn, err.Error())
	}

	return nil
}
