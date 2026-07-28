package entity_public

import (
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

type LegalPerson struct {
	Id          uint32  `form:"id" json:"id"`
	CompanyName string  `form:"companyName" binding:"required" json:"companyName"`
	FantasyName *string `form:"fantasyName,omitempty" json:"fantasyName,omitempty"`
	Cnpj        string  `form:"cnpj" binding:"required" json:"cnpj"`
	Address
	Person Person
}

type NaturalPerson struct {
	Id   uint32 `form:"id" json:"id"`
	Name string `form:"name" binding:"required" json:"name"`
	Cpf  string `form:"cpf" binding:"required" json:"cpf"`
	Address
	Person Person
}

type PersonConfig struct {
	HumidityProgressionId *uint32         `form:"humidityProgressionId" json:"humidityProgressionId"`
	EntrySoyDiscount      decimal.Decimal `form:"entrySoyDiscount" json:"entrySoyDiscount"`
	EntryCornDiscount     decimal.Decimal `form:"entryCornDiscount" json:"entryCornDiscount"`
}

type PersonCND struct {
	CertificateNumber *string                 `form:"certificateNumber"`
	ExpDate           *time.Time              `form:"expDate" time_format:"2006-01-02"`
	Meta              *map[string]interface{} `form:"meta"`
}

func (pc PersonConfig) GetProductEntryDiscount(product uint8) decimal.Decimal {
	if product == 1 {
		return pc.EntryCornDiscount
	}
	if product == 2 {
		return pc.EntrySoyDiscount
	}

	return decimal.Zero
}

type Person struct {
	Ie         string
	Id         uint32    `json:"id"`
	Farm       uint32    `form:"farm" binding:"gte=0" json:"farm"`
	ModifiedAt time.Time `json:"modifiedAt" db:"-"`
	PersonConfig
	PersonCND
}

type PersonOption struct {
	Id   *uint32
	Name string
	PersonConfig
}

type PersonDisplay struct {
	Type     uint8
	Name     string
	Document string
	IE       string
	Id       uint32
}

// FullPerson represents a complete person record (legal or natural) with all data.
type FullPerson struct {
	Type         uint8 // 1 = natural, 2 = legal
	Name         string
	Document     string // CPF or CNPJ
	IE           string
	Id           uint32
	Street       *string
	Cep          *string
	Number       *uint32
	Complement   *string
	Neighborhood *string
	City         *string
	State        *string
	Email        *string
	PhoneNumber  *string
	PersonCND
}

type PersonFilter struct {
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
	StartDate      time.Time `form:"startDate" time_format:"2006-01-02T15:04" json:"startDate"`
	EndDateMax     time.Time `form:"endDate" time_format:"2006-01-02T15:04" json:"endDate"`
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
