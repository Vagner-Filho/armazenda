package departure_model

import "slices"

var departures = []Departure{
	{
		Manifest: 1,
		BaseDeparture: BaseDeparture{
			DepartureDate: 1726967334411,
			Product:       0,
			VehiclePlate:  "OPA 2312",
			Weight:        20392,
		},
	},
	{
		Manifest: 2,
		BaseDeparture: BaseDeparture{
			DepartureDate: 1726967334411,
			Product:       1,
			VehiclePlate:  "OPA 2312",
			Weight:        12398,
		},
	},
	{
		Manifest: 3,
		BaseDeparture: BaseDeparture{
			DepartureDate: 1726967334411,
			Product:       0,
			VehiclePlate:  "OPA 2312",
			Weight:        40242,
		},
	},
}

func GetDepartures() []Departure {
	return departures
}

func GetDeparture(manifest uint32) *Departure {
    index := slices.IndexFunc(departures, func(d Departure) bool {
        return d.Manifest == manifest
    })
    if index > -1 {
        return &departures[index]
    }
    return nil
}

func AddDeparture(bd BaseDeparture) Departure {
	lastManifest := departures[len(departures)-1]
	departures = append(departures, Departure{
		Manifest: lastManifest.Manifest + 1,
		BaseDeparture: BaseDeparture{
			DepartureDate: bd.DepartureDate,
			Product:       bd.Product,
			Weight:        bd.Weight,
			VehiclePlate:  bd.VehiclePlate,
		},
	})
    return departures[len(departures)-1]
}

func PutDeparture(d Departure) (Departure, bool) {
    dIndex := slices.IndexFunc(departures, func(id Departure) bool {
        return d.Manifest == id.Manifest
    })

    if dIndex == -1 {
        return Departure{}, true
    }

    departures = slices.Replace(departures, dIndex, dIndex + 1, d)

    return d, false
}
