package stats_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/stats_model"
)

func GetTopSupplierStat(farmId uint32, cropID uint8) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetTopSupplier(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetTopBuyerStat(farmId uint32, cropID uint8) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetTopBuyer(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetMostFrequentSupplierStat(farmId uint32, cropID uint8) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetMostFrequentSupplier(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetBestQualitySupplierStat(farmId uint32, cropID uint8) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetBestQualitySupplier(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetWorstQualitySupplierStat(farmId uint32, cropID uint8) (*entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stat, err := sm.GetWorstQualitySupplier(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return nil, &toast
	}
	return &stat, nil
}

func GetProductTotals(farmId uint32, cropID uint8) ([]entity_public.ProductVolumeStat, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	totals, err := sm.GetProductTotals(farmId, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar totais por produto", "")
		return nil, &toast
	}
	return totals, nil
}

func GetFieldProductTotals(farmID uint32, cropID uint8) (*entity_public.FieldProductCharts, *entity_public.Toast) {
	model := stats_model.GetStatsModel()

	nominal, err := model.GetNominalFieldProductTotals(farmID, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar o volume por talhão e produto", "")
		return nil, &toast
	}

	relative, err := model.GetRelativeFieldProductTotals(farmID, cropID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a produtividade por talhão e produto", "")
		return nil, &toast
	}

	return &entity_public.FieldProductCharts{
		Nominal:  nominal,
		Relative: relative,
	}, nil
}
