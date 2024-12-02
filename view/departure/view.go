package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/vehicle_service"
	"armazenda/utils"
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
