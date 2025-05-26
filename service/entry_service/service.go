package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
)

func AddEntry(ge entity_public.Entry) (entity_public.DisplayEntry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	if ge.NetWeight.IsZero() == true {
		ge.NetWeight = ge.CargoWeight.GrossWeight.Sub(ge.Tare)
	}

	newEntry, addErr := eModel.AddEntry(ge)
	if addErr != nil {
		if addErr.IsServerErr == true {
			return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.DisplayEntry{}, entity_public.GetWarningToast(addErr.Message, "")
	}
	return newEntry, entity_public.GetSuccessToast("Entrada adicionada", "")
}

func GetEntry(id uint32) (*entity_public.Entry, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, err := eModel.GetEntry(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}
	return &entry, nil
}

func GetEntryPdf(id uint32) (*entity_public.EntryPdf, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, err := eModel.GetEntryPdf(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}
	return &entry, nil
}

func PutEntry(ge entity_public.Entry) (entity_public.DisplayEntry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, putErr := eModel.PutEntry(ge)
	if putErr != nil {
		if putErr.IsServerErr == true {
			return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Houve um erro interno ao editar a entrada", "")
		}
		return entity_public.DisplayEntry{}, entity_public.GetWarningToast(putErr.Message, "")
	}
	return entry, entity_public.GetSuccessToast("Entrada editada", "")
}

func DeleteEntry(id uint32) *entity_public.Toast {
	dModel := entry_model.GetEntryModel()
	err := dModel.DeleteEntry(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Entrada deletada", "")
	return &toast
}

func FilterEntries(ef entity_public.EntryFilter) ([]entity_public.DisplayEntry, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()
	entries, err := eModel.FilterEntries(ef)

	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar entradas", "")
		return entries, &toast
	}

	return entries, nil
}
