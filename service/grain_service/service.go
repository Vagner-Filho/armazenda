package grain_service

import (
	"armazenda/model/grain_model"
	"time"
)

type SimplifiedGrainEntry struct {
	Waybill      uint32
	Product      string
	Field        string
	VehiclePlate string
	NetWeight    float64
	ArrivalDate  string
}

func MakeSimplifiedGrainEntry(ge grain_model.GrainEntry) SimplifiedGrainEntry {
	return SimplifiedGrainEntry{
		Waybill:      ge.Waybill,
		Product:      grain_model.GrainMap[ge.Product],
		Field:        ge.Field,
		VehiclePlate: ge.VehiclePlate,
		NetWeight:    ge.NetWeight,
		ArrivalDate:  time.UnixMilli(ge.ArrivalDate).Format("02/Jan/2006 - 03:04"),
	}
}

func GetAllGrainEntrySimplified() []SimplifiedGrainEntry {
	entries := grain_model.GetAllEntries()
	var simplifiedEntries []SimplifiedGrainEntry
	for _, entry := range entries {
		simplifiedEntries = append(simplifiedEntries, MakeSimplifiedGrainEntry(entry))
	}
	return simplifiedEntries
}

func AddGrainEntry(ge grain_model.GrainEntry) grain_model.GrainEntry {
	return grain_model.AddGrainEntry(ge)
}

func DeleteGrainEntry(id uint32) int {
	return grain_model.DeleteGrainEntry(id)
}

func GetEntry(id uint32) grain_model.GrainEntry {
	return grain_model.GetEntry(id)
}

func PutEntry(ge grain_model.GrainEntry) *grain_model.GrainEntry {
	return grain_model.PutEntry(ge)
}
