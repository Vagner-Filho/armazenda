package departure_service

import (
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"time"
)

type ReadableDeparture struct {
	Manifest      uint32
	DepartureDate string
	Product       string
	VehiclePlate  string
	Weight        float64
}

func MakeReadableDeparture(gd departure_model.Departure) ReadableDeparture {
	return ReadableDeparture{
		Manifest:      gd.Manifest,
		Product:       entry_model.GrainMap[gd.Product],
		VehiclePlate:  gd.VehiclePlate,
		Weight:        gd.Weight,
		DepartureDate: time.UnixMilli(gd.DepartureDate).Format("02/Jan/2006 - 03:04"),
	}
}

func GetDepartures() []ReadableDeparture {
    var readable = []ReadableDeparture {}

    departures := departure_model.GetDepartures()

    for _, gd := range departures {
        readable = append(readable, MakeReadableDeparture(gd))
    }
	return readable
}
