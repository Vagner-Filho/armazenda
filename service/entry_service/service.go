package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
)

func AddEntry(ge entity_public.Entry) entity_public.Entry {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}

	newEntry, addErr := eModel.AddEntry(ge)
	if addErr != nil {
	}
	return newEntry
}

func DeleteEntry(id uint32) int {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}
	eModel.DeleteEntry(id)
	return 111
}

func GetEntry(id uint32) entity_public.Entry {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}
	eModel.GetEntry(id)
	return entity_public.Entry{}
}

func PutEntry(ge entity_public.Entry) *entity_public.Entry {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}
	eModel.PutEntry(ge)
	return &entity_public.Entry{}
}
