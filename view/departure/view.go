package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/crop"
	"armazenda/service/departure_service"
	"armazenda/service/person"
	"armazenda/service/vehicle_service"
	"armazenda/view"
)

type departureFilters struct {
	view.BaseTemplateData
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
	People   []entity_public.PersonOption
}

type DepartureContent struct {
	view.BaseTemplateData
	Departures  []entity_public.DisplayDeparture
	Drafts      []entity_public.DisplayDepartureDraft
	Filters     departureFilters
	NoContent   bool
	CurrentPage int
	TotalPages  int
	NextPage    int
	PrevPage    int
	HasNext     bool
	HasPrev     bool
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

func GetDepartureContent(farmId uint32, page int) (DepartureContent, []*entity_public.Toast) {
	if page == 0 {
		page = 1
	}
	pageSize := 10
	departures, totalDepartures, toast := departure_service.GetDisplayDepartures(farmId, page)
	drafts, draftToast := departure_service.GetAllDepartureDrafts(farmId)

	totalPages := 0
	if totalDepartures > 0 {
		totalPages = (totalDepartures + pageSize - 1) / pageSize
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	nextPage := page + 1
	if !hasNext {
		nextPage = page
	}

	prevPage := page - 1
	if !hasPrev {
		prevPage = page
	}

	return DepartureContent{
		Departures:  departures,
		Drafts:      drafts,
		NoContent:   len(departures) == 0 && len(drafts) == 0,
		Filters:     GetFiltersForm(farmId),
		CurrentPage: page,
		TotalPages:  totalPages,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}, []*entity_public.Toast{toast, draftToast}
}

type DepartureDraftForm struct {
	view.BaseTemplateData
	Vehicles          []entity_public.Vehicle
	Crops             []entity_public.Crop
	People            []entity_public.PersonOption
	Draft             entity_public.DepartureDraft
	SelectedRecipient *uint32
	SelectedOrigin    *uint32
	SelectedVehicle   uint16
	SelectedCrop      uint8
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
	view.BaseTemplateData
	Vehicles          []entity_public.Vehicle
	SelectedVehicle   uint16
	Crops             []entity_public.Crop
	SelectedCrop      uint8
	People            []entity_public.PersonOption
	SelectedRecipient *uint32
	SelectedOrigin    *uint32
	Departure         entity_public.DepartureDTO
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
	formFields.SelectedVehicle = departure.Vehicle
	if departure.Origin != nil {
		for i, p := range formFields.People {
			if p.Id != nil && *p.Id == *departure.Origin {
				formFields.SelectedOrigin = formFields.People[i].Id
				break
			}
		}
	}
	if departure.Recipient != nil {
		for i, p := range formFields.People {
			if p.Id != nil && *p.Id == *departure.Recipient {
				formFields.SelectedRecipient = formFields.People[i].Id
				break
			}
		}
	}
	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Departure = departure.ToDTO()
	return formFields, toasts
}

func GetDepartureFormFromDraft(draftId uint32, farm uint32) (DepartureForm, []*entity_public.Toast) {
	formFields, toasts := GetNewDepartureForm(farm)
	draft, toast := departure_service.GetDepartureDraft(draftId)

	if toast != nil {
		toasts = append(toasts, toast)
	}

	if draft.Id != 0 {
		formFields.SelectedCrop = uint8(draft.Crop)
		formFields.SelectedVehicle = draft.Vehicle
		if draft.Recipient != nil {
			for i, p := range formFields.People {
				if p.Id != nil && *p.Id == *draft.Recipient {
					formFields.SelectedRecipient = formFields.People[i].Id
					break
				}
			}
		}
		if draft.Origin != nil {
			for i, p := range formFields.People {
				if p.Id != nil && *p.Id == *draft.Origin {
					formFields.SelectedOrigin = formFields.People[i].Id
					break
				}
			}
		}

		if draft.Tare != nil {
			tare, _ := draft.Tare.Float64()
			formFields.Departure.Tare = tare
		}
	}

	return formFields, toasts
}

func GetDepartureDrafts(farmId uint32) ([]entity_public.DisplayDepartureDraft, *entity_public.Toast) {
	drafts, draftToast := departure_service.GetAllDepartureDrafts(farmId)
	return drafts, draftToast
}

// DepartureListItemView wraps Departure for template rendering with CSP nonce
type DepartureListItemView struct {
	view.BaseTemplateData
	Departure entity_public.DisplayDeparture
}

// DepartureDraftListItemView wraps DepartureDraft for template rendering with CSP nonce
type DepartureDraftListItemView struct {
	view.BaseTemplateData
	Draft entity_public.DisplayDepartureDraft
}
