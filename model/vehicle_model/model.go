package vehicle_model


type VehiclePlate struct {
    id uint8
    plateName string
}

var vehicle_plate = []VehiclePlate{
    {
        id: 0,
        plateName: "APB 7059",
    },
}

func GetPlates() []VehiclePlate {
    return vehicle_plate
}

func AddPlate(plateName string) {
    lastPlate := vehicle_plate[len(vehicle_plate) - 1]
    vehicle_plate = append(vehicle_plate, VehiclePlate{ id: lastPlate.id + 1, plateName: plateName })
}
