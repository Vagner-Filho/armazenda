package nfe_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/nfe_model"
)

func UpsertFarmConfig(config entity_public.FarmConfig) (entity_public.FarmConfig, entity_public.Toast) {
	nfeModel := nfe_model.GetNFeModel()
	modelErr := nfeModel.UpsertFarmConfig(config)
	if modelErr == nil {
		return config, entity_public.GetSuccessToast("Configuração Salva", "")
	}
	if modelErr.IsServerErr {
		return config, entity_public.GetErrorToast("No momento não foi possível salvar a configuração", "estamos investigando")
	}
	return config, entity_public.GetWarningToast(modelErr.Message, "")
}
