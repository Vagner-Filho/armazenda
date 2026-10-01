package product_service

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/farm_product_model"
	"armazenda/model/product_model"
)

// GetProducts returns the global catalog products — the ones selectable in
// romaneio flows (crops).
func GetProducts() ([]entity_public.Product, *entity_public.Toast) {
	pModel := product_model.GetProductModel()

	products, err := pModel.GetProducts()
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.Product{}, &toast
	}

	return products, nil
}

// GetFarmProducts lists the farm-created products for the products page.
func GetFarmProducts(farmId uint32) ([]entity_public.FarmProduct, *entity_public.Toast) {
	fm := farm_product_model.GetFarmProductModel()

	farmProducts, err := fm.GetFarmProductsByFarm(farmId)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.FarmProduct{}, &toast
	}

	return farmProducts, nil
}

// GetFarmProduct resolves one farm-created product, scoped to the farm.
func GetFarmProduct(id uint8, farmId uint32) (entity_public.FarmProduct, *entity_public.Toast) {
	fm := farm_product_model.GetFarmProductModel()

	farmProduct, err := fm.GetFarmProductById(id, farmId)
	if err != nil {
		if me, ok := err.(*model_error.ModelError); ok && me.IsServerErr {
			toast := entity_public.GetErrorToast(me.Error(), "")
			return entity_public.FarmProduct{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.FarmProduct{}, &toast
	}

	return farmProduct, nil
}

// AddFarmProduct creates a product owned by the given farm.
func AddFarmProduct(fp entity_public.FarmProduct) (entity_public.FarmProduct, *entity_public.Toast) {
	fm := farm_product_model.GetFarmProductModel()

	var t entity_public.Toast
	if len(fp.Name) == 0 {
		t = entity_public.GetWarningToast("Informe o nome do produto", "")
		return entity_public.FarmProduct{}, &t
	}
	if len(fp.NCM) != 8 {
		t = entity_public.GetWarningToast("NCM inválido", "informe os 8 dígitos do NCM")
		return entity_public.FarmProduct{}, &t
	}
	if fp.FarmId == 0 {
		t = entity_public.GetErrorToast("Fazenda não identificada", "")
		return entity_public.FarmProduct{}, &t
	}

	farmProduct, err := fm.AddFarmProduct(fp)
	if err != nil {
		if err.IsServerErr {
			t = entity_public.GetErrorToast(err.Error(), "")
			return entity_public.FarmProduct{}, &t
		}
		t = entity_public.GetWarningToast(err.Error(), "")
		return entity_public.FarmProduct{}, &t
	}

	return farmProduct, nil
}

// UpdateFarmProduct renames/re-NCMs a farm-created product.
func UpdateFarmProduct(fp entity_public.FarmProduct) (entity_public.FarmProduct, *entity_public.Toast) {
	fm := farm_product_model.GetFarmProductModel()

	var t entity_public.Toast
	if len(fp.Name) == 0 {
		t = entity_public.GetWarningToast("Informe o nome do produto", "")
		return entity_public.FarmProduct{}, &t
	}
	if len(fp.NCM) != 8 {
		t = entity_public.GetWarningToast("NCM inválido", "informe os 8 dígitos do NCM")
		return entity_public.FarmProduct{}, &t
	}
	if fp.FarmId == 0 {
		t = entity_public.GetErrorToast("Fazenda não identificada", "")
		return entity_public.FarmProduct{}, &t
	}

	farmProduct, err := fm.UpdateFarmProduct(fp)
	if err != nil {
		if err.IsServerErr {
			t = entity_public.GetErrorToast(err.Error(), "")
			return entity_public.FarmProduct{}, &t
		}
		t = entity_public.GetWarningToast(err.Error(), "")
		return entity_public.FarmProduct{}, &t
	}

	return farmProduct, nil
}
