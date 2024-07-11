package grain_service

import (
	"armazenda/entity"
	"armazenda/model/grain_model"
	"time"
)

type SimplifiedGrainEntry struct {
	Waybill      uint32
	Product      string
	Field        string
	VehiclePlate string
	Tare         float64
	ArrivalDate  string
}

func MakeSimplifiedGrainEntry(ge entity.GrainEntry) SimplifiedGrainEntry {
	return SimplifiedGrainEntry{
		Waybill:      ge.Waybill,
		Product:      entity.GrainMap[ge.Product],
		Field:        ge.Field,
		VehiclePlate: ge.VehiclePlate,
		Tare:         ge.Tare,
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

func AddGrainEntry(ge entity.GrainEntry) entity.GrainEntry {
	return grain_model.AddGrainEntry(ge)
}
