package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
)

func AddEntry(ge entity_public.Entry) (entity_public.Entry, entity_public.Toast) {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}

	if ge.NetWeight == 0 {
		ge.NetWeight = ge.GrossWeight - ge.Tare
	}

	newEntry, addErr := eModel.AddEntry(ge)
	if addErr != nil {
		if addErr.IsServerErr == true {
			return entity_public.Entry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.Entry{}, entity_public.GetWarningToast(addErr.Message, "")
	}
	return newEntry, entity_public.GetSuccessToast("Entrada adicionada", "")
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

func PutEntry(ge entity_public.Entry) (entity_public.Entry, entity_public.Toast) {
	eModel, getModelErr := entry_model.GetEntryModel()
	if getModelErr != nil {
	}

	entry, putErr := eModel.PutEntry(ge)
	if putErr != nil {
		if putErr.IsServerErr == true {
			return entity_public.Entry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.Entry{}, entity_public.GetWarningToast(putErr.Message, "")
	}
	return entry, entity_public.GetSuccessToast("Entrada adicionada", "")
}
