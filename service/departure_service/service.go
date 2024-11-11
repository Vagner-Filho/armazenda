package departure_service

import (
	"armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/utils"
)

func MakeReadableDeparture(gd entity_public.Departure) ReadableDeparture {
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

func GetDeparture(manifest uint32) (entity_public.Departure, bool) {
	departure := departure_model.GetDeparture(manifest)
	if departure == nil {
		return entity_public.Departure{}, true
	}
	return *departure, false
}

func AddDeparture(bd entity_public.Departure) entity_public.Departure {
	return departure_model.AddDeparture(bd)
}

func PutDeparture(d entity_public.Departure) (entity_public.Departure, bool) {
	return departure_model.PutDeparture(d)
}

func DeleteDeparture(manifest uint32) int {
	return departure_model.DeleteDeparture(manifest)
}
