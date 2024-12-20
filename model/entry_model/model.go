package entry_model

import (
	"armazenda/entity/public"
	"armazenda/model/vehicle_model"
	"slices"
	"time"
)

var GrainMap = make(map[entity_public.Grain]string)

type Field struct {
	Id   uint32
	Name string
}

var fields = []Field{
	{
		Id:   1,
		Name: "Talhão 1",
	},
	{
		Id:   2,
		Name: "bom dia talhão 2",
	},
}

var lastManifest uint32 = 3

var vehicles = vehicle_model.GetVehicles()

var entries = []entity_public.Entry{
	{
		Manifest:    1,
		Product:     entity_public.Corn,
		Field:       fields[0].Id,
		Harvest:     "Safra Milho 2024",
		Vehicle:     vehicles[0].Plate,
		GrossWeight: 15000,
		Tare:        5000,
		Humidity:    "10%",
		NetWeight:   15000 - 5000,
		ArrivalDate: time.Now().AddDate(0, -1, -3).Format(time.RFC3339),
	},
	{
		Manifest:    2,
		Product:     entity_public.Soy,
		Field:       fields[0].Id,
		Harvest:     "Safra Soja 23/24",
		Vehicle:     vehicles[0].Plate,
		GrossWeight: 15000,
		Tare:        5000,
		Humidity:    "10%",
		NetWeight:   15000 - 5000,
		ArrivalDate: time.Now().Format(time.RFC3339),
	},
	{
		Manifest:    3,
		Product:     entity_public.Corn,
		Field:       fields[0].Id,
		Harvest:     "Sofra Milho 2024/2",
		Vehicle:     vehicles[0].Plate,
		GrossWeight: 15000,
		Tare:        5000,
		Humidity:    "10%",
		NetWeight:   15000 - 5000,
		ArrivalDate: time.Now().Format(time.RFC3339),
	},
}

func InitGrainMap() {
	GrainMap[entity_public.Corn] = "Milho"
	GrainMap[entity_public.Soy] = "Soja"
}

func GetAllEntries() []entity_public.Entry {
	return entries
}

func AddEntry(ge entity_public.Entry) entity_public.Entry {
	lastManifest = lastManifest + 1
	ge.Manifest = lastManifest
	if ge.GrossWeight-ge.Tare != ge.NetWeight {
		ge.NetWeight = ge.GrossWeight - ge.Tare
	}
	entries = append(entries, ge)
	return ge
}

func DeleteEntry(id uint32) int {
	var indexToRemove = -1
	for i, ge := range entries {
		if ge.Manifest == id {
			indexToRemove = i
		}
	}
	if indexToRemove > -1 {
		entries = slices.Delete(entries, indexToRemove, indexToRemove+1)
	}
	return indexToRemove
}

func GetEntry(id uint32) entity_public.Entry {
	var entry *entity_public.Entry = nil
	for _, ge := range entries {
		if ge.Manifest == id {
			entry = &ge
		}
	}
	return *entry
}

func PutEntry(ge entity_public.Entry) *entity_public.Entry {
	var geIndex = slices.IndexFunc(entries, func(e entity_public.Entry) bool {
		return e.Manifest == ge.Manifest
	})

	if ge.NetWeight != ge.GrossWeight-ge.Tare {
		ge.NetWeight = ge.GrossWeight - ge.Tare
	}

	if geIndex > -1 {
		entries = slices.Replace(entries, geIndex, geIndex+1, ge)
		return &ge
	}
	return nil
}

func GetFields() []Field {
	return fields
}

func AddField(name string) uint32 {
	lastField := fields[len(fields)-1]
	fields = append(fields, Field{Name: name, Id: lastField.Id + 1})
	return lastField.Id + 1
}

func GetField(id uint32) *Field {
	fieldIndex := slices.IndexFunc(fields, func(e Field) bool {
		return e.Id == id
	})

	return &fields[fieldIndex]
}

func FilterEntries(filter entity_public.EntryFilter) []entity_public.Entry {
	var filteredEntries []entity_public.Entry

	arrivalFrom, _ := time.Parse(time.RFC3339, filter.ArrivalDateFrom)
	arrivalTo, _ := time.Parse(time.RFC3339, filter.ArrivalDateTo)

	for _, entry := range entries {
		arrivalFiltered, _ := time.Parse(time.RFC3339, entry.ArrivalDate)
		if arrivalFrom.Before(arrivalFiltered) && arrivalTo.After(arrivalFiltered) {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	return filteredEntries
}
