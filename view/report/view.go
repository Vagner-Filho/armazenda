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
	Products    []entity_public.Product
	Vehicles    map[string][]entity_public.Vehicle
	Fields      map[string][]entity_public.Field
	Operations  []entity_public.ReportDisplay
	InitialDate time.Time `form:"initialDate" binding:"required" time_format:"2006-01-02T15:04"`
	EndDate     time.Time `form:"endDate" binding:"required" time_format:"2006-01-02T15:04"`
}

func GetReportPage(farm uint32) reportView {
	products, _ := product_service.GetProducts()
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
	fields, _ := field_service.GetFieldsByFarm(farm)
	report, _ := report_service.GetReport(entity_public.ReportFilter{}, farm)
	return reportView{
		Products: products,
		Vehicles: map[string][]entity_public.Vehicle{
			"Vehicle": vehicles,
		},
		Fields: map[string][]entity_public.Field{
			"Fields": fields,
		},
		Operations: report,
	}
}
