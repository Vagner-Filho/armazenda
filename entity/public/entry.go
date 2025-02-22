package entity_public

import (
	"reflect"
	"time"
)

type SimplifiedEntry struct {
	Id          uint32
	Product     string
	Field       string
	Vehicle     string
	NetWeight   float64
	ArrivalDate time.Time
}

type Entry struct {
	Id          uint32    `form:"id"`
	Field       uint16    `form:"field" binding:"required"`
	Crop        uint8     `form:"crop" binding:"required"`
	Vehicle     string    `form:"vehiclePlate"`
	GrossWeight float64   `form:"grossWeight" binding:"required"`
	Tare        float64   `form:"tare" binding:"required"`
	NetWeight   float64   `form:"netWeight"`
	Humidity    string    `form:"humidity" binding:"required"`
	ArrivalDate time.Time `form:"arrivalDate" binding:"required" time_format:"2006-01-02T15:04"`
}

type EntryFilter struct {
	Id             uint32  `form:"id"`
	Product        uint8   `form:"product"`
	Field          uint16  `form:"field"`
	Crop           uint8   `form:"crop" binding:"gte=0"`
	Vehicle        string  `form:"vehiclePlate"`
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

type filterCollection map[string]func(e Entry, ef EntryFilter) bool

func (ef EntryFilter) GetFilters(availableFilters filterCollection) filterCollection {
	userFilters := make(filterCollection)

	values := reflect.ValueOf(ef)

	for i := 0; i < values.NumField(); i++ {
		field := values.Type().Field(i)
		fieldName := field.Name
		fieldValue := values.Field(i)

		if !fieldValue.IsZero() {
			userFilters[fieldName] = availableFilters[fieldName]
		}
	}
	return userFilters
}
