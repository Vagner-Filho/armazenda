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

type Entry struct {
	Manifest    uint32
	Product     entity_public.Grain
	Field       uint32
	Harvest     string
	Vehicle     string
	GrossWeight float64
	Tare        float64
	NetWeight   float64
	Humidity    string
	ArrivalDate int64
}

var lastManifest uint32 = 3

var vehicles = vehicle_model.GetVehicles()

var entries = []Entry{
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
		ArrivalDate: time.Now().UnixMilli(),
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
		ArrivalDate: time.Now().UnixMilli(),
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
		ArrivalDate: time.Now().UnixMilli(),
	},
}

func generateArrivalDate(offsetDays int) string {
	today := time.Now()
	arrivalDate := today.AddDate(0, 0, offsetDays)
	return arrivalDate.Local().Format("02/Jan/2006 - 03:04")
}

func InitGrainMap() {
	GrainMap[entity_public.Corn] = "Milho"
	GrainMap[entity_public.Soy] = "Soja"
}

func GetAllEntries() []Entry {
	return entries
}

func AddEntry(ge Entry) Entry {
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

func GetEntry(id uint32) Entry {
	var entry *Entry = nil
	for _, ge := range entries {
		if ge.Manifest == id {
			entry = &ge
		}
	}
	return *entry
}

func PutEntry(ge Entry) *Entry {
	var geIndex = slices.IndexFunc(entries, func(e Entry) bool {
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
