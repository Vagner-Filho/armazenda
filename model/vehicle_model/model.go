package vehicle_model

type Vehicle struct {
    Plate string
    Name string
}

var vehicles = []Vehicle{
    {
        Name: "Mercedão 1315",
        Plate: "APB 7059",
    },
}

func GetVehicles() []Vehicle {
    return vehicles
}

func AddVehicle(v Vehicle) {
    vehicles = append(vehicles, Vehicle{ Plate: v.Plate, Name: v.Name })
}
