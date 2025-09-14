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
	CargoWeight
	Farm   uint32  `form:"farm" binding:"gte=0"`
	Person *uint32 `form:"origin"`
	Analysis
}

func (d Departure) ToDTO() DepartureDTO {
	cargo := d.CargoWeight.ToDTO()
	analysis := d.Analysis.ToDTO()
	return DepartureDTO{
		Id:             d.Id,
		DepartureDate:  d.DepartureDate,
		VehiclePlate:   d.VehiclePlate,
		Crop:           d.Crop,
		CargoWeightDTO: cargo,
		Person:         d.Person,
		Farm:           d.Farm,
		AnalysisDTO:    analysis,
	}
}

type DepartureDTO struct {
	Id            uint32
	DepartureDate time.Time
	VehiclePlate  string
	Crop          uint8
	CargoWeightDTO
	Person *uint32
	Farm   uint32
	AnalysisDTO
}

func (dto DepartureDTO) ToEntity() Departure {
	cargo := dto.CargoWeightDTO.ToEntity()
	analysis, _ := dto.AnalysisDTO.ToEntity()
	return Departure{
		Id:            dto.Id,
		DepartureDate: dto.DepartureDate,
		VehiclePlate:  dto.VehiclePlate,
		Crop:          dto.Crop,
		CargoWeight:   cargo,
		Person:        dto.Person,
		Farm:          dto.Farm,
		Analysis:      analysis,
	}
}

type DisplayDeparture struct {
	Id            uint32
	Product       string
	VehiclePlate  string
	DepartureDate time.Time
	NetWeight     float64
	Farm          uint32 `db:"-"`
}

type DepartureFilter struct {
	DepartureDateMin time.Time `form:"departureDateMin" time_format:"2006-01-02T15:04"`
	DepartureDateMax time.Time `form:"departureDateMax" time_format:"2006-01-02T15:04"`
	Product          uint8     `form:"product"`
	VehiclePlate     string    `form:"vehiclePlate"`
	NetWeightMin     float64   `form:"netWeightMin"`
	NetWeightMax     float64   `form:"netWeightMax"`
	Person           string    `form:"person"`
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

type DeparturePdf struct {
	Id           uint32
	Safra        string
	VehiclePlate string
	CargoWeight
	DepartureDate     time.Time `time_format:"2006-01-02T15:04"`
	InscricaoEstadual string
	Produto           string
	PersonName        string
	Document          string
	FarmName          string
	FarmStreet        *string
	FarmCep           *string
	FarmNumber        *int
	FarmNeighborhood  *string
	FarmCity          *string
	FarmState         *string
	AnalysisDTO
}
