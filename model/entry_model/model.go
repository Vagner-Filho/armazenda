package entry_model

import (
	"armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"

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

var availableEntryFilters = map[string]func(ef entity_public.EntryFilter) string{
	"ArrivalDateMin": func(ef entity_public.EntryFilter) string {
		return fmt.Sprintf("e.arrivalDate >= '%v'", ef.ArrivalDateMin.Format(utils.DBTimeWithoutTimeZone))
	},
	"ArrivalDateMax": func(ef entity_public.EntryFilter) string {
		return fmt.Sprintf("e.arrivalDate <= '%v'", ef.ArrivalDateMax.Format(utils.DBTimeWithoutTimeZone))
	},
	"Vehicle": func(ef entity_public.EntryFilter) string {
		return fmt.Sprintf("e.vehicle = '%s'", ef.Vehicle)
	},
	"Product": func(ef entity_public.EntryFilter) string {
		return "p.id = " + strconv.FormatInt(int64(ef.Product), 10)
	},
	"Field": func(ef entity_public.EntryFilter) string {
		return "e.field = " + strconv.FormatInt(int64(ef.Field), 10)
	},
	"NetWeightMin": func(ef entity_public.EntryFilter) string {
		return "e.netweight >= " + strconv.FormatFloat(ef.NetWeightMin, 'f', -1, 64)
	},
	"NetWeightMax": func(ef entity_public.EntryFilter) string {
		return "e.netweight <= " + strconv.FormatFloat(ef.NetWeightMax, 'f', -1, 64)
	},
	"Crop": func(ef entity_public.EntryFilter) string {
		return fmt.Sprintf("c.id = %v", ef.Crop)
	},
}

func (em *entryModel) FilterEntries(ef entity_public.EntryFilter) ([]entity_public.SimplifiedEntry, error) {
	filters := ef.GetFilters(availableEntryFilters)

	stmt := `SELECT e.id, p.name, f.name, e.vehicle, e.netweight, e.arrivaldate
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			WHERE ie.entry_id IS NULL`

	for _, filter := range filters {
		stmt += "\nAND " + filter(ef)
	}

	rows, queryErr := em.conn.Query(context.Background(), stmt)
	if queryErr != nil {
		fmt.Print(stmt)
		fmt.Print(queryErr.Error())
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.SimplifiedEntry])
	if collectErr != nil {
		fmt.Print(stmt)
		fmt.Print(collectErr.Error())
	}

	fmt.Printf("\n%v\n", entries)

	return entries, nil
}
