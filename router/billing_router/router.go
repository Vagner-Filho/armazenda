package billing_router

import (
	"armazenda/model/owner_subscription_model"
	"armazenda/service/billing_service"
	"armazenda/service/user_service"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func BillingRoutes(router *gin.Engine) {
	router.POST("/stripe/webhook", stripeWebhook)
	router.GET("/pricing", pricingPage)
	router.GET("/payment/success", paymentSuccessPage)
	router.GET("/payment/cancel", paymentCancelPage)
	router.GET("/payment/required", paymentRequiredPage)
	router.POST("/payment/checkout", paymentCheckout)
	router.GET("/api/user/role", userRoleAPI)
	router.GET("/api/subscription/status", subscriptionStatusAPI)
	router.POST("/subscription/cancel", subscriptionCancel)
}

func translateSubscriptionStatus(status string) string {
	switch status {
	case "active":
		return "ativa"
	case "pending":
		return "pendente"
	case "past_due":
		return "vencida"
	case "canceled":
		return "cancelada"
	case "unpaid":
		return "não paga"
	case "trialing":
		return "em teste"
	case "paused":
		return "pausada"
	case "incomplete":
		return "incompleta"
	case "incomplete_expired":
		return "expirada"
	case "":
		return "inativa"
	default:
		return status
	}
}

func stripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read request body")
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if sigHeader == "" {
		c.String(http.StatusBadRequest, "Missing Stripe-Signature header")
		return
	}

	handleErr := billing_service.HandleWebhook(payload, sigHeader)
	if handleErr != nil {
		fmt.Printf("webhook error: %v\n", handleErr)
		c.String(http.StatusBadRequest, handleErr.Error())
		return
	}

	c.Status(http.StatusOK)
}

func pricingPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")

	featuredPriceID := os.Getenv("STRIPE_FEATURED_PRICE_ID")

	tiers, err := billing_service.GetPricingTiers()
	if err != nil {
		fmt.Printf("failed to fetch pricing tiers from Stripe: %v\n", err)
		c.String(http.StatusInternalServerError, "Falha ao carregar preços. Tente novamente mais tarde.")
		return
	}

	c.HTML(http.StatusOK, "pricing.html", gin.H{
		"CSPNonce":        nonce.(string),
		"PublishableKey":  billing_service.GetPublishableKey(),
		"Tiers":           tiers,
		"FeaturedPriceID": featuredPriceID,
	})
}

func paymentSuccessPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "payment-success.html", gin.H{
		"CSPNonce": nonce.(string),
	})
}

func paymentCancelPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "payment-cancel.html", gin.H{
		"CSPNonce": nonce.(string),
	})
}

func paymentRequiredPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")

	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	if farmID == 0 {
		if farmIDParam := c.Query("farmId"); farmIDParam != "" {
			if id, parseErr := strconv.ParseUint(farmIDParam, 10, 32); parseErr == nil {
				farmID = uint32(id)
			}
		}
	}

	osm := owner_subscription_model.GetOwnerSubscriptionModel()
	sub, err := osm.GetSubscriptionByFarm(farmID)
	if err != nil {
		fmt.Printf("subscription lookup error: %v\n", err.Error())
	}
	if err != nil || sub == nil {
		c.HTML(http.StatusOK, "subscription-inactive.html", gin.H{
			"CSPNonce":  nonce.(string),
			"Status":    "inativa",
			"PortalURL": "",
		})
		return
	}

	portalURL := ""
	if sub.StripeCustomerId != nil && *sub.StripeCustomerId != "" {
		url, portalErr := billing_service.CreateCustomerPortalSession(*sub.StripeCustomerId)
		if portalErr != nil {
			fmt.Printf("failed to create customer portal session: %v\n", portalErr)
		} else {
			portalURL = url
		}
	}

	status := translateSubscriptionStatus(sub.SubscriptionStatus)

	// If no Stripe customer yet, show checkout mode with tier selection
	if sub.StripeCustomerId == nil || *sub.StripeCustomerId == "" {
		users, usersErr := user_service.GetUsersByFarm(farmID, true)
		if usersErr != nil {
			fmt.Printf("failed to get users for farm %d: %v\n", farmID, usersErr)
		}

		var customerEmail string
		for _, u := range users {
			if u.Cpf == sub.OwnerDocument {
				customerEmail = u.Email
				break
			}
		}
		if customerEmail == "" && len(users) > 0 {
			customerEmail = users[0].Email
		}

		farms, farmsErr := user_service.GetFarmsByOwnerDocument(sub.OwnerDocument, sub.OwnerDocumentType)
		if farmsErr != nil {
			fmt.Printf("failed to get farms for owner %s: %v\n", sub.OwnerDocument, farmsErr)
		}
		farmCount := len(farms)
		if farmCount == 0 {
			farmCount = 1
		}

		featuredPriceID := os.Getenv("STRIPE_FEATURED_PRICE_ID")
		tierList, tierErr := billing_service.GetPricingTiers()
		if tierErr != nil {
			fmt.Printf("failed to fetch pricing tiers: %v\n", tierErr)
		}

		tempToken := billing_service.CreateTempOwnerSession(sub.OwnerDocument)

		c.HTML(http.StatusOK, "subscription-inactive.html", gin.H{
			"CSPNonce":        nonce.(string),
			"Status":          status,
			"PortalURL":       portalURL,
			"Tiers":           tierList,
			"FeaturedPriceID": featuredPriceID,
			"FarmCount":       farmCount,
			"TempToken":       tempToken,
			"CustomerEmail":   customerEmail,
			"FarmID":          farmID,
		})
		return
	}

	c.HTML(http.StatusOK, "subscription-inactive.html", gin.H{
		"CSPNonce":  nonce.(string),
		"Status":    status,
		"PortalURL": portalURL,
	})
}

