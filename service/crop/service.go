package crop_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
)

func GetCrops() ([]entity_public.Crop, *entity_public.Toast) {
	cModel, _ := crop_model.GetCropModel()
	crops, err := cModel.GetCrops()

	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.Crop{}, &toast
	}

	return crops, nil
}
