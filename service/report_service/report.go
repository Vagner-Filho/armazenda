package report_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/report_model"
	"armazenda/service/entry_service"
)

func GetReport(rf entity_public.ReportFilter, farm uint32, page int) ([]entity_public.ReportDisplay, int, float64, float64, float64, *entity_public.Toast) {
	rm := report_model.GetReportModel()
	report, totalCount, entryTotal, departureTotal, balance, err := rm.FilterReport(rf, farm, page)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.ReportDisplay{}, 0, 0, 0, 0, &toast
	}
	return report, totalCount, entryTotal, departureTotal, balance, nil
}

func FilterReport(rf entity_public.ReportFilter, farm uint32, page int) ([]entity_public.ReportDisplay, int, float64, float64, float64, *entity_public.Toast) {
	rModel := report_model.GetReportModel()
	report, totalCount, entryTotal, departureTotal, balance, err := rModel.FilterReport(rf, farm, page)
	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar relatório", "")
		return report, 0, 0, 0, 0, &toast
	}
	return report, totalCount, entryTotal, departureTotal, balance, nil
}

func GetFullReport(rf entity_public.ReportFilter, farm uint32) ([]entity_public.FullReport, *entity_public.Toast) {
	rModel := report_model.GetReportModel()
	report, err := rModel.GetFullReport(rf, farm)
	for i, r := range report {
		report[i].DiscountedHumidity = entry_service.DiscountHumidity(&r.Humidity, r.GrossWeight.Sub(r.Tare), &r.HumidityDiscount, nil)
		report[i].DiscountedDamage = entry_service.DiscountDamage(&r.Damage, r.GrossWeight.Sub(r.Tare))
		report[i].DiscountedImpurity = entry_service.DiscountImpurity(&r.Impurity, r.GrossWeight.Sub(r.Tare))
	}
	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao gerar relatório", "")
		return report, &toast
	}
	return report, nil
}
