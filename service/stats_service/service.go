package stats_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/stats_model"
)

func GetTopSupplierStat(farmId uint32) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetTopSupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetTopBuyerStat(farmId uint32) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetTopBuyer(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetMostFrequentSupplierStat(farmId uint32) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetMostFrequentSupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetBestQualitySupplierStat(farmId uint32) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetBestQualitySupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetWorstQualitySupplierStat(farmId uint32) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetWorstQualitySupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetProductTotals(farmId uint32) ([]entity_public.ProductVolumeStat, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	totals, err := sm.GetProductTotals(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar totais por produto", "")
		return nil, &toast
	}
	return totals, nil
}

func GetFieldProductTotals(farmID uint32) (*entity_public.FieldProductCharts, *entity_public.Toast) {
	model := stats_model.GetStatsModel()

	nominal, err := model.GetNominalFieldProductTotals(farmID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar o volume por talhão e produto", "")
		return nil, &toast
	}

	relative, err := model.GetRelativeFieldProductTotals(farmID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a produtividade por talhão e produto", "")
		return nil, &toast
	}

	return &entity_public.FieldProductCharts{
		Nominal:  nominal,
		Relative: relative,
	}, nil
}
