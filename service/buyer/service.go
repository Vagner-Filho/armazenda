package buyer_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
)

func GetBuyers() ([]entity_public.BuyerDisplay, *entity_public.Toast) {
	bmodel := buyer_model.GetBuyerModel()

	buyers, err := bmodel.GetBuyers()
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Erro ao buscar compradores", "")
			return []entity_public.BuyerDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.BuyerDisplay{}, &toast
	}
	return buyers, nil
}

func AddBuyerCompany(bc entity_public.BuyerCompany) (entity_public.BuyerDisplay, *entity_public.Toast) {
	bmodel := buyer_model.GetBuyerModel()

	buyer, err := bmodel.AddBuyerCompany(bc)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.BuyerDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.BuyerDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Comprador cadastrado!", "")
	return buyer, &toast
}

func AddBuyerPerson(bp entity_public.BuyerPerson) (entity_public.BuyerDisplay, *entity_public.Toast) {
	bmodel := buyer_model.GetBuyerModel()

	buyer, err := bmodel.AddBuyerPerson(bp)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.BuyerDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.BuyerDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Comprador cadastrado!", "")
	return buyer, &toast
}
