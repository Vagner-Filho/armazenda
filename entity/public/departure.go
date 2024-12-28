package entity_public

import "reflect"

type Departure struct {
	Manifest      uint32  `form:"manifest"`
	DepartureDate string  `form:"departureDate" binding:"required"`
	Product       Grain   `form:"product"  binding:"gte=0"`
	VehiclePlate  string  `form:"vehiclePlate" binding:"required"`
	Weight        float64 `form:"weight" binding:"required"`
	Buyer         string  `form:"buyer" binding:"required"`
}

type DepartureFilter struct {
	DepartureDateMin string  `form:"departureDateMin"`
	DepartureDateMax string  `form:"departureDateMax"`
	Product          Grain   `form:"product"`
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
