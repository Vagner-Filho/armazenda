package entry_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
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
		ArrivalDate:  ge.ArrivalDate,
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
