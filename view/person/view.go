package person_view

import entity_public "armazenda/entity/public"

type personPage struct {
	People []entity_public.PersonDisplay
}

func GetPersonPage(farm uint32) personPage {
	return personPage{}
}
