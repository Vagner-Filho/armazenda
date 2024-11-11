package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
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
