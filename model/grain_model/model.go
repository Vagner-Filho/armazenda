package grain_model

import (
	"slices"
	"time"
)

type Grain int

const (
	Corn Grain = iota
	Soy  Grain = iota
)

var GrainMap = make(map[Grain]string)

type GrainEntry struct {
	Waybill      uint32
	Product      Grain
	Field        string
	Harvest      string
	Vehicle      string
	VehiclePlate string
	GrossWeight  float64
	Tare         float64
	NetWeight    float64
	Humidity     string
	ArrivalDate  int64
}

var lastWaybill uint32 = 3

var grain_entry = []GrainEntry{
	{
		Waybill:      1,
		Product:      Corn,
		Field:        "talhão 1",
		Harvest:      "Safra Milho 2024",
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		NetWeight:    15000 - 5000,
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Waybill:      2,
		Product:      Soy,
		Field:        "talhão 2",
		Harvest:      "Safra Soja 23/24",
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		NetWeight:    15000 - 5000,
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Waybill:      3,
		Product:      Corn,
		Field:        "talhão 3",
		Harvest:      "Sofra Milho 2024/2",
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		NetWeight:    15000 - 5000,
		ArrivalDate:  time.Now().UnixMilli(),
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

func GetAllEntries() []GrainEntry {
	return grain_entry
}

func AddGrainEntry(ge GrainEntry) GrainEntry {
	lastWaybill = lastWaybill + 1
	ge.Waybill = lastWaybill
    if ge.GrossWeight - ge.Tare != ge.NetWeight {
        ge.NetWeight = ge.GrossWeight - ge.Tare
    }
	grain_entry = append(grain_entry, ge)
	return ge
}

func DeleteGrainEntry(id uint32) int {
	var indexToRemove = -1
	for i, ge := range grain_entry {
		if ge.Waybill == id {
			indexToRemove = i
		}
	}
	if indexToRemove > -1 {
		grain_entry = slices.Delete(grain_entry, indexToRemove, indexToRemove+1)
	}
	return indexToRemove
}

func GetEntry(id uint32) GrainEntry {
	var entry *GrainEntry = nil
	for _, ge := range grain_entry {
		if ge.Waybill == id {
			entry = &ge
		}
	}
	return *entry
}

func PutEntry(ge GrainEntry) *GrainEntry {
	var geIndex = slices.IndexFunc(grain_entry, func(e GrainEntry) bool {
		return e.Waybill == ge.Waybill
	})
	if geIndex > -1 {
		grain_entry = slices.Replace(grain_entry, geIndex, geIndex+1, ge)
		return &ge
	}
	return nil
}
