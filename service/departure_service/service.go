package departure_service

import (
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/utils"
)

func MakeReadableDeparture(gd departure_model.Departure) ReadableDeparture {
	return ReadableDeparture{
		Manifest:      gd.Manifest,
		Product:       entry_model.GrainMap[gd.Product],
		VehiclePlate:  gd.VehiclePlate,
		Weight:        gd.Weight,
		DepartureDate: utils.GetReadableDate(gd.DepartureDate),
	}
}

func GetDepartures() []ReadableDeparture {
	var readable = []ReadableDeparture{}

	departures := departure_model.GetDepartures()

	for _, gd := range departures {
		readable = append(readable, MakeReadableDeparture(gd))
	}
	return readable
}

func GetDeparture(manifest uint32) (departure_model.Departure, bool) {
	departure := departure_model.GetDeparture(manifest)
	if departure == nil {
		return departure_model.Departure{}, true
	}
	return *departure, false
}

func AddDeparture(bd departure_model.BaseDeparture) departure_model.Departure {
	return departure_model.AddDeparture(bd)
}

func PutDeparture(d departure_model.Departure) (departure_model.Departure, bool) {
    return departure_model.PutDeparture(d)
}
