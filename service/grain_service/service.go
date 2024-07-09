package grain_service

import (
	"armazenda/entity"
	"armazenda/model/grain_model"
	"time"
)

type SimplifiedGrainEntry struct {
	Product     string
	Field       string
	Vehicle     string
	ArrivalDate string
}

func MakeSimplifieidGrainEntry(ge entity.GrainEntry) SimplifiedGrainEntry {
	return SimplifiedGrainEntry{
		Product:     entity.GrainMap[ge.Product],
		Field:       ge.Field,
		Vehicle:     ge.Vehicle,
		ArrivalDate: time.UnixMilli(ge.ArrivalDate).Format("02/Jan/2006 - 03:04"),
	}
}

func GetAllGrainEntrySimplified() []SimplifiedGrainEntry {
	entries := grain_model.GetAllEntries()
	var simplifiedEntries []SimplifiedGrainEntry
	for _, entry := range entries {
		simplifiedEntries = append(simplifiedEntries, MakeSimplifieidGrainEntry(entry))
	}
	return simplifiedEntries
}

func AddGrainEntry(ge entity.GrainEntry) entity.GrainEntry {
	return grain_model.AddGrainEntry(ge)
}
