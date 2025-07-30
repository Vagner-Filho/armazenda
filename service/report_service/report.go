package report_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/report_model"
)

func GetReport(rf entity_public.ReportFilter, farm uint32) ([]entity_public.ReportDisplay, *entity_public.Toast) {
	rm := report_model.GetReportModel()
	report, err := rm.FilterReport(rf, farm)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.ReportDisplay{}, &toast
	}
	return report, nil
}

func FilterReport(rf entity_public.ReportFilter, farm uint32) ([]entity_public.ReportDisplay, *entity_public.Toast) {
	rModel := report_model.GetReportModel()
	report, err := rModel.FilterReport(rf, farm)
	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar relatório", "")
		return report, &toast
	}
	return report, nil
}
