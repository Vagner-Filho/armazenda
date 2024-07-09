package grain_model

import (
	"armazenda/entity"
	"time"
)

var grain_entry = []entity.GrainEntry{
	{
		Product:      entity.Corn,
		Field:        "talhão 1",
		HarvestYear:  time.Now().UnixMilli(),
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Product:      entity.Soy,
		Field:        "talhão 2",
		HarvestYear:  time.Now().UnixMilli(),
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Product:      entity.Corn,
		Field:        "talhão 3",
		HarvestYear:  time.Now().UnixMilli(),
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		ArrivalDate:  time.Now().UnixMilli(),
	},
}

func generateArrivalDate(offsetDays int) string {
	today := time.Now()
	arrivalDate := today.AddDate(0, 0, offsetDays)
	return arrivalDate.Local().Format("02/Jan/2006 - 03:04")
}

func InitGrainMap() {
	entity.GrainMap[entity.Corn] = "Milho"
	entity.GrainMap[entity.Soy] = "Soja"
}

func GetAllEntries() []entity.GrainEntry {
	return grain_entry
}

func AddGrainEntry(ge entity.GrainEntry) entity.GrainEntry {
	grain_entry = append(grain_entry, ge)
	return ge
}
