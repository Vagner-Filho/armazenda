package entry_model

import (
	"armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type EntryModel struct {
	pool *pgxpool.Pool
}

var entryModelImpl *EntryModel

func InitEntryModel(pool *pgxpool.Pool) (*EntryModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if entryModelImpl == nil {
		entryModelImpl = &EntryModel{
			pool: pool,
		}
	}

	return entryModelImpl, nil
}

func GetEntryModel() *EntryModel {
	if entryModelImpl == nil {
		panic("entry model hasnt been initialized")
	}
	return entryModelImpl
}

func (em *EntryModel) GetDisplayEntriesByFarm(farm uint32, page int) ([]entity_public.DisplayEntry, int, *model_error.ModelError) {
	pageSize := 10
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM entry e
		LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
		WHERE ie.entry_id IS NULL AND e.farm = @userFarm
	`
	var totalEntries int
	countRow := em.pool.QueryRow(context.Background(), countQuery, pgx.NamedArgs{"userFarm": farm})
	if err := countRow.Scan(&totalEntries); err != nil {
		fmt.Printf("\n countErr: %v\n", err.Error())
		return nil, 0, &model_error.ModelError{Message: err.Error()}
	}

	rows, queryErr := em.pool.Query(context.Background(), `
		SELECT e.id, p.name, f.name, v.plate, e.netweight, e.arrivaldate, e.farm, COALESCE(origin_union.name, 'Própria')
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN inactive_entry ie ON ie.entry_Id = e.id
			LEFT JOIN vehicle v ON v.id = e.vehicle
			LEFT JOIN entry_origin eo ON eo.entry_id = e.id
			LEFT JOIN (
				SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name FROM legal_person lp
				UNION ALL
				SELECT np.personid, np.name FROM natural_person np
			) origin_union ON origin_union.personid = eo.person_id
			WHERE ie.entry_id IS NULL AND e.farm = @userFarm
			ORDER BY e.arrivaldate DESC
			LIMIT @pageSize OFFSET @offset
	`, pgx.NamedArgs{"userFarm": farm, "pageSize": pageSize, "offset": offset})

	if queryErr != nil {
		fmt.Printf("\n queryErr: %v\n", queryErr.Error())
		return []entity_public.DisplayEntry{}, 0, &model_error.ModelError{Message: queryErr.Error()}
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		fmt.Printf("\n collectErr: %v\n", collectErr.Error())
		return []entity_public.DisplayEntry{}, 0, &model_error.ModelError{Message: collectErr.Error()}
	}

	return entries, totalEntries, nil
}

func (em *EntryModel) GetEntryDraftsByFarm(farm uint32) ([]entity_public.DisplayEntryDraft, *model_error.ModelError) {

	rows, queryErr := em.pool.Query(context.Background(), `
		SELECT ed.id, ed.name, f.name, c.name, v.plate, ed.tare, COALESCE(np.name, lp.fantasyname, lp.companyname, 'Própria') AS origin
			FROM entry_draft ed
			JOIN field f ON ed.field = f.id
			JOIN crop c ON ed.crop = c.id
			LEFT JOIN entry_draft_origin edo ON edo.entry_draft_id = ed.id
			LEFT JOIN person p ON p.id = edo.person_id
			LEFT JOIN natural_person np ON np.personid = p.id
			LEFT JOIN legal_person lp ON lp.personid = p.id
			LEFT JOIN vehicle v ON v.id = ed.vehicle
			WHERE ed.farm = @userFarm;
	`, pgx.NamedArgs{"userFarm": farm})

	if queryErr != nil {
		fmt.Printf("\n queryErr: %v\n", queryErr.Error())

		rows.Close()
		return []entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	drafts, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntryDraft])
	if collectErr != nil {
		fmt.Printf("\n collectErr: %v\n", collectErr.Error())
		return []entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return drafts, nil
}

func (em *EntryModel) AddEntryDraft(ge entity_public.EntryDraft) (entity_public.DisplayEntryDraft, *model_error.ModelError) {
	row, queryErr := em.pool.Query(context.Background(), `
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
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntryDraft])
	if collectErr != nil {
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}
func (em *EntryModel) GetEntryDraft(id uint32) (entity_public.EntryDraft, *model_error.ModelError) {
	row, queryErr := em.pool.Query(context.Background(), `
		SELECT ed.*, edo.person_id AS origin
			FROM entry_draft ed
			JOIN field f ON ed.field = f.id
			JOIN crop c ON ed.crop = c.id
			LEFT JOIN entry_draft_origin edo ON edo.entry_draft_id = ed.id
			WHERE ed.id = @id;
		`, pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.EntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}
	draft, collectErr := pgx.CollectExactlyOneRow(row, pgx.RowToStructByPos[entity_public.EntryDraft])
	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return entity_public.EntryDraft{}, &model_error.ModelError{Message: "Rascunho não encontrado"}
		}
		return entity_public.EntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return draft, nil
}

