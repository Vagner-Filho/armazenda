package farm_config_service

import (
	"armazenda/entity/public"
	"armazenda/model/farm_config_model"
)

func UpsertFarmConfig(config *entity_public.Farm) error {
	m := farm_config_model.GetFarmConfigModel()
	return m.UpsertFarmConfig(config)
}

func GetFarmConfig(farmID uint32) (*entity_public.Farm, error) {
	m := farm_config_model.GetFarmConfigModel()
	return m.GetFarmConfig(farmID)
}

func SetFarmHumidityProgression(farmID uint32, progressionID *uint32) error {
	m := farm_config_model.GetFarmConfigModel()
	return m.SetFarmHumidityProgression(farmID, progressionID)
}
