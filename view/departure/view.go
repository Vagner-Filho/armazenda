package departure_view

import (
	entity_public "armazenda/entity/public"
	crop_service "armazenda/service/crop"
	"armazenda/service/departure_service"
	person_service "armazenda/service/person"
	"armazenda/service/vehicle_service"
)

type DepartureForm struct {
	Vehicles        []entity_public.Vehicle
	People          []entity_public.PersonOption
	Departure       entity_public.DepartureDTO
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	SelectedVehicle string
	SelectedPerson  uint8
}

func GetNewDepartureForm(farm uint32) (DepartureForm, []*entity_public.Toast) {
	vehicles, vtoast := vehicle_service.GetVehiclesByFarm(farm)
	crops, ctoast := crop_service.GetCropsByFarm(farm)
	people, btoast := person_service.GetPeopleByFarm(farm)

	return DepartureForm{
		Vehicles:  vehicles,
		People:    people,
		Departure: entity_public.DepartureDTO{},
		Crops:     crops,
	}, []*entity_public.Toast{vtoast, ctoast, btoast}
}

func GetExistingDepartureForm(departureId uint32, farm uint32) (DepartureForm, []*entity_public.Toast) {
	form, toasts := GetNewDepartureForm(farm)
	departure, toast := departure_service.GetDeparture(departureId)

	form.Departure = departure.ToDTO()
	form.SelectedCrop = departure.Crop
	form.SelectedVehicle = departure.VehiclePlate
	form.SelectedPerson = uint8(departure.Person)

	toasts = append(toasts, toast)
	return form, toasts
}

type departureFilter struct {
	People   []entity_public.PersonOption
	Vehicles []entity_public.Vehicle
}
type departureContent struct {
	Departures []entity_public.DisplayDeparture
	Filters    departureFilter
	NoContent  bool
}

func GetDepartureContent(farm uint32) (departureContent, []*entity_public.Toast) {
	departures, dtoast := departure_service.GetDisplayDepartures(farm)
	vehicles, vtoast := vehicle_service.GetVehiclesByFarm(farm)
	people, btoast := person_service.GetPeopleByFarm(farm)

	return departureContent{
		Departures: departures,
		Filters: departureFilter{
			People:   people,
			Vehicles: vehicles,
		},
		NoContent: len(departures) == 0,
	}, []*entity_public.Toast{dtoast, vtoast, btoast}
}
