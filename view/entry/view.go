package entry_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/model/entry_model"
	"armazenda/model/field_model"
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
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
		fmt.Printf("%v", getModelErr.Error())
	}
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
