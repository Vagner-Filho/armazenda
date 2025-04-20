package crop_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
)

func GetCropsByFarm(farm uint32) ([]entity_public.Crop, *entity_public.Toast) {
	cModel, _ := crop_model.GetCropModel()
	crops, err := cModel.GetCropsByFarm(farm)

	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.Crop{}, &toast
	}

	return crops, nil
}
