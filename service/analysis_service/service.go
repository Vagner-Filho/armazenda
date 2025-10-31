package analysis_service

import (
	"armazenda/entity/public"
	"armazenda/model/analysis_model"
)

func GetProductiveFields(farmID uint32) (*entity_public.ProductiveFields, *entity_public.Toast) {
	model := analysis_model.GetAnalysisModel()
	nominal, err := model.GetNominalMostProductiveField(farmID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar o campo produtivo nominal", "")
		return nil, &toast
	}

	relative, err := model.GetRelativeMostProductiveField(farmID)
	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro interno ao buscar o campo produtivo relativo", "")
		return nil, &toast
	}

	return &entity_public.ProductiveFields{
		Nominal:  nominal,
		Relative: relative,
	}, nil
}
