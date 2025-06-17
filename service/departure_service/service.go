package departure_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
)

func GetDeparture(id uint32) (entity_public.Departure, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.GetDeparture(id)
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

func GetDisplayDepartures(farm uint32) ([]entity_public.DisplayDeparture, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.GetDisplayDepartures(farm)
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

func PutDeparture(d entity_public.Departure) (entity_public.DisplayDeparture, entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()
	departure, error := dModel.PutDeparture(d)
	if error != nil {
		if error.IsServerErr == true {
			return entity_public.DisplayDeparture{}, entity_public.GetErrorToast("Houve um erro interno ao editar a saída", "")
		}
		return entity_public.DisplayDeparture{}, entity_public.GetWarningToast(error.Message, "")
	}
	return departure, entity_public.GetSuccessToast("Saída editada", "")
}

func DeleteDeparture(id uint32) *entity_public.Toast {
	dModel := departure_model.GetDepartureModel()
	err := dModel.DeleteDeparture(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Saída deletada", "")
	return &toast
}

func GetDeparturePdf(id uint32) (*entity_public.DeparturePdf, *entity_public.Toast) {
	eModel := departure_model.GetDepartureModel()

	departure, err := eModel.GetDeparturePdf(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}
	return &departure, nil
}