func (em *EntryModel) AddEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
	row, queryErr := em.pool.Query(context.Background(), `
		SELECT * FROM add_get_entry(@field, @crop, @grossWeight, @tare, @humidity, @vehicle, @netWeight, @arrivalDate, @farm, @damage, @impurity, @humidity_discount_modifier, @origin)
		`, pgx.NamedArgs{
		"field":                      ge.Field,
		"crop":                       ge.Crop,
		"vehicle":                    ge.Vehicle,
		"grossWeight":                ge.GrossWeight,
		"tare":                       ge.Tare,
		"netWeight":                  ge.NetWeight,
		"humidity":                   ge.Humidity,
		"arrivalDate":                ge.ArrivalDate,
		"farm":                       ge.Farm,
		"damage":                     ge.Damage,
		"impurity":                   ge.Impurity,
		"origin":                     ge.Origin,
		"humidity_discount_modifier": ge.HumidityDiscountModifier,
	})

	if queryErr != nil {
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *EntryModel) DeleteEntryDraft(id uint32) error {
	_, err := em.pool.Exec(context.Background(), "DELETE FROM entry_draft WHERE id = @draftId", pgx.NamedArgs{"draftId": id})

	if err != nil {
	}

	return nil
}

func (em *EntryModel) DeleteEntry(id uint32) error {
	_, err := em.pool.Exec(context.Background(), "INSERT INTO inactive_entry (entry_id) VALUES (@entryId)", pgx.NamedArgs{"entryId": id})

	if err != nil {
	}

	return nil
}

