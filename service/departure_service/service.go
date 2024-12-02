package departure_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
)

func GetDeparture(manifest uint32) (entity_public.Departure, bool) {
	departure := departure_model.GetDeparture(manifest)
	if departure == nil {
		return entity_public.Departure{}, true
	}
	return *departure, false
}

func AddDeparture(bd entity_public.Departure) entity_public.Departure {
	return departure_model.AddDeparture(bd)
}

func PutDeparture(d entity_public.Departure) (entity_public.Departure, bool) {
	return departure_model.PutDeparture(d)
}

func DeleteDeparture(manifest uint32) int {
	return departure_model.DeleteDeparture(manifest)
}
