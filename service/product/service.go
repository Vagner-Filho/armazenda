package product_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/product_model"
)

func GetProducts() ([]entity_public.Product, *entity_public.Toast) {
	pModel, modelError := product_model.GetProductModel()
	if modelError != nil {
		toast := entity_public.GetErrorToast(modelError.Error(), "")
		return []entity_public.Product{}, &toast
	}

	products, err := pModel.GetProducts()
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.Product{}, &toast
	}

	return products, nil
}
