package entry_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	crop_service "armazenda/service/crop"
	"armazenda/service/entry_service"
	field_service "armazenda/service/field"
	person_service "armazenda/service/person"
	product_service "armazenda/service/product"
	"armazenda/service/vehicle_service"
	"armazenda/view"
	"armazenda/view/filters"
	"fmt"
)

func BuildEntryChips(rf entity_public.EntryFilter, farm uint32) []filters.ChipEntry {
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
	if rf.Crop != 0 {
		crops, _ := crop_service.GetCropsByFarm(farm)
		for _, c := range crops {
			if c.Id == rf.Crop {
				chips = append(chips, filters.ChipEntry{Key: "crop", Label: "Safra", Value: c.Name})
				break
			}
		}
	}
	if rf.Field != 0 {
		fields, _ := field_service.GetFieldsByFarm(farm)
		for _, f := range fields {
			if f.Id == rf.Field {
				chips = append(chips, filters.ChipEntry{Key: "field", Label: "Talhão", Value: f.Name})
				break
			}
		}
	}
	if rf.Vehicle != "" {
		vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)
		display := rf.Vehicle
		for _, v := range vehicles {
			if fmt.Sprintf("%v", v.Id) == rf.Vehicle {
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
	if !rf.ArrivalDateMin.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "arrivalDateMin", Label: "Início", Value: rf.ArrivalDateMin.Format("02/01/2006 15:04")})
	}
	if !rf.ArrivalDateMax.IsZero() {
		chips = append(chips, filters.ChipEntry{Key: "arrivalDateMax", Label: "Fim", Value: rf.ArrivalDateMax.Format("02/01/2006 15:04")})
	}

	return chips
}

type entryFilters struct {
	view.BaseTemplateData
	Fields   []entity_public.Field
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
}

type entryContent struct {
	view.BaseTemplateData
	Entries      []entity_public.DisplayEntry
	Drafts       []entity_public.DisplayEntryDraft
	Filters      entryFilters
	AppliedChips filters.FilterChips
	NoContent    bool
	CurrentPage  int
	TotalPages   int
	NextPage     int
	PrevPage     int
	HasNext      bool
	HasPrev      bool
}

type EntryFilterApplyResponse struct {
	view.BaseTemplateData
	Chips      filters.FilterChips
	Entries    []entity_public.DisplayEntry
	TotalPages int
	CurrentPage int
	HasNext    bool
	HasPrev    bool
	NextPage   int
	PrevPage   int
	NoResults  bool
}

type EntryClearedView struct {
	view.BaseTemplateData
	Form    entryFilters
	Chips   filters.FilterChips
	Content entryContent
}

func BuildEntryFilterApplyResponse(filter entity_public.EntryFilter, entries []entity_public.DisplayEntry, total int, page int, farm uint32) EntryFilterApplyResponse {
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
	return EntryFilterApplyResponse{
		Chips:       filters.FilterChips{Items: BuildEntryChips(filter, farm), OOB: true},
		Entries:     entries,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		NoResults:   len(entries) == 0,
	}
}

func GetClearedEntryView(farm uint32) EntryClearedView {
	form := GetFiltersForm(farm)
	content := GetEntryContent(farm, 1)
	return EntryClearedView{
		Form:    form,
		Chips:   filters.FilterChips{Items: []filters.ChipEntry{}, OOB: true},
		Content: content,
	}
}

func GetAllEntryDisplay(farm uint32, page int) ([]entity_public.DisplayEntry, int) {
	eModel := entry_model.GetEntryModel()
	entries, totalEntries, getDataErr := eModel.GetDisplayEntriesByFarm(farm, page)
	if getDataErr != nil {
		return []entity_public.DisplayEntry{}, 0
	}
	return entries, totalEntries
}

func GetAllEntryDraftsDisplay(farm uint32) []entity_public.DisplayEntryDraft {
	eModel := entry_model.GetEntryModel()
	drafts, getDataErr := eModel.GetEntryDraftsByFarm(farm)
	if getDataErr != nil {
		return []entity_public.DisplayEntryDraft{}
	}
	return drafts
}

func GetFiltersForm(farm uint32) entryFilters {
	crops, _ := crop_service.GetCropsByFarm(farm)
	fields, _ := field_service.GetFieldsByFarm(farm)
	vehicles, _ := vehicle_service.GetVehiclesByFarm(farm)

	return entryFilters{
		Vehicles: vehicles,
		Fields:   fields,
		Crops:    crops,
	}
}

