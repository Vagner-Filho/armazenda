package departure_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/crop"
	"armazenda/service/departure_service"
	"armazenda/service/person"
	product_service "armazenda/service/product"
	"armazenda/service/vehicle_service"
	"armazenda/view"
	"armazenda/view/filters"
	"fmt"
)

func BuildDepartureChips(rf entity_public.DepartureFilter, farm uint32) []filters.ChipEntry {
	chips := []filters.ChipEntry{}

	if rf.Product != 0 {
		products, _ := product_service.GetProducts()
		for _, p := range products {
			if p.Id == rf.Product {
				chips = append(chips, filters.ChipEntry{Key: "product", Label: "Grão", Value: p.Name})
				break
			}
		}
	}
	if rf.Person != "" {
		if rf.Person == "NULL" {
			chips = append(chips, filters.ChipEntry{Key: "person", Label: "Pessoa", Value: "Própria"})
		} else {
			people, _ := person_service.GetPeopleByFarm(farm)
			for _, p := range people {
				if p.Id != nil && fmt.Sprintf("%v", *p.Id) == rf.Person {
					chips = append(chips, filters.ChipEntry{Key: "person", Label: "Pessoa", Value: p.Name})
					break
				}
			}
		}
	}
	if rf.VehiclePlate != "" {
		vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
		display := rf.VehiclePlate
		for _, v := range vehicles {
			if fmt.Sprintf("%v", v.Id) == rf.VehiclePlate {
				display = v.Plate
				if v.Name != "" {
					display = v.Plate + " | " + v.Name
				}
				break
			}
		}
		chips = append(chips, filters.ChipEntry{Key: "vehiclePlate", Label: "Veículo", Value: display})
	}
	if rf.NetWeightMin != 0 {
		chips = append(chips, filters.ChipEntry{Key: "netWeightMin", Label: "Peso mín.", Value: fmt.Sprintf("%.0f kg", rf.NetWeightMin)})
	}
	if rf.NetWeightMax != 0 {
		chips = append(chips, filters.ChipEntry{Key: "netWeightMax", Label: "Peso máx.", Value: fmt.Sprintf("%.0f kg", rf.NetWeightMax)})
	}
	if !rf.DepartureDateMin.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "departureDateMin", Label: "Início", Value: rf.DepartureDateMin.Format("02/01/2006 15:04")})
	}
	if !rf.DepartureDateMax.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "departureDateMax", Label: "Fim", Value: rf.DepartureDateMax.Format("02/01/2006 15:04")})
	}

	return chips
}

type departureFilters struct {
	view.BaseTemplateData
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
	People   []entity_public.PersonOption
}

type DepartureContent struct {
	view.BaseTemplateData
	Departures   []entity_public.DisplayDeparture
	Drafts       []entity_public.DisplayDepartureDraft
	Filters      departureFilters
	AppliedChips filters.FilterChips
	NoContent    bool
	CurrentPage  int
	TotalPages   int
	NextPage     int
	PrevPage     int
	HasNext      bool
	HasPrev      bool
}

type DepartureFilterApplyResponse struct {
	view.BaseTemplateData
	Chips       filters.FilterChips
	Departures  []entity_public.DisplayDeparture
	TotalPages  int
	CurrentPage int
	HasNext     bool
	HasPrev     bool
	NextPage    int
	PrevPage    int
	NoResults   bool
}

type DepartureClearedView struct {
	view.BaseTemplateData
	Form    departureFilters
	Chips   filters.FilterChips
	Content DepartureContent
}

func BuildDepartureFilterApplyResponse(filter entity_public.DepartureFilter, departures []entity_public.DisplayDeparture, total int, page int, farm uint32) DepartureFilterApplyResponse {
	pageSize := 10
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
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
	return DepartureFilterApplyResponse{
		Chips:       filters.FilterChips{Items: BuildDepartureChips(filter, farm), OOB: true},
		Departures:  departures,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		NoResults:   len(departures) == 0,
	}
}

func GetClearedDepartureView(farm uint32) DepartureClearedView {
	form := GetFiltersForm(farm)
	content, _ := GetDepartureContent(farm, 1)
	return DepartureClearedView{
		Form:    form,
		Chips:   filters.FilterChips{Items: []filters.ChipEntry{}, OOB: true},
		Content: content,
	}
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
		Departures:   departures,
		Drafts:       drafts,
		NoContent:    len(departures) == 0 && len(drafts) == 0,
		Filters:      GetFiltersForm(farmId),
		AppliedChips: filters.FilterChips{Items: []filters.ChipEntry{}},
		CurrentPage:  page,
		TotalPages:   totalPages,
		NextPage:     nextPage,
		PrevPage:     prevPage,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
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
