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
	"armazenda/view/filters"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

func BuildAppliedChips(rf entity_public.ReportFilter, farm uint32) []filters.ChipEntry {
	chips := []filters.ChipEntry{}

	if rf.Type != 0 {
		v := ""
		if rf.Type == 1 {
			v = "Entrada"
		} else if rf.Type == 2 {
			v = "Saída"
		}
		if v != "" {
			chips = append(chips, filters.ChipEntry{Key: "type", Label: "Operação", Value: v})
		}
	}
	if rf.Product != 0 {
		products, _ := product_service.GetProducts()
		for _, p := range products {
			if p.Id == rf.Product {
				chips = append(chips, filters.ChipEntry{Key: "product", Label: "Grão", Value: p.Name})
				break
			}
		}
	}
	if rf.OriginId != "" {
		if rf.OriginId == "NULL" {
			chips = append(chips, filters.ChipEntry{Key: "origin", Label: "Origem", Value: "Própria"})
		} else {
			people, _ := person_service.GetPeopleByFarm(farm)
			for _, p := range people {
				if p.Id != nil && fmt.Sprintf("%v", *p.Id) == rf.OriginId {
					chips = append(chips, filters.ChipEntry{Key: "origin", Label: "Origem", Value: p.Name})
					break
				}
			}
		}
	}
	if rf.RecipientId != "" {
		if rf.RecipientId == "NULL" {
			chips = append(chips, filters.ChipEntry{Key: "recipient", Label: "Destino", Value: "Própria"})
		} else {
			people, _ := person_service.GetPeopleByFarm(farm)
			for _, p := range people {
				if p.Id != nil && fmt.Sprintf("%v", *p.Id) == rf.RecipientId {
					chips = append(chips, filters.ChipEntry{Key: "recipient", Label: "Destino", Value: p.Name})
					break
				}
			}
		}
	}
	if rf.FieldId != 0 {
		fields, _ := field_service.GetFieldsByFarm(farm)
		for _, f := range fields {
			if f.Id == rf.FieldId {
				chips = append(chips, filters.ChipEntry{Key: "field", Label: "Talhão", Value: f.Name})
				break
			}
		}
	}
	if rf.Vehicle != "" {
		vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
		display := rf.Vehicle
		for _, v := range vehicles {
			if fmt.Sprintf("%v", v.Id) == rf.Vehicle {
				display = v.Plate
				if v.Name != "" {
					display = v.Plate + " | " + v.Name
				}
				break
			}
		}
		chips = append(chips, filters.ChipEntry{Key: "vehiclePlate", Label: "Veículo", Value: display})
	}
	if rf.NetWeightMin != 0 {
		chips = append(chips, filters.ChipEntry{Key: "netWeightMin", Label: "Peso mín.", Value: fmt.Sprintf("%.0f kg", rf.NetWeightMin)})
	}
	if rf.NetWeightMax != 0 {
		chips = append(chips, filters.ChipEntry{Key: "netWeightMax", Label: "Peso máx.", Value: fmt.Sprintf("%.0f kg", rf.NetWeightMax)})
	}
	if !rf.StartDate.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "startDate", Label: "Início", Value: rf.StartDate.Format("02/01/2006 15:04")})
	}
	if !rf.EndDate.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "endDate", Label: "Fim", Value: rf.EndDate.Format("02/01/2006 15:04")})
	}

	return chips
}

type reportView struct {
	view.BaseTemplateData
	Products []entity_public.Product
	Vehicles []entity_public.Vehicle
	Fields   map[string][]entity_public.Field
	reportContent
	StartDate    time.Time `form:"initialDate" binding:"required" time_format:"2006-01-02T15:04"`
	EndDate      time.Time `form:"endDate" binding:"required" time_format:"2006-01-02T15:04"`
	People       []entity_public.PersonOption
	Stats        []entity_public.StatCard
	AppliedChips filters.FilterChips
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
	OOB            bool
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
		AppliedChips: filters.FilterChips{Items: []filters.ChipEntry{}},
	}
}

type ReportFiltersForm struct {
	Vehicles []entity_public.Vehicle
	Fields   map[string][]entity_public.Field
	People   []entity_public.PersonOption
}

type ClearedReportView struct {
	Form    ReportFiltersForm
	Chips   filters.FilterChips
	Content reportContent
}

func GetReportFiltersData(farm uint32) ReportFiltersForm {
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
	fields, _ := field_service.GetFieldsByFarm(farm)
	people, _ := person_service.GetPeopleByFarm(farm)
	return ReportFiltersForm{
		Vehicles: vehicles,
		Fields: map[string][]entity_public.Field{
			"Fields": fields,
		},
		People: people,
	}
}

