package entity_public

import (
	"reflect"
)

type Entry struct {
	Manifest    uint32  `form:"manifest"`
	Product     Grain   `form:"product" binding:"gte=0"`
	Field       uint32  `form:"field" binding:"required"`
	Harvest     string  `form:"harvest" binding:"required"`
	Vehicle     string  `form:"vehiclePlate"`
	GrossWeight float64 `form:"grossWeight" binding:"required"`
	Tare        float64 `form:"tare" binding:"required"`
	NetWeight   float64 `form:"netWeight"`
	Humidity    string  `form:"humidity" binding:"required"`
	ArrivalDate string  `form:"arrivalDate" binding:"required"`
}

type EntryFilter struct {
	Manifest       uint32  `form:"manifest"`
	Product        Grain   `form:"product"`
	Field          uint32  `form:"field"`
	Harvest        string  `form:"harvest"`
	VehiclePlate   string  `form:"vehiclePlate"`
	GrossWeightMin float64 `form:"grossWeightMin"`
	GrossWeightMax float64 `form:"grossWeightMax"`
	TareMin        float64 `form:"tareMin"`
	TareMax        float64 `form:"tareMax"`
	NetWeightMin   float64 `form:"netWeightMin"`
	NetWeightMax   float64 `form:"netWeightMax"`
	HumidityMin    string  `form:"humidityMin"`
	HumidityMax    string  `form:"humidityMax"`
	ArrivalDateMin string  `form:"arrivalDateMin"`
	ArrivalDateMax string  `form:"arrivalDateMax"`
}

func (ef EntryFilter) GetFilters(filters map[string]func(e Entry, ef EntryFilter) bool) map[string]func(e Entry, ef EntryFilter) bool {
	mp := make(map[string]func(e Entry, ef EntryFilter) bool)

	values := reflect.ValueOf(ef)

	for i := 0; i < values.NumField(); i++ {
		field := values.Type().Field(i)
		fieldName := field.Name
		fieldValue := values.Field(i)

		if !fieldValue.IsZero() {
			mp[fieldName] = filters[fieldName]
		}
	}
	return mp
}
