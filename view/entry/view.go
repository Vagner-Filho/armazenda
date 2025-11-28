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
)

type entryFilters struct {
	Fields   []entity_public.Field
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
}

type entryContent struct {
	Entries     []entity_public.DisplayEntry
	Drafts      []entity_public.DisplayEntryDraft
	Filters     entryFilters
	NoContent   bool
	CurrentPage int
	TotalPages  int
	NextPage    int
	PrevPage    int
	HasNext     bool
	HasPrev     bool
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
		Entries:     entries,
		Drafts:      drafts,
		NoContent:   len(entries) == 0 && len(drafts) == 0,
		Filters:     GetFiltersForm(farm),
		CurrentPage: page,
		TotalPages:  totalPages,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}
}

type EntryForm struct {
	Vehicles        []entity_public.Vehicle
	SelectedVehicle string
	Crops           []entity_public.Crop
	SelectedCrop    uint8
	Fields          []entity_public.Field
	SelectedField   uint16
	Products        []entity_public.Product
	Entry           entity_public.EntryDTO
	People          []entity_public.PersonOption
	SelectedOrigin  *uint32
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

		tare, _ := draft.Tare.Float64()
		formFields.Entry.Tare = tare

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
	Vehicles        []entity_public.Vehicle
	SelectedVehicle string
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
