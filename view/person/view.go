package person_view

import (
	entity_public "armazenda/entity/public"
	person_service "armazenda/service/person"
)

type personPage struct {
	People []entity_public.PersonDisplay
}

func GetPersonPage(farm uint32) personPage {
	return personPage{}
}

type FilledPersonForm struct {
	Legal   *entity_public.LegalPerson
	Natural *entity_public.NaturalPerson
}

func GetFilledLegalPersonForm(id uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetLegalPerson(id)
	pform.Legal = &person
	return pform, t
}

func GetFilledNaturalPersonForm(id uint32) (FilledPersonForm, *entity_public.Toast) {
	var pform FilledPersonForm
	person, t := person_service.GetNaturalPerson(id)
	pform.Natural = &person
	return pform, t
}
