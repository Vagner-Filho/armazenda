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
	Entries   []entity_public.DisplayEntry
	Drafts    []entity_public.DisplayEntryDraft
	Filters   entryFilters
	NoContent bool
}

func GetAllEntryDisplay(farm uint32) []entity_public.DisplayEntry {
	eModel := entry_model.GetEntryModel()
	entries, getDataErr := eModel.GetDisplayEntriesByFarm(farm)
	if getDataErr != nil {
		return []entity_public.DisplayEntry{}
	}
	return entries
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

func GetEntryContent(farm uint32) entryContent {
	entries := GetAllEntryDisplay(farm)
	drafts := GetAllEntryDraftsDisplay(farm)
	return entryContent{
		Entries:   entries,
		Drafts:    drafts,
		NoContent: len(entries) == 0 && len(drafts) == 0,
		Filters:   GetFiltersForm(farm),
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
	SelectedPerson  *uint32
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
		formFields.SelectedPerson = entry.Origin
	}

	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Entry = entry.ToDTO()
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
}

func GetEntryDraftForm(farm uint32) (EntryDraftForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehiclesByFarm(farm)
	crops, cToast := crop_service.GetCropsByFarm(farm)
	fields, fToast := field_service.GetFieldsByFarm(farm)

	return EntryDraftForm{
		Vehicles: vehicles,
		Crops:    crops,
		Fields:   fields,
	}, []*entity_public.Toast{vToast, cToast, fToast}
}
