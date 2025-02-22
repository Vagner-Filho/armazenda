package departure_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"time"

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

var availableDepartureFilters = map[string]func(e entity_public.Departure, ef entity_public.DepartureFilter) bool{
	"DepartureDateMin": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		departureMin, departureFilterDateError := time.Parse(utils.TimeLayout, df.DepartureDateMin)
		if departureFilterDateError != nil {
			return false
		}
		return departureMin.Before(d.DepartureDate)
	},
	"DepartureDateMax": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		departureMax, departureFilterDateError := time.Parse(utils.TimeLayout, df.DepartureDateMax)
		if departureFilterDateError != nil {
			return false
		}
		return departureMax.After(d.DepartureDate)
	},
	"VehiclePlate": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.VehiclePlate == df.VehiclePlate
	},
	"Product": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return false
		//return d.Product == df.Product
	},
	"WeightMin": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Weight >= df.WeightMin
	},
	"WeightMax": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Weight <= df.WeightMax
	},
	"Buyer": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Buyer == df.Buyer
	},
}

func FilterDepartures(filter entity_public.DepartureFilter) ([]entity_public.Departure, error) {
	var filteredDepartures []entity_public.Departure

	filters := filter.GetFilters(availableDepartureFilters)

	departures := []entity_public.Departure{}
	for _, departure := range departures {
		include := true
		for f := range filters {
			fff := filters[f]

			if fff == nil {
				continue
			}

			include = fff(departure, filter)
			if !include {
				break
			}
		}
		if include {
			filteredDepartures = append(filteredDepartures, departure)
		}
	}
	return filteredDepartures, nil
}

func (dm *departureModel) GetDepartures() ([]entity_public.Departure, error) {
	rows, queryErr := dm.conn.Query(context.Background(), `
		SELECT d.id, p.name, d.vehicle, d.weight, d.departureDate
		FROM departure d
		JOIN product p ON d.crop = p.id
	`)
	if queryErr != nil {
		return []entity_public.Departure{}, queryErr
	}

	departures, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Departure])
	if collectErr != nil {
		return []entity_public.Departure{}, collectErr
	}

	return departures, nil
}

func (dm *departureModel) GetDisplayDepartures() ([]entity_public.DisplayDeparture, *model_error.ModelError) {
	rows, queryErr := dm.conn.Query(context.Background(), `
		SELECT d.id, p.name, d.vehicle, d.weight, d.departureDate
		FROM departure d
		JOIN crop c ON d.crop = c.id
		JOIN product p ON c.product = p.id
		WHERE d.id NOT IN (SELECT departure_id FROM inactive_departure)
	`)

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
		SELECT * FROM add_get_departure(@crop, @buyer, @vehicle, @weight, @departureDate)
		`, pgx.NamedArgs{
		"crop":          d.Crop,
		"buyer":         d.Buyer,
		"vehicle":       d.VehiclePlate,
		"weight":        d.Weight,
		"departureDate": d.DepartureDate,
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
