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

func (em *entryModel) GetDisplayEntriesByFarm(farm uint32) ([]entity_public.DisplayEntry, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), `
		SELECT e.id, p.name, f.name, e.vehicle, e.netweight, e.arrivaldate, e.farm
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			WHERE ie.entry_id IS NULL AND e.farm = @userFarm
			ORDER BY c.startdate DESC
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		fmt.Printf("\n queryErr: %v\n", queryErr.Error())
		return []entity_public.DisplayEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		fmt.Printf("\n collectErr: %v\n", collectErr.Error())
		return []entity_public.DisplayEntry{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return entries, nil
}

func (em *entryModel) GetEntryDraftsByFarm(farm uint32) ([]entity_public.DisplayEntryDraft, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), `
		SELECT ed.id, ed.name, f.name, c.name, ed.vehicle, ed.tare, COALESCE(np.name, lp.fantasyname, lp.companyname, 'Própria') AS origin
			FROM entry_draft ed
			JOIN field f ON ed.field = f.id
			JOIN crop c ON ed.crop = c.id
			LEFT JOIN entry_draft_origin edo ON edo.entry_draft_id = ed.id
			LEFT JOIN person p ON p.id = edo.person_id
			LEFT JOIN natural_person np ON np.personid = p.id
			LEFT JOIN legal_person lp ON lp.personid = p.id
			WHERE ed.farm = @userFarm;
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		fmt.Printf("\n queryErr: %v\n", queryErr.Error())
		return []entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	drafts, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntryDraft])
	if collectErr != nil {
		fmt.Printf("\n collectErr: %v\n", collectErr.Error())
		return []entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return drafts, nil
}

func (em *entryModel) AddEntryDraft(ge entity_public.EntryDraft) (entity_public.DisplayEntryDraft, *model_error.ModelError) {
	row, queryErr := em.conn.Query(context.Background(), `
		SELECT * FROM add_get_entry_draft(@name, @field, @crop, @vehicle, @tare, @farm, @origin)
		`, pgx.NamedArgs{
		"name":    ge.Name,
		"field":   ge.Field,
		"crop":    ge.Crop,
		"vehicle": ge.Vehicle,
		"tare":    ge.Tare,
		"farm":    ge.Farm,
		"origin":  ge.Origin,
	})

	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntryDraft])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}
func (em *entryModel) GetEntryDraft(id uint32) (entity_public.EntryDraft, *model_error.ModelError) {
	row, queryErr := em.conn.Query(context.Background(), `
		SELECT ed.*, edo.person_id AS origin
			FROM entry_draft ed
			JOIN field f ON ed.field = f.id
			JOIN crop c ON ed.crop = c.id
			LEFT JOIN entry_draft_origin edo ON edo.entry_draft_id = ed.id
			WHERE ed.id = @id;
		`, pgx.NamedArgs{"id": id})
	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.EntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}
	draft, collectErr := pgx.CollectExactlyOneRow(row, pgx.RowToStructByPos[entity_public.EntryDraft])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return entity_public.EntryDraft{}, &model_error.ModelError{Message: "Rascunho não encontrado"}
		}
		return entity_public.EntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return draft, nil
}

func (em *entryModel) AddEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
	row, queryErr := em.conn.Query(context.Background(), `
		SELECT * FROM add_get_entry(@field, @crop, @grossWeight, @tare, @humidity, @vehicle, @netWeight, @arrivalDate, @farm, @damage, @impurity, @origin)
		`, pgx.NamedArgs{
		"field":       ge.Field,
		"crop":        ge.Crop,
		"vehicle":     ge.Vehicle,
		"grossWeight": ge.GrossWeight,
		"tare":        ge.Tare,
		"netWeight":   ge.NetWeight,
		"humidity":    ge.Humidity,
		"arrivalDate": ge.ArrivalDate,
		"farm":        ge.Farm,
		"damage":      ge.Damage,
		"impurity":    ge.Impurity,
		"origin":      ge.Origin,
	})

	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *entryModel) DeleteEntryDraft(id uint32) error {
	_, err := em.conn.Exec(context.Background(), "DELETE FROM entry_draft WHERE id = @draftId", pgx.NamedArgs{"draftId": id})

	if err != nil {
		model_error.Logger(em.conn, err.Error())
	}

	return nil
}

