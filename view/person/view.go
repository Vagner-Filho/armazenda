package person_view

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/humidity_progression_model"
	farm_config_service "armazenda/service/farm_config"
	person_service "armazenda/service/person"
	"armazenda/view"
)

type personPage struct {
	People []entity_public.PersonDisplay
}

func GetPersonPage(farm uint32) personPage {
	return personPage{}
}

type FilledPersonForm struct {
	view.BaseTemplateData
	Legal             *entity_public.LegalPerson
	Natural           *entity_public.NaturalPerson
	Progressions      []entity_public.HumidityProgression
	FarmProgressionId *uint32
}

func GetFilledLegalPersonForm(id uint32, farm uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetLegalPerson(id)
	fc, err := farm_config_service.GetFarmConfig(farm)
	if err != nil {
		model_error.GetLoggerModel().Log(err.Error())
	}
	if fc != nil {
		pform.FarmProgressionId = fc.HumidityProgressionId
	}
	pform.Legal = &person
	pform.Progressions = getProgressions(farm)
	return pform, t
}

func GetFilledNaturalPersonForm(id uint32, farm uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetNaturalPerson(id)
	fc, err := farm_config_service.GetFarmConfig(farm)
	if err != nil {
		model_error.GetLoggerModel().Log(err.Error())
	}
	if fc != nil {
		pform.FarmProgressionId = fc.HumidityProgressionId
	}
	pform.Natural = &person
	pform.Progressions = getProgressions(farm)
	return pform, t
}

func GetProgressionsForForm(farm uint32) []entity_public.HumidityProgression {
	return getProgressions(farm)
}

func getProgressions(farm uint32) []entity_public.HumidityProgression {
	hpm := humidity_progression_model.GetHumidityProgressionModel()
	progressions, err := hpm.ListProgressions(farm)
	if err != nil {
		return []entity_public.HumidityProgression{}
	}
	return progressions
}

// PersonListItemView wraps PersonDisplay for template rendering with CSP nonce
type PersonListItemView struct {
	view.BaseTemplateData
	entity_public.PersonDisplay
}
