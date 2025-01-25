package entry_model

import (
	"armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type entryModel struct {
	conn *pgx.Conn
}

var entryModelImpl *entryModel

func InitEntryModel(conn *pgx.Conn) (*entryModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if entryModelImpl == nil {
		entryModelImpl = &entryModel{
			conn: conn,
		}
	}

	return entryModelImpl, nil
}

func GetEntryModel() (*entryModel, error) {
	if entryModelImpl == nil {
		return nil, errors.New("entry model hasnt been initialized")
	}
	return entryModelImpl, nil
}

var GrainMap = make(map[entity_public.Grain]string)

func InitGrainMap() {
	GrainMap[entity_public.Corn] = "Milho"
	GrainMap[entity_public.Soy] = "Soja"
}

func (em *entryModel) GetAllEntriesSimplified() ([]entity_public.SimplifiedEntry, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), "SELECT id, product, field, vehicle, netWeight, arrivalDate FROM entry")
	if queryErr != nil {
		fmt.Printf("\n queryErr: %v\n", queryErr.Error())
		return []entity_public.SimplifiedEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.SimplifiedEntry])
	if collectErr != nil {
		fmt.Printf("\n collectErr: %v\n", collectErr.Error())
		return []entity_public.SimplifiedEntry{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return entries, nil
}

func (em *entryModel) AddEntry(ge entity_public.Entry) (entity_public.Entry, *model_error.ModelError) {
	row, queryErr := em.conn.Query(context.Background(), `INSERT INTO entry (product, field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate) VALUES (@product, @field, @crop, @vehicle, @grossweight, @tare, @netweight, @humidity, @arrivalDate) RETURNING product, field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate`)

	if queryErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil

}

func (em *entryModel) DeleteEntry(id uint32) {
}

func (em *entryModel) GetEntry(id uint32) {

}

func (em *entryModel) PutEntry(ge entity_public.Entry) {

}

var availableEntryFilters = map[string]func(e entity_public.Entry, ef entity_public.EntryFilter) bool{
	"ArrivalDateMin": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		entryArrival, entryDateError := time.Parse(utils.TimeLayout, e.ArrivalDate)
		arrivalFrom, entryFilterDateError := time.Parse(utils.TimeLayout, ef.ArrivalDateMin)
		if entryDateError != nil || entryFilterDateError != nil {
			return false
		}
		return arrivalFrom.Before(entryArrival)
	},
	"ArrivalDateMax": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		entryArrival, entryDateError := time.Parse(utils.TimeLayout, e.ArrivalDate)
		arrivalTo, entryFilterDateError := time.Parse(utils.TimeLayout, ef.ArrivalDateMax)
		if entryDateError != nil || entryFilterDateError != nil {
			return false
		}
		return arrivalTo.After(entryArrival)
	},
	"VehiclePlate": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Vehicle == ef.VehiclePlate
	},
	"Product": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Product == ef.Product
	},
	"Field": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Field == ef.Field
	},
	"NetWeightMin": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.NetWeight >= ef.NetWeightMin
	},
	"NetWeightMax": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.NetWeight <= ef.NetWeightMax
	},
	"Crop": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Crop == ef.Crop
	},
}

func FilterEntries(filter entity_public.EntryFilter) ([]entity_public.Entry, error) {
	var filteredEntries []entity_public.Entry

	bomdia := []entity_public.Entry{}
	filters := filter.GetFilters(availableEntryFilters)
	for _, entry := range bomdia {
		include := true
		for f := range filters {
			fff := filters[f]

			if fff == nil {
				continue
			}

			include = fff(entry, filter)
			if !include {
				break
			}
		}
		if include {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	return filteredEntries, nil
}
