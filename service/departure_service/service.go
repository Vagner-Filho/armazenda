package departure_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
)

func GetDeparture(manifest uint32) (entity_public.Departure, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.GetDeparture(manifest)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a saída", "")
			return entity_public.Departure{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.Departure{}, &toast
	}
	return departure, nil
}

func GetDisplayDepartures() ([]entity_public.DisplayDeparture, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.GetDisplayDepartures()
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a saída", "")
			return []entity_public.DisplayDeparture{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return []entity_public.DisplayDeparture{}, &toast
	}
	return departure, nil
}

func AddDeparture(bd entity_public.Departure) (entity_public.DisplayDeparture, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.AddDeparture(bd)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao adicionar a saída", "")
			return entity_public.DisplayDeparture{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.DisplayDeparture{}, &toast
	}
	toast := entity_public.GetSuccessToast("Saída cadastrada", "")
	return departure, &toast
}

func PutDeparture(d entity_public.Departure) (entity_public.Departure, bool) {
	return departure_model.PutDeparture(d)
}

func DeleteDeparture(id uint32) *entity_public.Toast {
	dModel := departure_model.GetDepartureModel()
	err := dModel.DeleteDeparture(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Saída deletada", "")
	return &toast
}
