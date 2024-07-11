package grain_model

import (
	"armazenda/entity"
	"time"
)

var lastWaybill uint32 = 3
var grain_entry = []entity.GrainEntry{
	{
		Waybill:      1,
		Product:      entity.Corn,
		Field:        "talhão 1",
		Harvest:      "Safra Milho 2024",
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Waybill:      2,
		Product:      entity.Soy,
		Field:        "talhão 2",
		Harvest:      "Safra Soja 23/24",
		Vehicle:      "Mercedão 1113",
		VehiclePlate: "APB7755",
		GrossWeight:  15000,
		Tare:         5000,
		Humidity:     "10%",
		ArrivalDate:  time.Now().UnixMilli(),
	},
	{
		Waybill:      3,
		Product:      entity.Corn,
		Field:        "talhão 3",
		Harvest:      "Sofra Milho 2024/2",
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
    lastWaybill = lastWaybill + 1
    ge.Waybill = lastWaybill
	grain_entry = append(grain_entry, ge)
	return ge
}