func userRoleAPI(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims := user_service.GetClaimsFromToken(sessionCookie.Value)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": claims.Role})
}

func subscriptionStatusAPI(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims := user_service.GetClaimsFromToken(sessionCookie.Value)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	osm := owner_subscription_model.GetOwnerSubscriptionModel()
	sub, err := osm.GetSubscriptionByFarm(claims.Farm)
	if err != nil || sub == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":              "inativa",
			"tierKey":             claims.TierKey,
			"periodEnd":           nil,
			"cancelAtPeriodEnd":   false,
			"quantity":            0,
			"hasStripeCustomerId": false,
		})
		return
	}

	cancelAtPeriodEnd := false
	if sub.StripeSubscriptionId != nil && *sub.StripeSubscriptionId != "" {
		stripeSub, stripeErr := billing_service.GetStripeSubscription(*sub.StripeSubscriptionId)
		if stripeErr == nil && stripeSub != nil {
			cancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd
		}
	}

	var periodEnd *time.Time
	if sub.SubscriptionCurrentPeriodEnd != nil {
		periodEnd = sub.SubscriptionCurrentPeriodEnd
	}

	quantity, _ := osm.CountFarmsByOwner(sub.OwnerDocument, sub.OwnerDocumentType)

	c.JSON(http.StatusOK, gin.H{
		"status":              translateSubscriptionStatus(sub.SubscriptionStatus),
		"rawStatus":           sub.SubscriptionStatus,
		"tierKey":             sub.TierKey,
		"periodEnd":           periodEnd,
		"cancelAtPeriodEnd":   cancelAtPeriodEnd,
		"quantity":            quantity,
		"hasStripeCustomerId": sub.StripeCustomerId != nil && *sub.StripeCustomerId != "",
	})
}

func subscriptionCancel(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims := user_service.GetClaimsFromToken(sessionCookie.Value)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can cancel subscriptions"})
		return
	}

	osm := owner_subscription_model.GetOwnerSubscriptionModel()
	sub, err := osm.GetSubscriptionByFarm(claims.Farm)
	if err != nil || sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if sub.StripeSubscriptionId == nil || *sub.StripeSubscriptionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active stripe subscription"})
		return
	}

	cancelErr := billing_service.CancelSubscriptionAtPeriodEnd(*sub.StripeSubscriptionId)
	if cancelErr != nil {
		fmt.Printf("failed to cancel subscription: %v\n", cancelErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription cancellation scheduled"})
}

func paymentCheckout(c *gin.Context) {
	var form struct {
		FarmID    uint32 `form:"farmId"`
		PriceID   string `form:"priceId"`
		TempToken string `form:"tempToken"`
	}
	if err := c.Bind(&form); err != nil {
		c.String(http.StatusBadRequest, "Dados inválidos")
		return
	}

	ownerDoc, valid := billing_service.ValidateTempOwnerSession(form.TempToken)
	if !valid {
		c.String(http.StatusForbidden, "Sessão expirada. Tente novamente.")
		return
	}

	osm := owner_subscription_model.GetOwnerSubscriptionModel()
	sub, err := osm.GetSubscriptionByFarm(form.FarmID)
	if err != nil || sub == nil {
		c.String(http.StatusNotFound, "Assinatura não encontrada")
		return
	}

	if sub.OwnerDocument != ownerDoc {
		c.String(http.StatusForbidden, "Documento do proprietário não corresponde")
		return
	}

	if sub.StripeCustomerId != nil && *sub.StripeCustomerId != "" {
		c.String(http.StatusConflict, "Já existe um cliente Stripe vinculado")
		return
	}

	users, usersErr := user_service.GetUsersByFarm(form.FarmID, true)
	if usersErr != nil {
		c.String(http.StatusInternalServerError, "Erro ao buscar usuários")
		return
	}

	var customerEmail string
	for _, u := range users {
		if u.Cpf == sub.OwnerDocument {
			customerEmail = u.Email
			break
		}
	}
	if customerEmail == "" && len(users) > 0 {
		customerEmail = users[0].Email
	}
	if customerEmail == "" {
		c.String(http.StatusBadRequest, "Email do proprietário não encontrado")
		return
	}

	farms, farmsErr := user_service.GetFarmsByOwnerDocument(sub.OwnerDocument, sub.OwnerDocumentType)
	if farmsErr != nil {
		c.String(http.StatusInternalServerError, "Erro ao buscar fazendas")
		return
	}
	quantity := int64(len(farms))
	if quantity == 0 {
		quantity = 1
	}

	checkoutURL, checkoutErr := billing_service.CreateCheckoutSessionForOwner(sub.Id, form.PriceID, quantity, customerEmail)
	if checkoutErr != nil {
		fmt.Printf("failed to create checkout session for owner: %v\n", checkoutErr)
		c.String(http.StatusInternalServerError, "Erro ao criar sessão de pagamento")
		return
	}

	c.Redirect(http.StatusSeeOther, checkoutURL)
}
