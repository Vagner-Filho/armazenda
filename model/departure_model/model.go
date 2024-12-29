package departure_model

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/vehicle_model"
	"armazenda/utils"
	"slices"
	"time"
)

var vehicles = vehicle_model.GetVehicles()

var availableDepartureFilters = map[string]func(e entity_public.Departure, ef entity_public.DepartureFilter) bool{
	"DepartureDateMin": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		departureDate, departureDateError := time.Parse(utils.TimeLayout, d.DepartureDate)
		departureMin, departureFilterDateError := time.Parse(utils.TimeLayout, df.DepartureDateMin)
		if departureDateError != nil || departureFilterDateError != nil {
			return false
		}
		return departureMin.Before(departureDate)
	},
	"DepartureDateMax": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		departureDate, departureDateError := time.Parse(utils.TimeLayout, d.DepartureDate)
		departureMax, departureFilterDateError := time.Parse(utils.TimeLayout, df.DepartureDateMax)
		if departureDateError != nil || departureFilterDateError != nil {
			return false
		}
		return departureMax.After(departureDate)
	},
	"VehiclePlate": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.VehiclePlate == df.VehiclePlate
	},
	"Product": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Product == df.Product
	},
	"WeightMin": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Weight >= df.WeightMin
	},
	"WeightMax": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Weight <= df.WeightMax
	},
	"Buyer": func(d entity_public.Departure, df entity_public.DepartureFilter) bool {
		return d.Buyer == df.Buyer
	},
}
var departures = []entity_public.Departure{
	{
		Manifest:      0,
		DepartureDate: time.Now().AddDate(0, -1, -3).Format(utils.TimeLayout),
		Product:       2,
		VehiclePlate:  vehicles[0].Plate,
		Weight:        5000,
		Buyer:         "0-12345678901",
	},
	{
		Manifest:      1,
		DepartureDate: time.Now().AddDate(0, 0, -15).Format(utils.TimeLayout),
		Product:       2,
		VehiclePlate:  vehicles[1].Plate,
		Weight:        10000,
		Buyer:         "0-12345678901",
	},
	{
		Manifest:      2,
		DepartureDate: time.Now().AddDate(0, 0, -7).Format(utils.TimeLayout),
		Product:       1,
		VehiclePlate:  vehicles[2].Plate,
		Weight:        15000,
		Buyer:         "0-12345678901",
	},
}

func FilterDepartures(filter entity_public.DepartureFilter) ([]entity_public.Departure, error) {
	var filteredDepartures []entity_public.Departure

	filters := filter.GetFilters(availableDepartureFilters)
	for _, departure := range departures {
		include := true
		for f := range filters {
			fff := filters[f]

			if fff == nil {
				continue
			}

			include = fff(departure, filter)
			if !include {
				break
			}
		}
		if include {
			filteredDepartures = append(filteredDepartures, departure)
		}
	}
	return filteredDepartures, nil
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
