package departure_model

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/vehicle_model"
	"slices"
)

var vehicles = vehicle_model.GetVehicles()
var departures = []entity_public.Departure{
	{
		Manifest:      0,
		DepartureDate: 1726967334411,
		Product:       0,
		VehiclePlate:  vehicles[0].Plate,
		Weight:        20392,
		Buyer:         "0-12345678901",
	},
	{
		Manifest:      1,
		DepartureDate: 1726967334411,
		Product:       0,
		VehiclePlate:  vehicles[1].Plate,
		Weight:        20392,
		Buyer:         "0-12345678901",
	},
	{
		Manifest:      2,
		DepartureDate: 1726967334411,
		Product:       0,
		VehiclePlate:  vehicles[2].Plate,
		Weight:        20392,
		Buyer:         "0-12345678901",
	},
}

func GetDepartures() []entity_public.Departure {
	return departures
}

func GetDeparture(manifest uint32) *entity_public.Departure {
	index := slices.IndexFunc(departures, func(d entity_public.Departure) bool {
		return d.Manifest == manifest
	})
	if index > -1 {
		return &departures[index]
	}
	return nil
}

func AddDeparture(bd entity_public.Departure) entity_public.Departure {
	lastManifest := departures[len(departures)-1]
	bd.Manifest = lastManifest.Manifest + 1
	departures = append(departures, bd)
	return departures[len(departures)-1]
}

func PutDeparture(d entity_public.Departure) (entity_public.Departure, bool) {
	dIndex := slices.IndexFunc(departures, func(id entity_public.Departure) bool {
		return d.Manifest == id.Manifest
	})

	if dIndex == -1 {
		return entity_public.Departure{}, true
	}

	departures = slices.Replace(departures, dIndex, dIndex+1, d)

	return d, false
}

func DeleteDeparture(manifest uint32) int {
	dIndex := slices.IndexFunc(departures, func(d entity_public.Departure) bool {
		return manifest == d.Manifest
	})

	if dIndex > -1 {
		departures = slices.Delete(departures, dIndex, dIndex+1)
	}

	return dIndex
}
