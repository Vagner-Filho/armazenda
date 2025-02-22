package entry_model

import (
	"armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
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

func GetEntryModel() *entryModel {
	if entryModelImpl == nil {
		panic("entry model hasnt been initialized")
	}
	return entryModelImpl
}

func (em *entryModel) GetAllEntriesSimplified() ([]entity_public.SimplifiedEntry, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), `
		SELECT e.id, p.name, f.name, e.vehicle, e.netweight, e.arrivaldate
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			WHERE ie.entry_id IS NULL
			ORDER BY c.startdate DESC
		`)
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
	row, queryErr := em.conn.Query(context.Background(), `
		INSERT INTO entry (field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate)
		VALUES (@field, @crop, @vehicle, @grossweight, @tare, @netweight, @humidity, @arrivalDate)
		RETURNING id, field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate
		`, pgx.NamedArgs{"field": ge.Field, "crop": ge.Crop, "vehicle": ge.Vehicle, "grossweight": ge.GrossWeight, "tare": ge.Tare, "netweight": ge.NetWeight, "humidity": ge.Humidity, "arrivalDate": ge.ArrivalDate})

	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *entryModel) DeleteEntry(id uint32) error {
	_, err := em.conn.Exec(context.Background(), "INSERT INTO inactive_entry (entry_id) VALUES (@entryId)", pgx.NamedArgs{"entryId": id})

	if err != nil {
		model_error.Logger(em.conn, err.Error())
	}

	return nil
}

func (em *entryModel) GetEntry(id uint32) (entity_public.Entry, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), "SELECT * FROM entry WHERE id = @id", pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	fmt.Printf("\n%+v\n", entry)
	return entry, nil
}

func (em *entryModel) PutEntry(ge entity_public.Entry) (entity_public.Entry, *model_error.ModelError) {
	row, queryErr := em.conn.Query(context.Background(), `
		UPDATE entry SET
			(field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate)
		VALUES (@field, @crop, @vehicle, @grossweight, @tare, @netweight, @humidity, @arrivalDate)
		WHERE id = @id
		RETURNING id, field, crop, vehicle, grossweight, tare, netweight, humidity, arrivalDate
		`, pgx.NamedArgs{"id": ge.Id, "field": ge.Field, "crop": ge.Crop, "vehicle": ge.Vehicle, "grossweight": ge.GrossWeight, "tare": ge.Tare, "netweight": ge.NetWeight, "humidity": ge.Humidity, "arrivalDate": ge.ArrivalDate})

	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

var availableEntryFilters = map[string]func(e entity_public.Entry, ef entity_public.EntryFilter) string{
	"ArrivalDateMin": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		arrivalFrom, entryFilterDateError := time.Parse(utils.TimeLayout, ef.ArrivalDateMin)
		if entryFilterDateError != nil {
			return false
		}
		return arrivalFrom.Before(e.ArrivalDate)
	},
	"ArrivalDateMax": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		arrivalTo, entryFilterDateError := time.Parse(utils.TimeLayout, ef.ArrivalDateMax)
		if entryFilterDateError != nil {
			return false
		}
		return arrivalTo.After(e.ArrivalDate)
	},
	"Vehicle": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Vehicle == ef.Vehicle
	},
	"Product": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		//return e.Product == ef.Product
		return false
	},
	"Field": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.Field == ef.Field
	},
	"NetWeightMin": func(e entity_public.Entry, ef entity_public.EntryFilter) bool {
		return e.NetWeight >= ef.NetWeightMin
	},
	"NetWeightMax": func(e entity_public.Entry, ef entity_public.EntryFilter) string {
		return "e.netweight <= " + strconv.FormatFloat(ef.NetWeightMax, 'f', -1, 64)
	},
	"Crop": func(e entity_public.Entry, ef entity_public.EntryFilter) string {
		return "c.id = " + string(ef.Crop)
	},
}

func (em *entryModel) FilterEntries(filter entity_public.EntryFilter) ([]entity_public.SimplifiedEntry, error) {
	//filters := filter.GetFilters(availableEntryFilters)
	//rows, queryErr := em.conn.Query(context.Background(), `

	//`)

	stmt := `SELECT e.id, p.name, f.name, e.vehicle, e.netweight, e.arrivaldate
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			WHERE ie.entry_id IS NULL
			ORDER BY c.startdate DESC
		`
	return []entity_public.SimplifiedEntry{}, nil
}
