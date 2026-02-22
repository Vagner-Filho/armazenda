package entity_public

import (
	model_error "armazenda/model/error"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type Analysis struct {
	Humidity                 *decimal.Decimal `form:"humidity" json:"humidity,omitempty"`
	Damage                   *decimal.Decimal `form:"damage" json:"damage,omitempty"`
	Impurity                 *decimal.Decimal `form:"impurity,omitempty" json:"impurity,omitempty"`
	HumidityDiscountModifier *decimal.Decimal `form:"humidityDiscountModifier,omitempty" json:"humidityDiscountModifier,omitempty"`
}

func (a *Analysis) ToDTO() AnalysisDTO {
	var adto AnalysisDTO
	if a.Humidity != nil {
		h := a.Humidity.String()
		adto.Humidity = &h
	}
	if a.Damage != nil {
		d := a.Damage.String()
		adto.Damage = &d
	}
	if a.Impurity != nil {
		i := a.Impurity.String()
		adto.Impurity = &i
	}

	return adto
}

type AnalysisDTO struct {
	Humidity *string `form:"humidity" json:"humidity,omitempty"`
	Damage   *string `form:"damage" json:"damage,omitempty"`
	Impurity *string `form:"impurity,omitempty" json:"impurity,omitempty"`
}

func (a *AnalysisDTO) ToEntity() (Analysis, error) {
	lm := model_error.GetLoggerModel()

	var entityHum *decimal.Decimal
	var hasErr bool = false
	var errStr string = "Falha interna para os valores:"
	if a.Humidity != nil {
		decHum, herr := decimal.NewFromString(*a.Humidity)
		if herr != nil {
			hasErr = true
			errStr += fmt.Sprintf(" %v ", a.Humidity)
			lm.Log(herr.Error())
		}
		entityHum = &decHum
	}
	var entityDam *decimal.Decimal
	if a.Damage != nil {
		decDam, derr := decimal.NewFromString(*a.Damage)
		if derr != nil {
			hasErr = true
			errStr += fmt.Sprintf(" %v ", a.Damage)
			lm.Log(derr.Error())
		}
		entityDam = &decDam
	}
	var entityImp *decimal.Decimal
	if a.Impurity != nil {
		decImp, ierr := decimal.NewFromString(*a.Impurity)
		if ierr != nil {
			hasErr = true
			errStr += fmt.Sprintf(" %v ", a.Impurity)
			lm.Log(ierr.Error())
		}
		entityImp = &decImp
	}

	if hasErr == false {
		return Analysis{
			Humidity: entityHum,
			Damage:   entityDam,
			Impurity: entityImp,
		}, nil
	}
	return Analysis{
		Humidity: entityHum,
		Damage:   entityDam,
		Impurity: entityImp,
	}, errors.New(errStr)
}
