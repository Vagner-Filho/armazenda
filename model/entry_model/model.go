package entry_model

import (
	"armazenda/model/vehicle_model"
	"slices"
	"time"
)

type Grain int

const (
	Corn Grain = iota
	Soy  Grain = iota
)

var GrainMap = make(map[Grain]string)

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
	Waybill     uint32
	Product     Grain
	Field       Field
	Harvest     string
	Vehicle     vehicle_model.Vehicle
	GrossWeight float64
	Tare        float64
	NetWeight   float64
	Humidity    string
	ArrivalDate int64
}

var lastWaybill uint32 = 3

var entries = []Entry{
	{
		Waybill:     1,
		Product:     Corn,
		Field:       fields[0],
		Harvest:     "Safra Milho 2024",
		Vehicle:     vehicle_model.Vehicle{Name: "Mercedão", Plate: "OPA 0192"},
		GrossWeight: 15000,
		Tare:        5000,
		Humidity:    "10%",
		NetWeight:   15000 - 5000,
		ArrivalDate: time.Now().UnixMilli(),
	},
	{
		Waybill:     2,
		Product:     Soy,
		Field:       fields[1],
		Harvest:     "Safra Soja 23/24",
		Vehicle:     vehicle_model.Vehicle{Name: "Scania", Plate: "EPA 0192"},
		GrossWeight: 15000,
		Tare:        5000,
		Humidity:    "10%",
		NetWeight:   15000 - 5000,
		ArrivalDate: time.Now().UnixMilli(),
	},
	{
		Waybill:     3,
		Product:     Corn,
		Field:       fields[0],
		Harvest:     "Sofra Milho 2024/2",
		Vehicle:     vehicle_model.Vehicle{Name: "FH400", Plate: "PPK 6969"},
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
	GrainMap[Corn] = "Milho"
	GrainMap[Soy] = "Soja"
}

func GetAllEntries() []Entry {
	return entries
}

func AddEntry(ge Entry) Entry {
	lastWaybill = lastWaybill + 1
	ge.Waybill = lastWaybill
	if ge.GrossWeight-ge.Tare != ge.NetWeight {
		ge.NetWeight = ge.GrossWeight - ge.Tare
	}
	entries = append(entries, ge)
	return ge
}

func DeleteEntry(id uint32) int {
	var indexToRemove = -1
	for i, ge := range entries {
		if ge.Waybill == id {
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
		if ge.Waybill == id {
			entry = &ge
		}
	}
	return *entry
}

func PutEntry(ge Entry) *Entry {
	var geIndex = slices.IndexFunc(entries, func(e Entry) bool {
		return e.Waybill == ge.Waybill
	})
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
    fields = append(fields, Field{ Name: name, Id: lastField.Id + 1 })
    return lastField.Id + 1
}
