package entry_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	crop_service "armazenda/service/crop"
	"armazenda/service/entry_service"
	field_service "armazenda/service/field"
	product_service "armazenda/service/product"
	"armazenda/service/vehicle_service"
)

type entryFilters struct {
	Fields   []entity_public.Field
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
}

type entryContent struct {
	Entries   []entity_public.SimplifiedEntry
	Filters   entryFilters
	NoContent bool
}

func GetAllEntrySimplified() []entity_public.SimplifiedEntry {
	eModel := entry_model.GetEntryModel()
	entries, getDataErr := eModel.GetAllEntriesSimplified()
	if getDataErr != nil {
		return []entity_public.SimplifiedEntry{}
	}
	return entries
}

func GetFiltersForm(farm uint32) entryFilters {
	crops, _ := crop_service.GetCropsByFarm(farm)
	fields, _ := field_service.GetFields()
	vehicles, _ := vehicle_service.GetVehicles()

	return entryFilters{
		Vehicles: vehicles,
		Fields:   fields,
		Crops:    crops,
	}
}

func GetEntryContent(farm uint32) entryContent {
	entries := GetAllEntrySimplified()
	return entryContent{
		Entries:   entries,
		NoContent: len(entries) == 0,
		Filters:   GetFiltersForm(farm),
	}
}

type EntryForm struct {
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
	Fields   []entity_public.Field
	Products []entity_public.Product
	Entry    entity_public.Entry
}

func GetEntryForm(farm uint32) (EntryForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehicles()
	crops, cToast := crop_service.GetCropsByFarm(farm)
	fields, fToast := field_service.GetFields()
	products, pToast := product_service.GetProducts()

	return EntryForm{
		Vehicles: vehicles,
		Crops:    crops,
		Fields:   fields,
		Products: products,
	}, []*entity_public.Toast{vToast, cToast, fToast, pToast}
}

func GetExistingEntryForm(entryId uint32, farm uint32) (EntryForm, []*entity_public.Toast) {
	formFields, toasts := GetEntryForm(farm)
	entry, toast := entry_service.GetEntry(entryId)

	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Entry = entry
	return formFields, toasts
}