func (em *EntryModel) GetEntry(id uint32) (entity_public.Entry, *model_error.ModelError) {

	rows, queryErr := em.pool.Query(context.Background(), "SELECT e.id, e.field, e.crop, e.vehicle, e.grossweight, e.tare, e.netweight, ea.humidity, ea.damage, ea.impurity, ea.humidity_discount_modifier, e.arrivaldate, e.farm, eo.person_id, e.modified_at FROM entry e LEFT JOIN entry_analysis ea ON ea.entryid = e.id LEFT JOIN entry_origin eo ON eo.entry_id = e.id WHERE e.id = @id", pgx.NamedArgs{"id": id})
	if queryErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.Entry])
	if collectErr != nil {
		return entity_public.Entry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *EntryModel) GetEntryPdf(id uint32) (entity_public.EntryPdf, *model_error.ModelError) {

	rows, queryErr := em.pool.Query(
		context.Background(),
		`SELECT
			e.id,
			c.name AS safra,
			v.plate,
			e.grossweight,
			e.tare,
			e.netweight,
			ea.humidity,
			ea.damage,
			ea.impurity,
			e.arrivaldate,
			f.inscricao_estadual,
			p.name AS produto,
			fc.name AS farm_name,
			fa.street AS farm_street,
			fa.cep AS farm_cep,
			fa.number AS farm_number,
			fa.neighborhood AS farm_neighborhood,
			fa.city AS farm_city,
			fa.state AS farm_state,
			COALESCE(person_union.name, fc.name, 'Pŕopria') AS origin,
			COALESCE(person_union.document, f.inscricao_estadual) AS document,
			fc.storage_name,
			COALESCE(et.weight, 0.0),
			COALESCE(et.applied_tax, 0.0)
		FROM entry e
		LEFT JOIN entry_analysis ea ON ea.entryid = e.id
		JOIN field fi ON fi.id = e.field
		JOIN crop c ON c.id = e.crop
		JOIN product p ON p.id = c.product
		JOIN farm f ON f.id = e.farm
		JOIN vehicle v ON v.id = e.vehicle
		LEFT JOIN farm_config fc ON fc.farm_id = f.id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN entry_origin eo ON eo.entry_id = e.id
		LEFT JOIN (
			SELECT lp.personid, COALESCE(lp.companyname, lp.fantasyname) AS name, lp.cnpj AS document FROM legal_person lp
			UNION ALL
			SELECT np.personid, np.name, np.cpf AS document FROM natural_person np
		) person_union ON person_union.personid = eo.person_id
		LEFT JOIN entry_tax et ON et.entry_id = e.id
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

func (em *EntryModel) PutEntry(ge entity_public.Entry) (entity_public.DisplayEntry, *model_error.ModelError) {
	row, queryErr := em.pool.Query(
		context.Background(),
		`SELECT * FROM update_get_display_entry(@field, @crop, @grossWeight, @tare, @humidity, @id, @vehicle, @netWeight, @arrivalDate, @damage, @impurity, @humidity_discount_modifier, @origin)`,
		pgx.NamedArgs{
			"id":                         ge.Id,
			"field":                      ge.Field,
			"crop":                       ge.Crop,
			"vehicle":                    ge.Vehicle,
			"grossWeight":                ge.GrossWeight,
			"tare":                       ge.Tare,
			"netWeight":                  ge.NetWeight,
			"humidity":                   ge.Humidity,
			"arrivalDate":                ge.ArrivalDate,
			"damage":                     ge.Damage,
			"impurity":                   ge.Impurity,
			"origin":                     ge.Origin,
			"humidity_discount_modifier": ge.HumidityDiscountModifier,
		})

	if queryErr != nil {
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		return entity_public.DisplayEntry{}, &model_error.ModelError{Message: collectErr.Error()}
	}
	return entry, nil
}

func (em *EntryModel) PutEntryDraft(ge entity_public.EntryDraft) (entity_public.DisplayEntryDraft, *model_error.ModelError) {
	var tare *decimal.Decimal = ge.Tare

	if ge.Tare.Equal(decimal.Zero) {
		tare = nil
	}
	row, queryErr := em.pool.Query(context.Background(), `
		SELECT * FROM update_get_entry_draft(@id, @name, @field, @crop, @vehicle, @tare, @farm, @origin)
		`, pgx.NamedArgs{
		"id":      ge.Id,
		"name":    ge.Name,
		"field":   ge.Field,
		"crop":    ge.Crop,
		"vehicle": ge.Vehicle,
		"tare":    tare,
		"farm":    ge.Farm,
		"origin":  ge.Origin,
	})

	if queryErr != nil {
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	entry, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.DisplayEntryDraft])
	if collectErr != nil {
		return entity_public.DisplayEntryDraft{}, &model_error.ModelError{Message: collectErr.Error()}
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

func (em *EntryModel) FilterEntries(ef entity_public.EntryFilter, page int, farm uint32) ([]entity_public.DisplayEntry, int, error) {
	filters := ef.GetFilters(availableEntryFilters)
	pageSize := 10
	offset := (page - 1) * pageSize

	whereClauses := []string{"ie.entry_id IS NULL"}
	for _, filter := range filters {
		whereClauses = append(whereClauses, filter(ef))
	}
	whereCondition := strings.Join(whereClauses, " AND ")

	countQuery := `SELECT COUNT(*)
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			JOIN vehicle v ON v.id = e.vehicle
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			WHERE ` + whereCondition + " AND e.farm = @farm"

	var totalEntries int
	countRow := em.pool.QueryRow(context.Background(), countQuery, pgx.NamedArgs{"farm": farm})
	if err := countRow.Scan(&totalEntries); err != nil {
		fmt.Printf("\n countErr: %v\n", err.Error())
		return nil, 0, err
	}

	if totalEntries == 0 {
		return []entity_public.DisplayEntry{}, 0, nil
	}

	query := `SELECT e.id, p.name, f.name, v.plate, e.netweight, e.arrivaldate, e.farm, COALESCE(origin_union.name, 'Própria')
			FROM entry e
			JOIN field f ON e.field = f.id
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			JOIN vehicle v ON v.id = e.vehicle
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			LEFT JOIN entry_origin eo ON eo.entry_id = e.id
			LEFT JOIN (
				SELECT lp.personid, COALESCE(lp.fantasyname, lp.companyname) AS name FROM legal_person lp
				UNION ALL
				SELECT np.personid, np.name FROM natural_person np
			) origin_union ON origin_union.personid = eo.person_id
			WHERE ` + whereCondition + `
			AND e.farm = @farm
			ORDER BY e.arrivaldate DESC
			LIMIT @pageSize OFFSET @offset`

	rows, queryErr := em.pool.Query(context.Background(), query, pgx.NamedArgs{"pageSize": pageSize, "offset": offset, "farm": farm})
	if queryErr != nil {
		return []entity_public.DisplayEntry{}, 0, queryErr
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.DisplayEntry])
	if collectErr != nil {
		return []entity_public.DisplayEntry{}, 0, collectErr
	}

	return entries, totalEntries, nil
}

func (em *EntryModel) AddEntryTax(id uint32, tax decimal.Decimal, appliedTax decimal.Decimal) error {
	_, err := em.pool.Exec(context.Background(), `
		INSERT INTO entry_tax (entry_id, weight, applied_tax) VALUES (@id, @tax, @applied_tax)
		`, pgx.NamedArgs{"id": id, "tax": tax, "applied_tax": appliedTax})

	if err != nil {

	}

	return nil
}

// GetEntriesModifiedSince retrieves entries modified after the given timestamp
func (em *EntryModel) GetEntriesModifiedSince(since time.Time, farm uint32) ([]entity_public.Entry, error) {
	query := `
		SELECT e.id, e.field, e.crop, e.vehicle, e.grossweight, e.tare, e.netweight,
		       ea.humidity, ea.damage, ea.impurity, ea.humidity_discount_modifier,
		       e.arrivaldate, e.farm, eo.person_id as origin, e.modified_at
		FROM entry e
		LEFT JOIN entry_analysis ea ON ea.entryid = e.id
		LEFT JOIN entry_origin eo ON eo.entry_id = e.id
		WHERE e.farm = @farm AND e.modified_at > @since
		ORDER BY e.modified_at ASC
	`

	rows, err := em.pool.Query(context.Background(), query, pgx.NamedArgs{"farm": farm, "since": since})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []entity_public.Entry
	for rows.Next() {
		var entry entity_public.Entry
		var humidity, damage, impurity, humidityModifier *float64
		var origin *uint32
		var modifiedAt time.Time

		err := rows.Scan(
			&entry.Id, &entry.Field, &entry.Crop, &entry.Vehicle,
			&entry.CargoWeight.GrossWeight, &entry.CargoWeight.Tare, &entry.CargoWeight.NetWeight,
			&humidity, &damage, &impurity, &humidityModifier,
			&entry.ArrivalDate, &entry.Farm, &origin, &modifiedAt,
		)
		if err != nil {
			return nil, err
		}

		entry.Origin = origin
		entry.ModifiedAt = modifiedAt

		// Convert float64 pointers to decimal.Decimal pointers
		if humidity != nil {
			h := decimal.NewFromFloat(*humidity)
			entry.Analysis.Humidity = &h
		}
		if damage != nil {
			d := decimal.NewFromFloat(*damage)
			entry.Analysis.Damage = &d
		}
		if impurity != nil {
			i := decimal.NewFromFloat(*impurity)
			entry.Analysis.Impurity = &i
		}
		if humidityModifier != nil {
			hm := decimal.NewFromFloat(*humidityModifier)
			entry.HumidityDiscountModifier = &hm
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetModifiedCount returns the count of entries modified since the given timestamp
func (em *EntryModel) GetModifiedCount(since time.Time, farm uint32) (int, error) {
	query := `SELECT COUNT(*) FROM entry WHERE farm = @farm AND modified_at > @since`

	var count int
	err := em.pool.QueryRow(context.Background(), query, pgx.NamedArgs{"farm": farm, "since": since}).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
