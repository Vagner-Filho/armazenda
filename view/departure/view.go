package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/crop"
	"armazenda/service/departure_service"
	"armazenda/service/person"
	"armazenda/service/vehicle_service"
)

type departureFilters struct {
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
	People   []entity_public.PersonOption
}

type DepartureContent struct {
	Departures []entity_public.DisplayDeparture
	Drafts     []entity_public.DisplayDepartureDraft
	Filters    departureFilters
	NoContent  bool
}

func GetFiltersForm(farm uint32) departureFilters {
	crops, _ := crop_service.GetCropsByFarm(farm)
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
	people, _ := person_service.GetPeopleByFarm(farm)

	return departureFilters{
		Vehicles: vehicles,
		Crops:    crops,
		People:   people,
	}
}

func GetDepartureContent(farmId uint32) (DepartureContent, []*entity_public.Toast) {
	departures, toast := departure_service.GetDisplayDepartures(farmId)
	drafts, draftToast := departure_service.GetAllDepartureDrafts(farmId)

	return DepartureContent{
		Departures: departures,
		Drafts:     drafts,
		NoContent:  len(departures) == 0 && len(drafts) == 0,
		Filters:    GetFiltersForm(farmId),
	}, []*entity_public.Toast{toast, draftToast}
}

type DepartureDraftForm struct {
	Vehicles        []entity_public.Vehicle
	SelectedVehicle string
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	People          []entity_public.PersonOption
	SelectedPerson  *uint32
	Draft           entity_public.DepartureDraft
}

func GetDepartureDraftForm(farm uint32) (DepartureDraftForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehiclesByFarm(farm)
	crops, cToast := crop_service.GetCropsByFarm(farm)
	people, pToast := person_service.GetPeopleByFarm(farm)

	return DepartureDraftForm{
		Vehicles: vehicles,
		Crops:    crops,
		People:   people,
	}, []*entity_public.Toast{vToast, cToast, pToast}
}

type DepartureForm struct {
	Vehicles        []entity_public.Vehicle
	SelectedVehicle string
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	People          []entity_public.PersonOption
	SelectedPerson  *uint32
	Departure       entity_public.DepartureDTO
}

func GetNewDepartureForm(farm uint32) (DepartureForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehiclesByFarm(farm)
	crops, cToast := crop_service.GetCropsByFarm(farm)
	people, pToast := person_service.GetPeopleByFarm(farm)

	return DepartureForm{
		Vehicles: vehicles,
		Crops:    crops,
		People:   people,
	}, []*entity_public.Toast{vToast, cToast, pToast}
}

func GetExistingDepartureForm(departureId uint32, farm uint32) (DepartureForm, []*entity_public.Toast) {
	formFields, toasts := GetNewDepartureForm(farm)
	departure, toast := departure_service.GetDeparture(departureId)
	formFields.SelectedCrop = departure.Crop
	formFields.SelectedVehicle = departure.VehiclePlate
	formFields.SelectedPerson = departure.Person

	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Departure = departure.ToDTO()
	return formFields, toasts
}

