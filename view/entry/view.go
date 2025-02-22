package entry_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/model/entry_model"
	"armazenda/model/field_model"
	crop_service "armazenda/service/crop"
	"armazenda/service/entry_service"
	field_service "armazenda/service/field"
	product_service "armazenda/service/product"
	"armazenda/service/vehicle_service"
	"fmt"
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

func GetFiltersForm() entryFilters {
	cropModel, _ := crop_model.GetCropModel()
	crops, cropsErr := cropModel.GetCrops()

	fieldModel, _ := field_model.GetFieldModel()
	fields, fieldsErr := fieldModel.GetFields()

	vehicles, _ := vehicle_service.GetVehicles()

	if cropsErr != nil {
		fmt.Printf("cropsErr: %v\n", cropsErr.Error())
	}

	if fieldsErr != nil {
		fmt.Printf("fieldsErr: %v\n", fieldsErr.Error())
	}

	return entryFilters{
		Vehicles: vehicles,
		Fields:   fields,
		Crops:    crops,
	}
}

func GetEntryContent() entryContent {
	entries := GetAllEntrySimplified()
	return entryContent{
		Entries:   entries,
		NoContent: len(entries) == 0,
		Filters:   GetFiltersForm(),
	}
}

type EntryForm struct {
	Vehicles []entity_public.Vehicle
	Crops    []entity_public.Crop
	Fields   []entity_public.Field
	Products []entity_public.Product
	Entry    entity_public.Entry
}

func GetEntryForm() (EntryForm, []*entity_public.Toast) {
	vehicles, vToast := vehicle_service.GetVehicles()
	crops, cToast := crop_service.GetCrops()
	fields, fToast := field_service.GetFields()
	products, pToast := product_service.GetProducts()

	return EntryForm{
		Vehicles: vehicles,
		Crops:    crops,
		Fields:   fields,
		Products: products,
	}, []*entity_public.Toast{vToast, cToast, fToast, pToast}
}

func GetExistingEntryForm(entryId uint32) (EntryForm, []*entity_public.Toast) {
	formFields, toasts := GetEntryForm()
	entry, toast := entry_service.GetEntry(entryId)

	if toast != nil {
		toasts = append(toasts, toast)
	}

	formFields.Entry = entry
	return formFields, toasts
}
