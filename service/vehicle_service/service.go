package vehicle_service

import "armazenda/model/vehicle_model"

func GetPlates() []vehicle_model.VehiclePlate {
    return vehicle_model.GetPlates()
}
