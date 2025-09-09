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
