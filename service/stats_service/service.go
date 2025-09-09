package stats_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/stats_model"
)

func GetPersonPageStats(farmId uint32) ([]entity_public.StatCard, *entity_public.Toast) {
	sm := stats_model.GetStatsModel()
	stats := make([]entity_public.StatCard, 0, 5)

	topSupplier, err := sm.GetTopSupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return []entity_public.StatCard{}, &toast
	}
	stats = append(stats, topSupplier)

	topBuyer, err := sm.GetTopBuyer(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return []entity_public.StatCard{}, &toast
	}
	stats = append(stats, topBuyer)

	mostFrequent, err := sm.GetMostFrequentSupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return []entity_public.StatCard{}, &toast
	}
	stats = append(stats, mostFrequent)

	bestQuality, err := sm.GetBestQualitySupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return []entity_public.StatCard{}, &toast
	}
	stats = append(stats, bestQuality)

	worstQuality, err := sm.GetWorstQualitySupplier(farmId)
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao buscar estatísticas", "")
		return []entity_public.StatCard{}, &toast
	}
	stats = append(stats, worstQuality)

	return stats, nil
}
