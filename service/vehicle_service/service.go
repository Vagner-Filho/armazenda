package vehicle_service

import "armazenda/model/vehicle_model"

func GetVehicles() []vehicle_model.Vehicle {
    return vehicle_model.GetVehicles()
}
