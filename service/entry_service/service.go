package entry_service

import (
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"time"
)

type SimplifiedEntry struct {
	Waybill      uint32
	Product      string
	Field        string
	VehiclePlate string
	NetWeight    float64
	ArrivalDate  string
}

func MakeSimplifiedEntry(ge entry_model.Entry) SimplifiedEntry {
    field := entry_model.GetField(ge.Field)
    vehicle := vehicle_model.GetVehicle(ge.Vehicle) 
	return SimplifiedEntry{
		Waybill:      ge.Waybill,
		Product:      entry_model.GrainMap[ge.Product],
		Field:        field.Name,
		VehiclePlate: vehicle.Plate,
		NetWeight:    ge.NetWeight,
		ArrivalDate:  time.UnixMilli(ge.ArrivalDate).Format("02/Jan/2006 - 03:04"),
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

func AddEntry(ge entry_model.Entry) entry_model.Entry {
	return entry_model.AddEntry(ge)
}

func DeleteEntry(id uint32) int {
	return entry_model.DeleteEntry(id)
}

func GetEntry(id uint32) entry_model.Entry {
	return entry_model.GetEntry(id)
}

func PutEntry(ge entry_model.Entry) *entry_model.Entry {
	return entry_model.PutEntry(ge)
}

func GetFields() []entry_model.Field {
    return entry_model.GetFields()
}

func AddField(name string) uint32 {
    return entry_model.AddField(name)
}
