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

func GetFilledLegalPersonForm(id uint32) (entity_public.LegalPerson, *entity_public.Toast) {
	person, t := person_service.GetLegalPerson(id)
	return person, t
}

func GetFilledNaturalPersonForm(id uint32) (entity_public.NaturalPerson, *entity_public.Toast) {
	person, t := person_service.GetNaturalPerson(id)
	return person, t
}