func GetClearedReport(farm uint32, cspNonce string) ClearedReportView {
	form := GetReportFiltersData(farm)
	report, totalCount, entryAmount, departureAmount, balance, _ := report_service.GetReport(entity_public.ReportFilter{}, farm, 1)
	currentPage, totalPages, nextPage, prevPage, hasNext, hasPrev := buildReportPagination(1, totalCount)
	content := reportContent{
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
		OOB:            true,
	}
	content.CSPNonce = cspNonce
	return ClearedReportView{
		Form:    form,
		Chips:   filters.FilterChips{Items: []filters.ChipEntry{}, OOB: true},
		Content: content,
	}
}

type FilterApplyResponse struct {
	Chips   filters.FilterChips
	Content reportContent
}

func FilterReport(rf entity_public.ReportFilter, farm uint32, page int, cspNonce string) (FilterApplyResponse, *entity_public.Toast) {
	report, totalCount, entryAmount, departureAmount, balance, toast := report_service.FilterReport(rf, farm, page)
	if toast != nil {
		return FilterApplyResponse{}, toast
	}
	currentPage, totalPages, nextPage, prevPage, hasNext, hasPrev := buildReportPagination(page, totalCount)
	content := reportContent{
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
		OOB:            true,
	}
	content.CSPNonce = cspNonce
	return FilterApplyResponse{
		Chips:   filters.FilterChips{Items: BuildAppliedChips(rf, farm), OOB: true},
		Content: content,
	}, nil
}

type FullReportRow struct {
	entity_public.ReportDisplay
	GrossWeight        decimal.Decimal
	Tare               decimal.Decimal
	City               string
	State              string
	Humidity           decimal.Decimal
	Damage             decimal.Decimal
	Impurity           decimal.Decimal
	HumidityDiscount   decimal.Decimal
	DiscountedHumidity string
	DiscountedDamage   string
	DiscountedImpurity string
	ServiceTax         *decimal.Decimal
	WeightTax          string
}

type FullReportView struct {
	view.BaseTemplateData
	FullOperations []FullReportRow
	EntryTotal     float64
	DepartureTotal float64
	Balance        float64
	RequestedAt    time.Time `time_format:"2006-01-02T15:04"`
	AppliedFilters map[string]string
	FarmConfig     *entity_public.Farm
}

func formatDec(d decimal.Decimal, places int32) string {
	return d.StringFixed(places)
}

func formatDecPtr(d *decimal.Decimal, places int32) string {
	if d == nil {
		return decimal.Zero.StringFixed(places)
	}
	return d.StringFixed(places)
}

func GetFullReport(rf entity_public.ReportFilter, farm uint32) (FullReportView, *entity_public.Toast) {
	report, toast := report_service.GetFullReport(rf, farm)

	rows := make([]FullReportRow, len(report))
	for i, r := range report {
		rows[i] = FullReportRow{
			ReportDisplay:      r.ReportDisplay,
			GrossWeight:        r.GrossWeight,
			Tare:               r.Tare,
			City:               r.City,
			State:              r.State,
			Humidity:           r.Humidity,
			Damage:             r.Damage,
			Impurity:           r.Impurity,
			HumidityDiscount:   r.HumidityDiscount,
			DiscountedHumidity: formatDec(r.DiscountedHumidity, 2),
			DiscountedDamage:   formatDec(r.DiscountedDamage, 2),
			DiscountedImpurity: formatDec(r.DiscountedImpurity, 2),
			ServiceTax:         r.ServiceTax,
			WeightTax:          formatDecPtr(r.WeightTax, 2),
		}
	}

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
			case "OriginId":
				if len(report) > 0 && report[0].OriginId != nil {
					appliedFilters["Origem"] = report[0].OriginName
				}
			case "RecipientId":
				if len(report) > 0 && report[0].RecipientId != nil {
					appliedFilters["Destino"] = report[0].RecipientName
				}
			case "FieldId":
				fields, _ := field_service.GetFieldsByFarm(farm)
				for _, f := range fields {
					if f.Id == rf.FieldId {
						appliedFilters["Talhão"] = f.Name
						break
					}
				}
			}
		}
	}

	fcModel := farm_config_model.GetFarmConfigModel()
	farmConfig, errConfig := fcModel.GetFarmConfig(farm)
	if errConfig != nil {
		toast := entity_public.GetInfoToast("Falha ao obter configuração da fazenda", "")
		return FullReportView{
			FullOperations: rows,
			EntryTotal:     entry,
			DepartureTotal: departure,
			Balance:        balance,
			RequestedAt:    time.Now(),
			AppliedFilters: appliedFilters,
		}, &toast
	}

	return FullReportView{
		FullOperations: rows,
		EntryTotal:     entry,
		DepartureTotal: departure,
		Balance:        balance,
		RequestedAt:    time.Now(),
		AppliedFilters: appliedFilters,
		FarmConfig:     farmConfig,
	}, toast
}
