package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
)

func AddEntry(ge entity_public.Entry) (entity_public.Entry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

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

func GetEntry(id uint32) (entity_public.Entry, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, err := eModel.GetEntry(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return entity_public.Entry{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.Entry{}, &toast
	}
	return entry, nil
}

func PutEntry(ge entity_public.Entry) (entity_public.Entry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, putErr := eModel.PutEntry(ge)
	if putErr != nil {
		if putErr.IsServerErr == true {
			return entity_public.Entry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.Entry{}, entity_public.GetWarningToast(putErr.Message, "")
	}
	return entry, entity_public.GetSuccessToast("Entrada adicionada", "")
}

func DeleteEntry(id uint32) *entity_public.Toast {
	dModel := entry_model.GetEntryModel()
	err := dModel.DeleteEntry(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Entrada deletada", "")
	return &toast
}
