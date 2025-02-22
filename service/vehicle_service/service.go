package vehicle_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/vehicle_model"
)

func GetVehicles() ([]entity_public.Vehicle, *entity_public.Toast) {
	vModel, _ := vehicle_model.GetVehicleModel()

	vehicles, err := vModel.GetVehicles()
	if err != nil {
		toast := entity_public.GetErrorToast(err.Error(), "")
		return []entity_public.Vehicle{}, &toast
	}

	return vehicles, nil
}

func GetVehicle(plate string) (entity_public.Vehicle, error) {
	vModel, _ := vehicle_model.GetVehicleModel()
	return vModel.GetVehicle(plate)
}

func AddVehicle(v entity_public.Vehicle) (entity_public.Vehicle, error) {
	vModel, modelErr := vehicle_model.GetVehicleModel()
	if modelErr != nil {
		return entity_public.Vehicle{}, modelErr
	}

	vehicle, addErr := vModel.AddVehicle(v)
	if addErr != nil {
		return entity_public.Vehicle{}, addErr
	}
	return vehicle, nil
}
