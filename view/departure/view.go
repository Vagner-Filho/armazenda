package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/vehicle_service"
)

type departureFormView struct {
	Vehicles []vehicle_model.Vehicle
	Buyers   []entity_public.Buyer
	entity_public.Departure
}

func GetNewDepartureForm() departureFormView {
	return departureFormView{
		Vehicles:  vehicle_service.GetVehicles(),
		Buyers:    buyer_model.GetBuyers(),
		Departure: entity_public.Departure{},
	}
}

type ReadableDeparture struct {
	Manifest      uint32
	DepartureDate string
	Product       string
	VehiclePlate  string
	Weight        float64
}

func MakeReadableDeparture(gd entity_public.Departure) ReadableDeparture {
	return ReadableDeparture{
		Manifest:      gd.Manifest,
		Product:       entry_model.GrainMap[gd.Product],
		VehiclePlate:  gd.VehiclePlate,
		Weight:        gd.Weight,
		DepartureDate: gd.DepartureDate,
	}
}

type departureFilter struct {
	Buyers   []entity_public.Buyer
	Vehicles []vehicle_model.Vehicle
}
type departureContent struct {
	Departures []ReadableDeparture
	Filters    departureFilter
}

func GetDepartureContent() departureContent {
	var readable = []ReadableDeparture{}

	departures := departure_model.GetDepartures()

	for _, gd := range departures {
		readable = append(readable, MakeReadableDeparture(gd))
	}
	return departureContent{
		Departures: readable,
		Filters: departureFilter{
			Buyers:   buyer_model.GetBuyers(),
			Vehicles: vehicle_model.GetVehicles(),
		},
	}
}
