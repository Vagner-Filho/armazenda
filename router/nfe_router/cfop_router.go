package nfe_router

import (
	"fmt"
	"net/http"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/model/cfop_model"
	model_error "armazenda/model/error"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"
	cfop_view "armazenda/view/cfop"

	"github.com/gin-gonic/gin"
)

// cfopDescriptionMaxLength mirrors the catalog practice of accepting long
// official descriptions while keeping the modal input tight.
const cfopDescriptionMaxLength = 200

type cfopRegistrar interface {
	CatalogHasCfop(code string) (bool, *model_error.ModelError)
	AddFarmCfop(farmID uint32, code, description, originDestination string) (entity_public.CfopOption, *model_error.ModelError)
}

type cfopRegistrarProvider func() (cfopRegistrar, uint32)

// getCfopForm renders the "Cadastrar CFOP" modal opened from any item row's
// "+" button. It needs no data beyond the request nonce.
func getCfopForm(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-cfop-form", gin.H{
		"CSPNonce": nonce.(string),
	})
}

// getCfopOptions renders the farm-ordered optgroup fragment used to refresh a
// row's select after a farm CFOP is registered. The optional "selected" query
// parameter keeps the newest option selected.
func getCfopOptions(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	options, toast := cfop_view.GetCfopOptions(farmID)
	if toast != nil {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "cfop-options", gin.H{
		"Options":  options,
		"Selected": strings.TrimSpace(c.Query("selected")),
	})
}

// addCfop registers a farm-tied CFOP. The catalog is read-only: codes already
// in the system catalog are rejected (catalog ∪ farm stays disjoint), as are
// duplicates inside the same farm. All failures are pt-br warning/error
// toasts.
func addCfop(c *gin.Context) {
	addCfopWithProvider(c, func() (cfopRegistrar, uint32) {
		sid, _ := c.Cookie("session_id")
		farmID := user_service.GetFarmFromToken(sid)
		return cfop_model.GetCfopModel(), farmID
	})
}

// addCfopWithProvider validates form values before resolving the farm/model.
// The provider seam keeps request validation and server-derived classification
// testable without a database or authenticated integration fixture.
func addCfopWithProvider(c *gin.Context, provider cfopRegistrarProvider) {
	code := strings.TrimSpace(c.PostForm("code"))
	description := strings.TrimSpace(c.PostForm("description"))

	// Format validation runs before any DB access so it is unit-testable
	// without a database.
	if !nfe_service.ValidSelectorCFOP(code) {
		cfopToast(c, http.StatusBadRequest, entity_public.GetWarningToast(
			"CFOP inválido",
			"Informe um CFOP de 4 dígitos começando com 1, 2, 3, 5, 6 ou 7"))
		return
	}
	if description == "" {
		cfopToast(c, http.StatusBadRequest, entity_public.GetWarningToast(
			"Descrição obrigatória",
			"Informe uma descrição para o CFOP"))
		return
	}
	if len([]rune(description)) > cfopDescriptionMaxLength {
		cfopToast(c, http.StatusBadRequest, entity_public.GetWarningToast(
			"Descrição muito longa",
			fmt.Sprintf("A descrição deve ter no máximo %d caracteres", cfopDescriptionMaxLength)))
		return
	}
	// Origem/Destino is canonical data derived from the CFOP, never a
	// user-supplied form value.
	originDestination := nfe_service.OriginDestinationForCFOP(code)
	cfopModel, farmID := provider()

	catalogHasCode, catalogErr := cfopModel.CatalogHasCfop(code)
	if catalogErr != nil {
		cfopToast(c, http.StatusInternalServerError, entity_public.GetErrorToast("Falhamos ao verificar o CFOP", ""))
		return
	}
	if catalogHasCode {
		cfopToast(c, http.StatusBadRequest, entity_public.GetWarningToast(
			"CFOP já cadastrado",
			"Este CFOP já faz parte do catálogo do sistema"))
		return
	}

	option, addErr := cfopModel.AddFarmCfop(farmID, code, description, originDestination)
	if addErr != nil {
		if addErr.IsServerErr {
			cfopToast(c, http.StatusInternalServerError, entity_public.GetErrorToast(addErr.Message, ""))
			return
		}
		cfopToast(c, http.StatusBadRequest, entity_public.GetWarningToast(
			"CFOP já cadastrado",
			"Este CFOP já está cadastrado para esta fazenda"))
		return
	}

	toast := entity_public.GetSuccessToast("CFOP cadastrado", "")
	c.Header("HX-Trigger", toast.ToJsonStr())
	c.HTML(http.StatusCreated, "cfop-option", option)
}

// cfopToast writes a toast header and status without rendering a body.
func cfopToast(c *gin.Context, status int, toast entity_public.Toast) {
	c.Header("HX-Trigger", toast.ToJsonStr())
	c.Status(status)
}

// loadCfopOptions assembles the selector options for a farm; on failure it
// degrades to the system default and surfaces a toast header instead of
// blocking the render.
func loadCfopOptions(c *gin.Context, farmID uint32) cfop_view.CfopOptions {
	options, toast := cfop_view.GetCfopOptions(farmID)
	if toast != nil {
		c.Header("HX-Trigger", toast.ToJsonStr())
		return cfop_view.BuildCfopOptions(nil, cfop_view.DefaultCFOP)
	}
	return options
}

// UseCfopRoutes registers the CFOP selector routes inside the authenticated
// /nfe group (tier fiscal middleware applies).
func UseCfopRoutes(router gin.IRoutes) {
	router.GET("/cfop/form", getCfopForm)
	router.GET("/cfop/options", getCfopOptions)
	router.POST("/cfop", addCfop)
}