func GetEntryContent(farm uint32, page int) entryContent {
	if page == 0 {
		page = 1
	}
	pageSize := 10
	entries, totalEntries := GetAllEntryDisplay(farm, page)
	drafts := GetAllEntryDraftsDisplay(farm)

	totalPages := 0
	if totalEntries > 0 {
		totalPages = (totalEntries + pageSize - 1) / pageSize
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

	return entryContent{
		Entries:      entries,
		Drafts:       drafts,
		NoContent:    len(entries) == 0 && len(drafts) == 0,
		Filters:      GetFiltersForm(farm),
		AppliedChips: filters.FilterChips{Items: []filters.ChipEntry{}},
		CurrentPage:  page,
		TotalPages:   totalPages,
		NextPage:     nextPage,
		PrevPage:     prevPage,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}
}

type EntryForm struct {
	view.BaseTemplateData
	Vehicles        []entity_public.Vehicle
	SelectedVehicle uint16
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	Fields          []entity_public.Field
	SelectedField   uint16
	Products        []entity_public.Product
	Entry           entity_public.EntryDTO
	People          []entity_public.PersonOption
	SelectedOrigin  *uint32
	Farm            uint32
}

func GetEntryForm(farm uint32) (EntryForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehiclesByFarm(farm)
	crops, cToast := crop_service.GetCropsByFarm(farm)
	fields, fToast := field_service.GetFieldsByFarm(farm)
	products, pToast := product_service.GetProducts()
	people, pepToast := person_service.GetPeopleByFarm(farm)

	return EntryForm{
		Vehicles: vehicles,
		Crops:    crops,
		Fields:   fields,
		Products: products,
		People:   people,
		Farm:     farm,
	}, []*entity_public.Toast{vToast, cToast, fToast, pToast, pepToast}
}

func GetExistingEntryForm(entryId uint32, farm uint32) (EntryForm, []*entity_public.Toast) {
	formFields, toasts := GetEntryForm(farm)
	entry, toast := entry_service.GetEntry(entryId)
	formFields.SelectedCrop = entry.Crop
	formFields.SelectedVehicle = entry.Vehicle
	formFields.SelectedField = entry.Field
	if entry.Origin != nil {
		for i, p := range formFields.People {
			if p.Id != nil && *p.Id == *entry.Origin {
				formFields.SelectedOrigin = formFields.People[i].Id
				break
			}
		}
	}

	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Entry = entry.ToDTO()
	return formFields, toasts
}

func GetEntryFormFromDraft(draftId uint32, farm uint32) (EntryForm, []*entity_public.Toast) {
	formFields, toasts := GetEntryForm(farm)
	draft, toast := entry_service.GetEntryDraft(draftId)

	if toast != nil {
		toasts = append(toasts, toast)
	}

	if draft.Id != 0 {
		formFields.SelectedCrop = draft.Crop
		formFields.SelectedVehicle = draft.Vehicle
		formFields.SelectedField = draft.Field

		if draft.Tare != nil {
			tare, _ := draft.Tare.Float64()
			formFields.Entry.Tare = tare
		}

		if draft.Origin != nil {
			for i, p := range formFields.People {
				if p.Id != nil && *p.Id == *draft.Origin {
					formFields.SelectedOrigin = formFields.People[i].Id
					break
				}
			}
		}
	}

	return formFields, toasts
}

type EntryDraftForm struct {
	view.BaseTemplateData
	Vehicles        []entity_public.Vehicle
	SelectedVehicle uint16
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	Fields          []entity_public.Field
	SelectedField   uint16
	Draft           entity_public.EntryDraft
	People          []entity_public.PersonOption
	SelectedOrigin  *uint32
}

func GetEntryDraftForm(farm uint32) (EntryDraftForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehiclesByFarm(farm)
	crops, cToast := crop_service.GetCropsByFarm(farm)
	fields, fToast := field_service.GetFieldsByFarm(farm)
	people, pToast := person_service.GetPeopleByFarm(farm)

	return EntryDraftForm{
		Vehicles: vehicles,
		Crops:    crops,
		Fields:   fields,
		People:   people,
	}, []*entity_public.Toast{vToast, cToast, fToast, pToast}
}

func GetEntryDraftTable(farm uint32) []entity_public.DisplayEntryDraft {
	return GetAllEntryDraftsDisplay(farm)
}

// EntryListItemView wraps DisplayEntry for template rendering with CSP nonce
type EntryListItemView struct {
	view.BaseTemplateData
	entity_public.DisplayEntry
}

// EntryDraftListItemView wraps DisplayEntryDraft for template rendering with CSP nonce
type EntryDraftListItemView struct {
	view.BaseTemplateData
	entity_public.DisplayEntryDraft
}
