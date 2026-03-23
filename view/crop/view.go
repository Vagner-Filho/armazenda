package crop_view

import (
	entity_public "armazenda/entity/public"
	product_service "armazenda/service/product"
	"armazenda/view"
)

type CropForm struct {
	view.BaseTemplateData
	Products []entity_public.Product
	Pattern  string
}

func GetCropForm() (CropForm, *entity_public.Toast) {
	products, toast := product_service.GetProducts()
	return CropForm{Products: products}, toast
}
