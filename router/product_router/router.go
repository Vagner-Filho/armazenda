package product_router

import (
	entity_public "armazenda/entity/public"
	product_service "armazenda/service/product"
	"armazenda/service/user_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// toDisplayRows composes the products page list: global catalog products
// (read-only) first, then the farm-created ones (editable).
func toDisplayRows(products []entity_public.Product, farmProducts []entity_public.FarmProduct) []entity_public.ProductDisplay {
	rows := make([]entity_public.ProductDisplay, 0, len(products)+len(farmProducts))

	for _, p := range products {
		rows = append(rows, entity_public.ProductDisplay{
			Id:     p.Id,
			Name:   p.Name,
			NCM:    p.NCM,
			IsFarm: false,
		})
	}
	for _, fp := range farmProducts {
		rows = append(rows, entity_public.ProductDisplay{
			Id:     fp.Id,
			Name:   fp.Name,
			NCM:    fp.NCM,
			IsFarm: true,
		})
	}

	return rows
}

// getProductsPage renders the products management page: global catalog
// products (read-only) plus the farm's own products (editable).
func getProductsPage(c *gin.Context) {
	nonce, exists := c.Get("csp_nonce")
	if !exists {
		c.Status(http.StatusForbidden)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	products, pToast := product_service.GetProducts()
	farmProducts, fpToast := product_service.GetFarmProducts(farm)
	if pToast != nil {
		c.Header("HX-Trigger", string(pToast.ToJson()))
	}
	if fpToast != nil {
		c.Header("HX-Trigger", string(fpToast.ToJson()))
	}

	c.HTML(http.StatusOK, "produto.html", gin.H{
		"CSPNonce": nonce.(string),
		"TierKey":  user_service.GetTierKeyFromContext(c),
		"Products": toDisplayRows(products, farmProducts),
	})
}

// getProductForm serves the empty create dialog.
func getProductForm(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "product-form", gin.H{
		"CSPNonce": nonce.(string),
	})
}

// getProductEditForm serves the pre-filled edit dialog for a farm-created
// product. Global products cannot be edited.
func getProductEditForm(c *gin.Context) {
	id, parseErr := strconv.ParseUint(c.Param("id"), 10, 8)
	if parseErr != nil {
		toast := entity_public.GetWarningToast("Produto inválido", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	farmProduct, toast := product_service.GetFarmProduct(uint8(id), farm)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "product-form", gin.H{
		"CSPNonce": nonce.(string),
		"Product":  farmProduct,
	})
}

func addProduct(c *gin.Context) {
	var newProduct entity_public.Product
	bindErr := c.Bind(&newProduct)
	if bindErr != nil {
		toast := entity_public.GetWarningToast("Dados do produto inválidos", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	farmProduct, toast := product_service.AddFarmProduct(entity_public.FarmProduct{
		Product: newProduct,
		FarmId:  farm,
	})
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	successToast := entity_public.GetSuccessToast("Produto Cadastrado", "")
	c.Header("HX-Trigger", string(successToast.ToJson()))
	c.HTML(http.StatusCreated, "product-list-item", entity_public.ProductDisplay{
		Id:     farmProduct.Id,
		Name:   farmProduct.Name,
		NCM:    farmProduct.NCM,
		IsFarm: true,
	})
}

func updateProduct(c *gin.Context) {
	id, parseErr := strconv.ParseUint(c.Param("id"), 10, 8)
	if parseErr != nil {
		toast := entity_public.GetWarningToast("Produto inválido", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	var updatedProduct entity_public.Product
	bindErr := c.Bind(&updatedProduct)
	if bindErr != nil {
		toast := entity_public.GetWarningToast("Dados do produto inválidos", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	farmProduct, toast := product_service.UpdateFarmProduct(entity_public.FarmProduct{
		Product: entity_public.Product{
			Id:   uint8(id),
			Name: updatedProduct.Name,
			NCM:  updatedProduct.NCM,
		},
		FarmId: farm,
	})
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	successToast := entity_public.GetSuccessToast("Produto Atualizado", "")
	c.Header("HX-Trigger", string(successToast.ToJson()))
	c.HTML(http.StatusOK, "product-list-item", entity_public.ProductDisplay{
		Id:     farmProduct.Id,
		Name:   farmProduct.Name,
		NCM:    farmProduct.NCM,
		IsFarm: true,
	})
}

func UseProductRoutes(router *gin.Engine) {
	router.GET("/produto", getProductsPage)
	router.GET("/produto/form", getProductForm)
	router.GET("/produto/form/:id", getProductEditForm)
	router.POST("/produto", addProduct)
	router.PUT("/produto/:id", updateProduct)
}
