package vehicle_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/vehicle_model"
)

func GetVehiclesByFarm(farm uint32) ([]entity_public.Vehicle, *entity_public.Toast) {
	vModel := vehicle_model.GetVehicleModel()

	vehicles, err := vModel.GetVehiclesByFarm(farm)
	if err != nil {
		toast := entity_public.GetErrorToast(err.Error(), "")
		return []entity_public.Vehicle{}, &toast
	}

	return vehicles, nil
}

func GetVehicle(id uint16) (entity_public.Vehicle, error) {
	vModel := vehicle_model.GetVehicleModel()
	return vModel.GetVehicle(id)
}

func AddVehicle(v entity_public.Vehicle) (entity_public.Vehicle, error) {
	vModel := vehicle_model.GetVehicleModel()

	vehicle, addErr := vModel.AddVehicle(v)
	if addErr != nil {
		return entity_public.Vehicle{}, addErr
	}
	return vehicle, nil
}
