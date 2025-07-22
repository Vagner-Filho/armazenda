package report_view

import (
	entity_public "armazenda/entity/public"
	field_service "armazenda/service/field"
	product_service "armazenda/service/product"
	"armazenda/service/report_service"
	"armazenda/service/vehicle_service"
	"time"
)

type reportView struct {
	Products []entity_public.Product
	Vehicles []entity_public.Vehicle
	Fields   map[string][]entity_public.Field
	reportContent
	StartDate time.Time `form:"initialDate" binding:"required" time_format:"2006-01-02T15:04"`
	EndDate   time.Time `form:"endDate" binding:"required" time_format:"2006-01-02T15:04"`
}

type reportContent struct {
	Operations     []entity_public.ReportDisplay
	EntryTotal     float64
	DepartureTotal float64
	Balance        float64
}

func GetReportBalance(report []entity_public.ReportDisplay) (entry float64, departure float64, balance float64) {
	var entryAmount float64
	var departureAmount float64
	for _, r := range report {
		if r.OperationType == 0 {
			entryAmount += r.NetWeight
		} else {
			departureAmount += r.NetWeight
		}
	}
	return entryAmount, departureAmount, entryAmount - departureAmount

}

func GetReportPage(farm uint32) reportView {
	products, _ := product_service.GetProducts()
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
	fields, _ := field_service.GetFieldsByFarm(farm)
	report, _ := report_service.GetReport(entity_public.ReportFilter{}, farm)
	entryAmount, departureAmount, balance := GetReportBalance(report)
	return reportView{
		Products: products,
		Vehicles: vehicles,
		Fields: map[string][]entity_public.Field{
			"Fields": fields,
		},
		reportContent: reportContent{
			Operations:     report,
			EntryTotal:     entryAmount,
			DepartureTotal: departureAmount,
			Balance:        balance,
		},
	}
}

func FilterReport(rf entity_public.ReportFilter, farm uint32) (reportContent, *entity_public.Toast) {
	report, toast := report_service.FilterReport(rf, farm)
	entryAmount, departureAmount, balance := GetReportBalance(report)
	return reportContent{
		Operations:     report,
		EntryTotal:     entryAmount,
		DepartureTotal: departureAmount,
		Balance:        balance,
	}, toast
}
