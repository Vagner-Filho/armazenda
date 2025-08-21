package entity_public

import (
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

type DisplayEntry struct {
	Id          uint32
	Product     string
	Field       string
	Vehicle     string
	NetWeight   float64
	ArrivalDate time.Time `time_format:"2006-01-02T15:04"`
	Farm        uint32
}

type Entry struct {
	Id      uint32 `form:"id"`
	Field   uint16 `form:"field" binding:"required"`
	Crop    uint8  `form:"crop" binding:"required"`
	Vehicle string `form:"vehiclePlate"`
	/*GrossWeight float64   `form:"grossWeight" binding:"required"`
	Tare        float64   `form:"tare" binding:"required"`
	NetWeight   float64   `form:"netWeight" binding:"gte=0"`*/
	CargoWeight
	Humidity    *float32  `form:"humidity"`
	Damage      *float32  `form:"damage"`
	Impurity    *float32  `form:"impurity,omitempty"`
	ArrivalDate time.Time `form:"arrivalDate" binding:"required" time_format:"2006-01-02T15:04"`
	Farm        uint32    `form:"farm" binding:"gte=0"`
	Origin      *uint32   `form:"origin,omitempty"`
}

func (e Entry) ToDTO() EntryDTO {
	cargo := e.CargoWeight.ToDTO()
	return EntryDTO{
		Id:             e.Id,
		ArrivalDate:    e.ArrivalDate,
		Vehicle:        e.Vehicle,
		Crop:           e.Crop,
		CargoWeightDTO: cargo,
		Farm:           e.Farm,
		Field:          e.Field,
		Humidity:       e.Humidity,
		Damage:         e.Damage,
		Impurity:       e.Impurity,
	}
}

type EntryDTO struct {
	Id      uint32 `form:"id"`
	Field   uint16 `form:"field" binding:"required"`
	Crop    uint8  `form:"crop" binding:"required"`
	Vehicle string `form:"vehiclePlate"`
	/*GrossWeight float64   `form:"grossWeight" binding:"required"`
	Tare        float64   `form:"tare" binding:"required"`
	NetWeight   float64   `form:"netWeight" binding:"gte=0"`*/
	CargoWeightDTO
	Humidity    *float32  `form:"humidity" binding:"required"`
	Damage      *float32  `form:"damage" binding:"required"`
	Impurity    *float32  `form:"impurity" binding:"required"`
	ArrivalDate time.Time `form:"arrivalDate" binding:"required" time_format:"2006-01-02T15:04"`
	Farm        uint32    `form:"farm" binding:"gte=0"`
}

func (dto EntryDTO) ToEntity() Entry {
	cargo := dto.CargoWeightDTO.ToEntity()
	return Entry{
		Id:          dto.Id,
		ArrivalDate: dto.ArrivalDate,
		Vehicle:     dto.Vehicle,
		Crop:        dto.Crop,
		CargoWeight: cargo,
		Farm:        dto.Farm,
		Field:       dto.Field,
		Humidity:    dto.Humidity,
		Damage:      dto.Damage,
		Impurity:    dto.Impurity,
	}
}

type EntryFilter struct {
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
	ArrivalDateMin time.Time `form:"arrivalDateMin" time_format:"2006-01-02T15:04"`
	ArrivalDateMax time.Time `form:"arrivalDateMax" time_format:"2006-01-02T15:04"`
}

type filterCollection map[string]func(ef EntryFilter) string

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

type EntryPdf struct {
	Id                uint32
	Safra             string
	VehiclePlate      string
	CargoWeight
	Humidity          *float32
	Damage            *float32
	Impurity          *float32
	ArrivalDate       time.Time `time_format:"2006-01-02T15:04"`
	InscricaoEstadual string
	Produto           string
	FarmName          string
	FarmStreet        *string
	FarmCep           *string
	FarmNumber        *int
	FarmNeighborhood  *string
	FarmCity          *string
	FarmState         *string
	Origin            *string
}

type EntryDraft struct {
	Id      uint32          `form:"id"`
	Field   uint16          `form:"field" binding:"required"`
	Crop    uint8           `form:"crop" binding:"required"`
	Vehicle string          `form:"vehiclePlate"`
	Tare    decimal.Decimal `form:"tare" binding:"gte=0"`
	Farm    uint32          `form:"farm" binding:"gte=0"`
}

type DisplayEntryDraft struct {
	Id      uint32          `form:"id"`
	Field   string          `form:"field" binding:"required"`
	Crop    string          `form:"crop" binding:"required"`
	Vehicle string          `form:"vehiclePlate"`
	Tare    decimal.Decimal `form:"tare" binding:"gte=0"`
}
