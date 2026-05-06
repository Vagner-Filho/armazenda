package report_model

import (
	entity_public "armazenda/entity/public"
	"armazenda/utils"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reportModel struct {
	pool *pgxpool.Pool
}

var reportModelImpl *reportModel

func InitReportModel(pool *pgxpool.Pool) (*reportModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if reportModelImpl == nil {
		reportModelImpl = &reportModel{
			pool: pool,
		}
	}

	return reportModelImpl, nil
}

func GetReportModel() *reportModel {
	if reportModelImpl == nil {
		panic("report model hasnt been initialized")
	}
	return reportModelImpl
}

var availableReportFilters = map[string]func(ef entity_public.ReportFilter) string{
	"StartDate": func(ef entity_public.ReportFilter) string {
		return fmt.Sprintf("r.date >= '%v'", ef.StartDate.Format(utils.DBTimeWithoutTimeZone))
	},
	"EndDate": func(ef entity_public.ReportFilter) string {
		return fmt.Sprintf("r.date <= '%v'", ef.EndDate.Format(utils.DBTimeWithoutTimeZone))
	},
	"Vehicle": func(ef entity_public.ReportFilter) string {
		return fmt.Sprintf("r.vehicle = '%s'", ef.Vehicle)
	},
	"Product": func(ef entity_public.ReportFilter) string {
		return "r.product_id = " + strconv.FormatInt(int64(ef.Product), 10)
	},
	"NetWeightMin": func(ef entity_public.ReportFilter) string {
		return "r.netweight >= " + strconv.FormatFloat(ef.NetWeightMin, 'f', -1, 64)
	},
	"NetWeightMax": func(ef entity_public.ReportFilter) string {
		return "r.netweight <= " + strconv.FormatFloat(ef.NetWeightMax, 'f', -1, 64)
	},
	"OriginId": func(ef entity_public.ReportFilter) string {
		if ef.OriginId == "NULL" {
			return "r.origin_id IS NULL"
		}
		return fmt.Sprintf("r.origin_id = %s", ef.OriginId)
	},
	"RecipientId": func(ef entity_public.ReportFilter) string {
		if ef.RecipientId == "NULL" {
			return "r.recipient_id IS NULL"
		}
		return fmt.Sprintf("r.recipient_id = %s", ef.RecipientId)
	},
	"Type": func(rf entity_public.ReportFilter) string {
		return fmt.Sprintf("r.operation_type = %v", rf.Type)
	},
}

func buildReportWhereClause(rf entity_public.ReportFilter) string {
	filters := rf.GetFilters(availableReportFilters)

	if len(filters) == 0 {
		return ""
	}

	whereClause := " WHERE "
	idx := 0
	for _, filter := range filters {
		whereClause += filter(rf)
		if idx < len(filters)-1 {
			whereClause += " AND "
		}
		idx++
	}
	return whereClause
}

func getReportSubquery() string {
	return `
	FROM (SELECT e.id, 1 AS operation_type, p.name, v.plate AS vehicle, e.netweight, e.arrivaldate AS date, coalesce(prs_origin.name, 'Própria') AS origin_name, prs_origin.personid AS origin_id, p.id AS product_id, '-' AS recipient_name, NULL AS recipient_id
		FROM entry e
		JOIN crop c ON e.crop = c.id
		JOIN product p ON c.product = p.id
		LEFT JOIN entry_origin eo ON eo.entry_id = e.id
		LEFT JOIN people
		AS prs_origin ON prs_origin.personid = eo.person_id
		LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
		JOIN vehicle v ON v.id = e.vehicle
		WHERE e.farm = @userFarm AND ie.entry_id IS NULL
		UNION ALL
	SELECT d.id, 2 AS operation_type, p.name, v.plate AS vehicle, d.netweight , d.departuredate AS date, coalesce(prs_origin.name, 'Própria') AS origin_name, prs_origin.personid AS origin_id, p.id AS product_id, COALESCE(prs_recipient.name, 'Própria') AS recipient_name, prs_recipient.personid AS recipient_id
		FROM departure d
		JOIN crop c ON d.crop = c.id
		JOIN product p ON c.product = p.id
		LEFT JOIN departure_origin dor ON dor.departure_id = d.id
		LEFT JOIN people AS prs_origin ON prs_origin.personid = dor.person_id
		LEFT JOIN departure_recipient dre ON dre.departure_id = d.id
		LEFT JOIN people AS prs_recipient ON prs_recipient.personid = dre.person_id
		LEFT OUTER JOIN inactive_departure id ON id.departure_id = d.id
		JOIN vehicle v ON v.id = d.vehicle
		WHERE d.farm = @userFarm AND id.departure_id IS NULL) AS r`
}

func (rm *reportModel) FilterReport(rf entity_public.ReportFilter, farm uint32, page int) ([]entity_public.ReportDisplay, int, float64, float64, float64, error) {
	whereClause := buildReportWhereClause(rf)

	cte := `
	WITH people AS (SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT COALESCE(lp.fantasyname, lp.companyname) AS name, lp.personid FROM legal_person lp)
	`

	subquery := getReportSubquery()

	countStmt := cte + "SELECT COUNT(*), COALESCE(SUM(CASE WHEN r.operation_type = 1 THEN r.netweight ELSE 0 END), 0), COALESCE(SUM(CASE WHEN r.operation_type = 2 THEN r.netweight ELSE 0 END), 0) " + subquery + whereClause

	var totalCount int
	var entryTotal, departureTotal float64
	countRow := rm.pool.QueryRow(context.Background(), countStmt, pgx.NamedArgs{"userFarm": farm})
	if err := countRow.Scan(&totalCount, &entryTotal, &departureTotal); err != nil {
		return nil, 0, 0, 0, 0, err
	}

	balance := entryTotal - departureTotal

	if totalCount == 0 {
		return []entity_public.ReportDisplay{}, 0, 0, 0, 0, nil
	}

	pageSize := 10
	offset := (page - 1) * pageSize

	dataStmt := cte + "SELECT r.id, r.operation_type, r.name, r.vehicle, r.netweight, r.date, r.origin_name, r.origin_id, r.recipient_name, r.recipient_id " + subquery + whereClause + " ORDER BY r.date DESC LIMIT @pageSize OFFSET @offset"

	rows, queryErr := rm.pool.Query(context.Background(), dataStmt, pgx.NamedArgs{"userFarm": farm, "pageSize": pageSize, "offset": offset})
	if queryErr != nil {
		return nil, 0, 0, 0, 0, queryErr
	}

	result, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.ReportDisplay])
	if collectErr != nil {
		return nil, 0, 0, 0, 0, collectErr
	}

	return result, totalCount, entryTotal, departureTotal, balance, nil
}