func (em *entryModel) DeleteEntry(id uint32) error {
	_, err := em.conn.Exec(context.Background(), "INSERT INTO inactive_entry (entry_id) VALUES (@entryId)", pgx.NamedArgs{"entryId": id})

	if err != nil {
		model_error.Logger(em.conn, err.Error())
	}

	return nil
}

func (em *entryModel) GetEntry(id uint32) (entity_public.Entry, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(context.Background(), "SELECT e.id, e.field, e.crop, e.vehicle, e.grossweight, e.tare, e.netweight, ea.humidity, ea.damage, ea.impurity, e.arrivaldate, e.farm, eo.person_id FROM entry e LEFT JOIN entry_analysis ea ON ea.entryid = e.id LEFT JOIN entry_origin eo ON eo.entry_id = e.id WHERE e.id = @id", pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *entryModel) GetEntryPdf(id uint32) (entity_public.EntryPdf, *model_error.ModelError) {
	rows, queryErr := em.conn.Query(
		context.Background(),
		`SELECT
			e.id,
			c.name AS safra,
			e.vehicle,
			e.grossweight,
			e.tare,
			e.netweight,
			ea.humidity,
			ea.damage,
			ea.impurity,
			e.arrivaldate,
			f.inscricao_estadual,
			p.name AS produto,
			fc.name as farm_name,
			fa.street as farm_street,
			fa.cep as farm_cep,
			fa.number as farm_number,
			fa.neighborhood as farm_neighborhood,
			fa.city as farm_city,
			fa.state as farm_state,
			person_union.name as origin
		FROM entry e
		LEFT JOIN entry_analysis ea ON ea.entryid = e.id
		JOIN field fi ON fi.id = e.field
		JOIN crop c ON c.id = e.crop
		JOIN product p ON p.id = c.product
		JOIN farm f ON f.id = e.farm
		LEFT JOIN farm_config fc ON fc.farm_id = f.id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN entry_origin eo ON eo.entry_id = e.id
		LEFT JOIN (
			SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name FROM legal_person lp
			UNION ALL
			SELECT np.personid, np.name FROM natural_person np
		) person_union ON person_union.personid = eo.person_id
	 	WHERE e.id = @id`,
		pgx.NamedArgs{"id": id},
	)

	if queryErr != nil {
		return entity_public.EntryPdf{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.EntryPdf])
	if collectErr != nil {
		return entity_public.EntryPdf{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *entryModel) PutEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
	row, queryErr := em.conn.Query(
		context.Background(),
		`SELECT * FROM update_get_display_entry(@field, @crop, @grossWeight, @tare, @humidity, @id, @vehicle, @netWeight, @arrivalDate, @damage, @impurity)`,
		pgx.NamedArgs{
			"id":          ge.Id,
			"field":       ge.Field,
			"crop":        ge.Crop,
			"vehicle":     ge.Vehicle,
			"grossWeight": ge.GrossWeight,
			"tare":        ge.Tare,
			"netWeight":   ge.NetWeight,
			"humidity":    ge.Humidity,
			"arrivalDate": ge.ArrivalDate,
			"damage":      ge.Damage,
			"impurity":    ge.Impurity,
		})

	if queryErr != nil {
		model_error.Logger(em.conn, queryErr.Error())
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: collectErr.Error()}
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

func (em *entryModel) FilterEntries(ef entity_public.EntryFilter) ([]entity_public.DisplayEntry, error) {
	filters := ef.GetFilters(availableEntryFilters)

	stmt := `SELECT e.id, p.name, f.name, e.vehicle, e.netweight, e.arrivaldate, e.farm
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
		model_error.Logger(em.conn, queryErr.Error())
		return []entity_public.DisplayEntry{}, queryErr
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		model_error.Logger(em.conn, collectErr.Error())
		return []entity_public.DisplayEntry{}, collectErr
	}

	return entries, nil
}
