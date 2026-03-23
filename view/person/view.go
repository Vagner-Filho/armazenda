package person_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/humidity_progression_model"
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
	Legal                 *entity_public.LegalPerson
	Natural               *entity_public.NaturalPerson
	Progressions          []entity_public.HumidityProgression
	SelectedProgressionId *uint32
}

func GetFilledLegalPersonForm(id uint32, farm uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetLegalPerson(id)
	pform.Legal = &person
	pform.SelectedProgressionId = person.Person.HumidityProgressionId
	pform.Progressions = getProgressions(farm)
	return pform, t
}

func GetFilledNaturalPersonForm(id uint32, farm uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetNaturalPerson(id)
	pform.Natural = &person
	pform.SelectedProgressionId = person.Person.HumidityProgressionId
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
	Person entity_public.PersonDisplay
}
