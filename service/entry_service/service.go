package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/utils"
)

type SimplifiedEntry struct {
	Manifest     uint32
	Product      string
	Field        string
	VehiclePlate string
	NetWeight    float64
	ArrivalDate  string
}

func MakeSimplifiedEntry(ge entity_public.Entry) SimplifiedEntry {
	field := entry_model.GetField(ge.Field)
	vehicle := vehicle_model.GetVehicle(ge.Vehicle)
	return SimplifiedEntry{
		Manifest:     ge.Manifest,
		Product:      entry_model.GrainMap[ge.Product],
		Field:        field.Name,
		VehiclePlate: vehicle.Plate,
		NetWeight:    ge.NetWeight,
		ArrivalDate:  utils.GetReadableDate(ge.ArrivalDate),
	}
}

func GetAllEntrySimplified() []SimplifiedEntry {
	entries := entry_model.GetAllEntries()
	var simplifiedEntries []SimplifiedEntry
	for _, entry := range entries {
		simplifiedEntries = append(simplifiedEntries, MakeSimplifiedEntry(entry))
	}
	return simplifiedEntries
}

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
