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
		return fmt.Sprintf("e.arrivalDate >= '%v'", ef.StartDate.Format(utils.DBTimeWithoutTimeZone))
	},
	"EndDate": func(ef entity_public.ReportFilter) string {
		return fmt.Sprintf("e.arrivalDate <= '%v'", ef.EndDate.Format(utils.DBTimeWithoutTimeZone))
	},
	"Vehicle": func(ef entity_public.ReportFilter) string {
		return fmt.Sprintf("e.vehicle = '%s'", ef.Vehicle)
	},
	"Product": func(ef entity_public.ReportFilter) string {
		return "p.id = " + strconv.FormatInt(int64(ef.Product), 10)
	},
	"NetWeightMin": func(ef entity_public.ReportFilter) string {
		return "e.netweight >= " + strconv.FormatFloat(ef.NetWeightMin, 'f', -1, 64)
	},
	"NetWeightMax": func(ef entity_public.ReportFilter) string {
		return "e.netweight <= " + strconv.FormatFloat(ef.NetWeightMax, 'f', -1, 64)
	},
}

func (rm *reportModel) FilterEntries(rf entity_public.ReportFilter, farm uint32) ([]entity_public.ReportDisplay, error) {
	filters := rf.GetFilters(availableReportFilters)

	stmt := `SELECT e.id, 0 AS operation_type, p.name, e.vehicle, e.netweight, e.arrivaldate
			FROM entry e
			JOIN crop c ON e.crop = c.id
			JOIN product p ON c.product = p.id
			WHERE e.farm = @userFarm
			UNION ALL
		SELECT d.id, 1 AS operation_type, p.name, d.vehicle, d.netweight , d.departuredate
			FROM departure d
			JOIN crop c ON d.crop = c.id
			JOIN product p ON c.product = p.id
			WHERE d.farm = @userFarm`

	for _, filter := range filters {
		stmt += "\nAND " + filter(rf)
	}

	rows, queryErr := rm.conn.Query(context.Background(), stmt, pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		model_error.Logger(rm.conn, queryErr.Error())
		return []entity_public.ReportDisplay{}, queryErr
	}

	entries, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.ReportDisplay])
	if collectErr != nil {
		model_error.Logger(rm.conn, collectErr.Error())
		return []entity_public.ReportDisplay{}, collectErr
	}

	return entries, nil
}
