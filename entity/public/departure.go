package entity_public

import (
	"reflect"
	"time"
)

type Departure struct {
	Id            uint32    `form:"id"`
	DepartureDate time.Time `form:"departureDate" binding:"required" time_format:"2006-01-02T15:04"`
	VehiclePlate  string    `form:"vehiclePlate" binding:"required"`
	Crop          uint8     `form:"crop" binding:"required"`
	Weight        float64   `form:"weight" binding:"required"`
	Buyer         string    `form:"buyer" binding:"required"`
}

type DisplayDeparture struct {
	Id            uint32
	Product       string
	VehiclePlate  string
	Weight        float64
	DepartureDate time.Time
}

type DepartureFilter struct {
	DepartureDateMin string  `form:"departureDateMin"`
	DepartureDateMax string  `form:"departureDateMax"`
	Product          uint8   `form:"product"`
	VehiclePlate     string  `form:"vehiclePlate"`
	WeightMin        float64 `form:"weightMin"`
	WeightMax        float64 `form:"weightMax"`
	Buyer            string  `form:"buyer"`
}

type filterDepartureCollection map[string]func(d Departure, df DepartureFilter) bool

func (df DepartureFilter) GetFilters(availableFilters filterDepartureCollection) filterDepartureCollection {
	userFilters := make(filterDepartureCollection)

	values := reflect.ValueOf(df)

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
