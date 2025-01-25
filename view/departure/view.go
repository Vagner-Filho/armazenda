package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/service/vehicle_service"
)

type departureFormView struct {
	Vehicles []entity_public.Vehicle
	Buyers   []entity_public.Buyer
	entity_public.Departure
}

func GetNewDepartureForm() departureFormView {
	vehicles, _ := vehicle_service.GetVehicles()
	return departureFormView{
		Vehicles:  vehicles,
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
	Vehicles []entity_public.Vehicle
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

	vehicles, _ := vehicle_service.GetVehicles()
	return departureContent{
		Departures: readable,
		Filters: departureFilter{
			Buyers:   buyer_model.GetBuyers(),
			Vehicles: vehicles,
		},
	}
}
