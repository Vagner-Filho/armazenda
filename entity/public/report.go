package entity_public

import (
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

type ReportFilter struct {
	Product      uint8     `form:"product"`
	Vehicle      string    `form:"vehiclePlate"`
	NetWeightMin float64   `form:"netWeightMin"`
	NetWeightMax float64   `form:"netWeightMax"`
	StartDate    time.Time `form:"startDate" time_format:"2006-01-02"`
	EndDate      time.Time `form:"endDate" time_format:"2006-01-02"`
	OriginId     string    `form:"origin"`
	RecipientId  string    `form:"recipient"`
	Type         uint8     `form:"type"`
	Field        uint32    `form:"field"`
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
	OriginName    string
	OriginId      *uint32
	RecipientName string
	RecipientId   *uint32
	FieldName     string
	FieldId       *uint32
}

type FullReport struct {
	ReportDisplay
	GrossWeight        decimal.Decimal
	Tare               decimal.Decimal
	City               string
	State              string
	Humidity           decimal.Decimal
	Damage             decimal.Decimal
	Impurity           decimal.Decimal
	HumidityDiscount   decimal.Decimal
	DiscountedHumidity decimal.Decimal `db:"-"`
	DiscountedDamage   decimal.Decimal `db:"-"`
	DiscountedImpurity decimal.Decimal `db:"-"`
	ServiceTax         *decimal.Decimal
	WeightTax          *decimal.Decimal
}
