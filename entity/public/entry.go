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
	NetWeight   decimal.Decimal
	ArrivalDate time.Time `time_format:"2006-01-02T15:04"`
	Farm        uint32
	Origin      string
}

type Entry struct {
	Id      uint32 `form:"id" json:"id"`
	Field   uint16 `form:"field" binding:"required" json:"field"`
	Crop    uint8  `form:"crop" binding:"required" json:"crop"`
	Vehicle uint16 `form:"vehiclePlate" json:"vehiclePlate"`
	/*GrossWeight float64   `form:"grossWeight" binding:"required"`
	Tare        float64   `form:"tare" binding:"required"`
	NetWeight   float64   `form:"netWeight" binding:"gte=0"`*/
	CargoWeight
	Analysis
	ArrivalDate time.Time `form:"arrivalDate" binding:"required" time_format:"2006-01-02T15:04" json:"arrivalDate"`
	Farm        uint32    `form:"farm" binding:"gte=0" json:"farm"`
	Origin      *uint32   `form:"origin,omitempty" json:"origin,omitempty"`
	ModifiedAt  time.Time `json:"modifiedAt"`
}

func (e Entry) ToDTO() EntryDTO {
	cargo := e.CargoWeight.ToDTO()
	analysis := e.Analysis.ToDTO()
	return EntryDTO{
		Id:             e.Id,
		ArrivalDate:    e.ArrivalDate,
		Vehicle:        e.Vehicle,
		Crop:           e.Crop,
		CargoWeightDTO: cargo,
		Farm:           e.Farm,
		Field:          e.Field,
		AnalysisDTO:    analysis,
	}
}

type EntryDTO struct {
	Id      uint32 `form:"id" json:"id"`
	Field   uint16 `form:"field" binding:"required" json:"field"`
	Crop    uint8  `form:"crop" binding:"required" json:"crop"`
	Vehicle uint16 `form:"vehiclePlate" json:"vehiclePlate"`
	/*GrossWeight float64   `form:"grossWeight" binding:"required"`
	Tare        float64   `form:"tare" binding:"required"`
	NetWeight   float64   `form:"netWeight" binding:"gte=0"`*/
	CargoWeightDTO
	AnalysisDTO
	ArrivalDate time.Time `form:"arrivalDate" binding:"required" time_format:"2006-01-02T15:04" json:"arrivalDate"`
	Farm        uint32    `form:"farm" binding:"gte=0" json:"farm"`
}

func (dto EntryDTO) ToEntity() Entry {
	cargo := dto.CargoWeightDTO.ToEntity()
	analysis, _ := dto.AnalysisDTO.ToEntity()
	return Entry{
		Id:          dto.Id,
		ArrivalDate: dto.ArrivalDate,
		Vehicle:     dto.Vehicle,
		Crop:        dto.Crop,
		CargoWeight: cargo,
		Farm:        dto.Farm,
		Field:       dto.Field,
		Analysis:    analysis,
	}
}

type EntryFilter struct {
	Id             uint32    `form:"id" json:"id"`
	Product        uint8     `form:"product" json:"product"`
	Field          uint16    `form:"field" json:"field"`
	Crop           uint8     `form:"crop" binding:"gte=0" json:"crop"`
	Vehicle        string    `form:"vehiclePlate" json:"vehiclePlate"`
	GrossWeightMin float64   `form:"grossWeightMin" json:"grossWeightMin"`
	GrossWeightMax float64   `form:"grossWeightMax" json:"grossWeightMax"`
	TareMin        float64   `form:"tareMin" json:"tareMin"`
	TareMax        float64   `form:"tareMax" json:"tareMax"`
	NetWeightMin   float64   `form:"netWeightMin" json:"netWeightMin"`
	NetWeightMax   float64   `form:"netWeightMax" json:"netWeightMax"`
	HumidityMin    string    `form:"humidityMin" json:"humidityMin"`
	HumidityMax    string    `form:"humidityMax" json:"humidityMax"`
	ArrivalDateMin time.Time `form:"arrivalDateMin" time_format:"2006-01-02T15:04" json:"arrivalDateMin"`
	ArrivalDateMax time.Time `form:"arrivalDateMax" time_format:"2006-01-02T15:04" json:"arrivalDateMax"`
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
	Id           uint32
	Safra        string
	VehiclePlate string
	CargoWeight
	AnalysisDTO
	ArrivalDate        time.Time `time_format:"2006-01-02T15:04"`
	InscricaoEstadual  string
	Produto            string
	FarmName           *string
	FarmStreet         *string
	FarmCep            *string
	FarmNumber         *int
	FarmNeighborhood   *string
	FarmCity           *string
	FarmState          *string
	Origin             *string
	Document           string
	StorageName        *string
	DiscountedHumidity decimal.Decimal `db:"-"`
	DiscountedDamage   decimal.Decimal `db:"-"`
	DiscountedImpurity decimal.Decimal `db:"-"`
	StorageTax         decimal.Decimal
	StorageTaxModifier decimal.Decimal
}
