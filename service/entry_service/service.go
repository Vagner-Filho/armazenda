package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
)

func AddEntry(ge entity_public.Entry) entity_public.Entry {
	return entry_model.AddEntry(ge)
}

func DeleteEntry(id uint32) int {
	return entry_model.DeleteEntry(id)
}

func GetEntry(id uint32) entity_public.Entry {
	return entry_model.GetEntry(id)
}

func PutEntry(ge entity_public.Entry) *entity_public.Entry {
	return entry_model.PutEntry(ge)
}

func GetFields() []entry_model.Field {
	return entry_model.GetFields()
}

func AddField(name string) uint32 {
	return entry_model.AddField(name)
}
