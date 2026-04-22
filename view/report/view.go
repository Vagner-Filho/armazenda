package report_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/farm_config_model"
	field_service "armazenda/service/field"
	person_service "armazenda/service/person"
	product_service "armazenda/service/product"
	"armazenda/service/report_service"
	"armazenda/service/vehicle_service"
	"armazenda/view"
	"reflect"
	"strconv"
	"time"
)

type reportView struct {
	view.BaseTemplateData
	Products []entity_public.Product
	Vehicles []entity_public.Vehicle
	Fields   map[string][]entity_public.Field
	reportContent
	StartDate time.Time `form:"initialDate" binding:"required" time_format:"2006-01-02T15:04"`
	EndDate   time.Time `form:"endDate" binding:"required" time_format:"2006-01-02T15:04"`
	People    []entity_public.PersonOption
	Stats     []entity_public.StatCard
}

type reportContent struct {
	view.BaseTemplateData
	Operations     []entity_public.ReportDisplay
	EntryTotal     float64
	DepartureTotal float64
	Balance        float64
	CurrentPage    int
	TotalPages     int
	NextPage       int
	PrevPage       int
	HasNext        bool
	HasPrev        bool
}

func GetReportBalance(report []entity_public.ReportDisplay) (entry float64, departure float64, balance float64) {
	var entryAmount float64
	var departureAmount float64
	for _, r := range report {
		if r.OperationType == 1 {
			entryAmount += r.NetWeight
		} else {
			departureAmount += r.NetWeight
		}
	}
	return entryAmount, departureAmount, entryAmount - departureAmount

}

func buildReportPagination(page int, totalCount int) (currentPage int, totalPages int, nextPage int, prevPage int, hasNext bool, hasPrev bool) {
	if page < 1 {
		page = 1
	}
	pageSize := 10
	totalPages = 0
	if totalCount > 0 {
		totalPages = (totalCount + pageSize - 1) / pageSize
	}

	hasNext = page < totalPages
	hasPrev = page > 1

	nextPage = page + 1
	if !hasNext {
		nextPage = page
	}

	prevPage = page - 1
	if !hasPrev {
		prevPage = page
	}

	return page, totalPages, nextPage, prevPage, hasNext, hasPrev
}

func GetReportPage(farm uint32, page int) reportView {
	products, _ := product_service.GetProducts()
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
	fields, _ := field_service.GetFieldsByFarm(farm)
	report, totalCount, entryAmount, departureAmount, balance, _ := report_service.GetReport(entity_public.ReportFilter{}, farm, page)
	people, _ := person_service.GetPeopleByFarm(farm)
	currentPage, totalPages, nextPage, prevPage, hasNext, hasPrev := buildReportPagination(page, totalCount)
	return reportView{
		Products: products,
		Vehicles: vehicles,
		Fields: map[string][]entity_public.Field{
			"Fields": fields,
		},
		People: people,
		reportContent: reportContent{
			Operations:     report,
			EntryTotal:     entryAmount,
			DepartureTotal: departureAmount,
			Balance:        balance,
			CurrentPage:    currentPage,
			TotalPages:     totalPages,
			NextPage:       nextPage,
			PrevPage:       prevPage,
			HasNext:        hasNext,
			HasPrev:        hasPrev,
		},
	}
}

func FilterReport(rf entity_public.ReportFilter, farm uint32, page int) (reportContent, *entity_public.Toast) {
	report, totalCount, entryAmount, departureAmount, balance, toast := report_service.FilterReport(rf, farm, page)
	if toast != nil {
		return reportContent{}, toast
	}
	currentPage, totalPages, nextPage, prevPage, hasNext, hasPrev := buildReportPagination(page, totalCount)
	return reportContent{
		Operations:     report,
		EntryTotal:     entryAmount,
		DepartureTotal: departureAmount,
		Balance:        balance,
		CurrentPage:    currentPage,
		TotalPages:     totalPages,
		NextPage:       nextPage,
		PrevPage:       prevPage,
		HasNext:        hasNext,
		HasPrev:        hasPrev,
	}, nil
}

type FullReportView struct {
	view.BaseTemplateData
	FullOperations []entity_public.FullReport
	EntryTotal     float64
	DepartureTotal float64
	Balance        float64
	RequestedAt    time.Time `time_format:"2006-01-02T15:04"`
	AppliedFilters map[string]string
	FarmConfig     *entity_public.Farm
}

func GetFullReport(rf entity_public.ReportFilter, farm uint32) (FullReportView, *entity_public.Toast) {
	report, toast := report_service.GetFullReport(rf, farm)

	reportDisplay := make([]entity_public.ReportDisplay, 0, len(report))
	for _, r := range report {
		reportDisplay = append(reportDisplay, r.ReportDisplay)
	}
	entry, departure, balance := GetReportBalance(reportDisplay)

	appliedFilters := make(map[string]string)
	for i := range reflect.ValueOf(rf).NumField() {
		field := reflect.ValueOf(rf).Type().Field(i)
		fieldName := field.Name
		fieldValue := reflect.ValueOf(rf).Field(i)

		if !fieldValue.IsZero() {
			switch fieldName {
			case "StartDate":
				appliedFilters["Data Inicial"] = fieldValue.Interface().(time.Time).Format("02/01/2006 15:04")
			case "EndDate":
				appliedFilters["Data Final"] = fieldValue.Interface().(time.Time).Format("02/01/2006 15:04")
			case "Product":
				if len(report) > 0 {
					appliedFilters["Produto"] = report[0].Product
				}
			case "Vehicle":
				appliedFilters["Veículo"] = fieldValue.String()
			case "NetWeightMin":
				appliedFilters["Peso Mínimo"] = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
			case "NetWeightMax":
				appliedFilters["Peso Máximo"] = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
			case "PersonId":
				if len(report) > 0 {
					appliedFilters["Pessoa"] = report[0].Person
				}
			}
		}
	}

	fcModel := farm_config_model.GetFarmConfigModel()
	farmConfig, errConfig := fcModel.GetFarmConfig(farm)
	if errConfig != nil {
		toast := entity_public.GetInfoToast("Falha ao obter configuração da fazenda", "")
		return FullReportView{
			FullOperations: report,
			EntryTotal:     entry,
			DepartureTotal: departure,
			Balance:        balance,
			RequestedAt:    time.Now(),
			AppliedFilters: appliedFilters,
		}, &toast
	}

	return FullReportView{
		FullOperations: report,
		EntryTotal:     entry,
		DepartureTotal: departure,
		Balance:        balance,
		RequestedAt:    time.Now(),
		AppliedFilters: appliedFilters,
		FarmConfig:     farmConfig,
	}, toast
}
