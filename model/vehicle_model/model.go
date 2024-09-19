package vehicle_model

import "slices"

type Vehicle struct {
	Plate string
	Name  string
}

var vehicles = []Vehicle{
	{
		Name:  "Mercedão 1315",
		Plate: "APB 7059",
	},
	{
		Name:  "Scania",
		Plate: "JJK 7821",
	},
	{
		Name:  "",
		Plate: "UOU 1280",
	},
}

func GetVehicles() []Vehicle {
	return vehicles
}

func AddVehicle(v Vehicle) (Vehicle, bool) {
    var contains = slices.Contains(vehicles, v)
    if contains {
        return v, contains
    }

	vehicles = append(vehicles, Vehicle{Plate: v.Plate, Name: v.Name})
    return v, contains
}
