package vehicle_service

import "armazenda/model/vehicle_model"

func GetVehicles() []vehicle_model.Vehicle {
    return vehicle_model.GetVehicles()
}

func AddVehicle(v vehicle_model.Vehicle) (vehicle_model.Vehicle, *string) {
    var vehicle, contains = vehicle_model.AddVehicle(v)
    if contains {
        var containsMessage string = "Veículo com a placa " + vehicle.Plate + " já existe."
        return vehicle, &containsMessage
    }
    return vehicle, nil
}
