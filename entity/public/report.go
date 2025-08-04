package entity_public

import (
	"reflect"
	"time"
)

type ReportFilter struct {
	Product      uint8     `form:"product"`
	Vehicle      string    `form:"vehiclePlate"`
	NetWeightMin float64   `form:"netWeightMin"`
	NetWeightMax float64   `form:"netWeightMax"`
	StartDate    time.Time `form:"startDate" time_format:"2006-01-02T15:04"`
	EndDate      time.Time `form:"endDate" time_format:"2006-01-02T15:04"`
	PersonId     uint32    `form:"person"`
}

type reportFilterCollection map[string]func(rf ReportFilter) string

func (rf ReportFilter) GetFilters(availableFilters reportFilterCollection) reportFilterCollection {
	userFilters := make(reportFilterCollection)

	values := reflect.ValueOf(rf)

	for i := range values.NumField() {
		field := values.Type().Field(i)
		fieldName := field.Name
		fieldValue := values.Field(i)

		if !fieldValue.IsZero() {
			userFilters[fieldName] = availableFilters[fieldName]
		}
	}
	return userFilters
}

type ReportDisplay struct {
	Romaneio      uint32
	OperationType uint8
	Product       string
	Vehicle       string
	NetWeight     float64
	OperationDate time.Time `time_format:"2006-01-02T15:04"`
	Person        string
	// PersonId      *uint32
}

type FullReport struct {
	ReportDisplay
	GrossWeight float64
	Tare        float64
	City        string
	State       string
	Humidity    float64
	Damage      float64
	Impurity    float64
}