func (rm *reportModel) GetFullReport(rf entity_public.ReportFilter, farm uint32) ([]entity_public.FullReport, error) {
	filters := rf.GetFilters(availableReportFilters)

	stmt := `
	WITH people AS (SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT COALESCE(lp.fantasyname, lp.companyname) AS name, lp.personid FROM legal_person lp)
	SELECT r.id, r.operation_type, r.name, r.vehicle, r.netweight, r.date, r.origin_name, r.origin_id, r.recipient_name, r.recipient_id, r.grossweight, r.tare, r.city, r.state, r.humidity, r.damage, r.impurity, r.humidity_discount FROM
	(SELECT e.id, 1 AS operation_type,
			p.name, v.plate AS vehicle, e.netweight, e.arrivaldate AS date,
			COALESCE(prs_origin.name, 'Própria') AS origin_name, prs_origin.personid AS origin_id, '-' AS recipient_name, NULL AS recipient_id, e.grossweight, e.tare, COALESCE(a.city, 'N/A') AS city,
			COALESCE(a.state, 'N/A') AS state, COALESCE(ea.humidity, 0) AS humidity,
			COALESCE(ea.damage, 0) AS damage, COALESCE(ea.impurity, 0) AS impurity, p.id AS product_id,
			COALESCE(ea.humidity_discount_modifier, 1.15) AS humidity_discount
			FROM entry e
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN entry_origin eo ON eo.entry_id = e.id
			LEFT JOIN people
			AS prs_origin ON prs_origin.personid = eo.person_id
			LEFT JOIN address a ON a.person_id = eo.person_id
			LEFT JOIN entry_analysis ea ON ea.entryid = e.id
			LEFT JOIN person_config pc ON pc.person_id = eo.person_id
			LEFT OUTER JOIN inactive_entry ie ON ie.entry_Id = e.id
			JOIN vehicle v ON v.id = e.vehicle
			WHERE e.farm = @userFarm AND ie.entry_id IS NULL
			UNION ALL
		SELECT d.id, 2 AS operation_type,
			p.name, v.plate AS vehicle, d.netweight, d.departuredate AS date,
			COALESCE(prs_origin.name, 'Própria') AS origin_name, prs_origin.personid AS origin_id, COALESCE(prs_recipient.name, 'Pŕopria') AS recipient_name, prs_recipient.personid AS recipient_id, d.grossweight, d.tare, COALESCE(a.city, 'N/A') AS city,
			COALESCE(a.state, 'N/A') AS state, 0 AS humidity,
			0 AS damage, 0 AS impurity, p.id AS product_id,
			1 AS humidity_discount
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN departure_origin dor ON dor.departure_id = d.id
			LEFT JOIN people
			AS prs_origin ON prs_origin.personid = dor.person_id
			LEFT JOIN departure_recipient dre ON dre.departure_id = d.id
			LEFT JOIN people AS prs_recipient ON prs_recipient.personid = dre.person_id
			LEFT JOIN address a ON a.person_id = dor.person_id
			LEFT JOIN person_config pc ON pc.person_id = dor.person_id
			LEFT OUTER JOIN inactive_departure id ON id.departure_id = d.id
			JOIN vehicle v ON v.id = d.vehicle
			WHERE d.farm = @userFarm AND id.departure_id IS NULL) AS r
	`
	if len(filters) > 0 {
		stmt += " WHERE "
	}

	idx := 0
	for _, filter := range filters {
		stmt += filter(rf)
		if idx < len(filters)-1 {
			stmt += " AND "
		}
		idx++
	}

	rows, queryErr := rm.pool.Query(context.Background(), stmt, pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		return []entity_public.FullReport{}, queryErr
	}

	result, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.FullReport])
	if collectErr != nil {
		return []entity_public.FullReport{}, collectErr
	}

	return result, nil
}
