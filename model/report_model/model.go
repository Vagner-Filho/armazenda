package report_model

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

type reportModel struct {
	conn *pgx.Conn
}

var reportModelImpl *reportModel

func InitReportModel(conn *pgx.Conn) (*reportModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if reportModelImpl == nil {
		reportModelImpl = &reportModel{
			conn: conn,
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
		return "r.id = " + strconv.FormatInt(int64(ef.Product), 10)
	},
	"NetWeightMin": func(ef entity_public.ReportFilter) string {
		return "r.netweight >= " + strconv.FormatFloat(ef.NetWeightMin, 'f', -1, 64)
	},
	"NetWeightMax": func(ef entity_public.ReportFilter) string {
		return "r.netweight <= " + strconv.FormatFloat(ef.NetWeightMax, 'f', -1, 64)
	},
}

func (rm *reportModel) FilterReport(rf entity_public.ReportFilter, farm uint32) ([]entity_public.ReportDisplay, error) {
	filters := rf.GetFilters(availableReportFilters)

	stmt := `WITH people AS (SELECT np.name, np.personid FROM natural_person np UNION ALL SELECT lp.companyname AS name, lp.personid FROM legal_person lp)
		SELECT * FROM (SELECT e.id, 0 AS operation_type, p.name, e.vehicle, e.netweight, e.arrivaldate AS date, coalesce(prs.name, 'Pŕopria') AS pessoa
			FROM entry e
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN entry_origin eo ON eo.entry_id = e.id
			LEFT JOIN people
			AS prs ON prs.personid = eo.person_id
			WHERE e.farm = @userFarm
			UNION ALL
		SELECT d.id, 1 AS operation_type, p.name, d.vehicle, d.netweight , d.departuredate AS date, coalesce(prs.name, 'Pŕopria') AS pessoa
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			LEFT JOIN departure_recipient dr ON dr.departureid = d.id
			LEFT JOIN people
			AS prs ON prs.personid = dr.personid
			WHERE d.farm = @userFarm) AS r;`

	if len(filters) > 0 {
		stmt += " WHERE "
	}

	var idx int = 0
	for _, filter := range filters {
		stmt += filter(rf)
		if idx < len(filters)-1 {
			stmt += " AND "
		}
		idx++
	}

	rows, queryErr := rm.conn.Query(context.Background(), stmt, pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		model_error.Logger(rm.conn, queryErr.Error())
		return []entity_public.ReportDisplay{}, queryErr
	}

	result, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.ReportDisplay])
	if collectErr != nil {
		model_error.Logger(rm.conn, collectErr.Error())
		return []entity_public.ReportDisplay{}, collectErr
	}

	return result, nil
}
