package entity_public

import (
	"reflect"
	"time"
)

type LegalPerson struct {
	Id                uint8   `form:"id"`
	CompanyName       string  `form:"companyName" binding:"required"`
	FantasyName       string  `form:"fantasyName"`
	Cnpj              string  `form:"cnpj" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
	Person            Person
}

type NaturalPerson struct {
	Id                uint8   `form:"id"`
	Name              string  `form:"name" binding:"required"`
	Cpf               string  `form:"cpf" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
	Person            Person
}

type Person struct {
	Ie   string
	Id   uint32
	Farm uint32 `form:"farm" binding:"gte=0"`
}

type PersonOption struct {
	Id   uint8
	Name string
}

type PersonDisplay struct {
	Type     uint8
	Name     string
	Document string
	IE       string
	Id       uint32
}

type PersonFilter struct {
	Id             uint32    `form:"id"`
	Product        uint8     `form:"product"`
	Field          uint16    `form:"field"`
	Crop           uint8     `form:"crop" binding:"gte=0"`
	Vehicle        string    `form:"vehiclePlate"`
	GrossWeightMin float64   `form:"grossWeightMin"`
	GrossWeightMax float64   `form:"grossWeightMax"`
	TareMin        float64   `form:"tareMin"`
	TareMax        float64   `form:"tareMax"`
	NetWeightMin   float64   `form:"netWeightMin"`
	NetWeightMax   float64   `form:"netWeightMax"`
	HumidityMin    string    `form:"humidityMin"`
	HumidityMax    string    `form:"humidityMax"`
	StartDate      time.Time `form:"startDate" time_format:"2006-01-02T15:04"`
	EndDateMax     time.Time `form:"endDate" time_format:"2006-01-02T15:04"`
}

type personFilterCollection map[string]func(ef PersonFilter) string

func (ef PersonFilter) GetFilters(availableFilters personFilterCollection) personFilterCollection {
	userFilters := make(personFilterCollection)

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
