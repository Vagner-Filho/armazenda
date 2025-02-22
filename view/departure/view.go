package departure_view

import (
	entity_public "armazenda/entity/public"
	buyer_service "armazenda/service/buyer"
	crop_service "armazenda/service/crop"
	"armazenda/service/departure_service"
	"armazenda/service/vehicle_service"
)

type DepartureForm struct {
	Vehicles  []entity_public.Vehicle
	Buyers    []entity_public.BuyerDisplay
	Departure entity_public.Departure
	Crops     []entity_public.Crop
}

func GetNewDepartureForm() (DepartureForm, []*entity_public.Toast) {
	vehicles, vtoast := vehicle_service.GetVehicles()
	crops, ctoast := crop_service.GetCrops()
	buyers, btoast := buyer_service.GetBuyers()

	return DepartureForm{
		Vehicles:  vehicles,
		Buyers:    buyers,
		Departure: entity_public.Departure{},
		Crops:     crops,
	}, []*entity_public.Toast{vtoast, ctoast, btoast}
}

func GetExistingDepartureForm(departureId uint32) (DepartureForm, []*entity_public.Toast) {
	form, toasts := GetNewDepartureForm()
	departure, toast := departure_service.GetDeparture(departureId)

	form.Departure = departure
	toasts = append(toasts, toast)
	return form, toasts
}

type departureFilter struct {
	Buyers   []entity_public.BuyerDisplay
	Vehicles []entity_public.Vehicle
}
type departureContent struct {
	Departures []entity_public.DisplayDeparture
	Filters    departureFilter
}

func GetDepartureContent() (departureContent, []*entity_public.Toast) {
	departures, dtoast := departure_service.GetDisplayDepartures()
	vehicles, vtoast := vehicle_service.GetVehicles()
	buyers, btoast := buyer_service.GetBuyers()

	return departureContent{
		Departures: departures,
		Filters: departureFilter{
			Buyers:   buyers,
			Vehicles: vehicles,
		},
	}, []*entity_public.Toast{dtoast, vtoast, btoast}
}
