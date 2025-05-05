package entity_public

import (
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

type Departure struct {
	Id            uint32          `form:"id"`
	DepartureDate time.Time       `form:"departureDate" binding:"required" time_format:"2006-01-02T15:04"`
	VehiclePlate  string          `form:"vehiclePlate" binding:"required"`
	Crop          uint8           `form:"crop" binding:"required"`
	GrossWeight   decimal.Decimal `form:"grossWeight" binding:"required"`
	Tare          decimal.Decimal `form:"tare" binding:"required"`
	NetWeight     decimal.Decimal `form:"netWeight" binding:"gte=0"`
	Buyer         uint32          `form:"buyer" binding:"required"`
	Farm          uint32          `form:"farm" binding:"gte=0"`
}

func (d Departure) ToFormDeparture() FormDeparture {
	gw, _ := d.GrossWeight.Float64()
	tare, _ := d.Tare.Float64()
	nw, _ := d.NetWeight.Float64()
	return FormDeparture{
		Id:            d.Id,
		DepartureDate: d.DepartureDate,
		VehiclePlate:  d.VehiclePlate,
		Crop:          d.Crop,
		GrossWeight:   gw,
		Tare:          tare,
		NetWeight:     nw,
		Buyer:         d.Buyer,
		Farm:          d.Farm,
	}
}

type FormDeparture struct {
	Id            uint32
	DepartureDate time.Time
	VehiclePlate  string
	Crop          uint8
	GrossWeight   float64
	Tare          float64
	NetWeight     float64
	Buyer         uint32
	Farm          uint32
}

type DisplayDeparture struct {
	Id            uint32
	Product       string
	VehiclePlate  string
	NetWeight     float64
	DepartureDate time.Time
}

type DepartureFilter struct {
	DepartureDateMin time.Time `form:"departureDateMin" time_format:"2006-01-02T15:04"`
	DepartureDateMax time.Time `form:"departureDateMax" time_format:"2006-01-02T15:04"`
	Product          uint8     `form:"product"`
	VehiclePlate     string    `form:"vehiclePlate"`
	NetWeightMin     float64   `form:"netWeightMin"`
	NetWeightMax     float64   `form:"netWeightMax"`
	Buyer            string    `form:"buyer"`
}

type departureFilterCollection map[string]func(df DepartureFilter) string

func (df DepartureFilter) GetFilters(availableFilters departureFilterCollection) departureFilterCollection {
	userFilters := make(departureFilterCollection)

	values := reflect.ValueOf(df)

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
