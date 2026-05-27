package billing_router

import (
	"armazenda/service/billing_service"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func BillingRoutes(router *gin.Engine) {
	router.POST("/stripe/webhook", stripeWebhook)
	router.GET("/pricing", pricingPage)
	router.GET("/payment/success", paymentSuccessPage)
	router.GET("/payment/cancel", paymentCancelPage)
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

func formatBRL(value float64) string {
	s := fmt.Sprintf("%.2f", value)
	return strings.ReplaceAll(s, ".", ",")
}

func pricingPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")

	monthlyPriceID := os.Getenv("STRIPE_MONTHLY_PRICE_ID")
	yearlyPriceID := os.Getenv("STRIPE_YEARLY_PRICE_ID")

	monthlyPrice, err := billing_service.GetStripePrice(monthlyPriceID)
	if err != nil {
		fmt.Printf("failed to fetch monthly price from Stripe: %v\n", err)
		c.String(http.StatusInternalServerError, "Falha ao carregar preços. Tente novamente mais tarde.")
		return
	}

	yearlyPrice, err := billing_service.GetStripePrice(yearlyPriceID)
	if err != nil {
		fmt.Printf("failed to fetch yearly price from Stripe: %v\n", err)
		c.String(http.StatusInternalServerError, "Falha ao carregar preços. Tente novamente mais tarde.")
		return
	}

	monthlyAmount := float64(monthlyPrice.UnitAmount) / 100.0
	yearlyTotal := float64(yearlyPrice.UnitAmount) / 100.0
	yearlyMonthlyEquivalent := yearlyTotal / 12.0
	savingsPercent := math.Round((1.0 - yearlyMonthlyEquivalent/monthlyAmount) * 100)

	c.HTML(http.StatusOK, "pricing.html", gin.H{
		"CSPNonce":                nonce.(string),
		"PublishableKey":          billing_service.GetPublishableKey(),
		"MonthlyPriceID":          monthlyPriceID,
		"YearlyPriceID":           yearlyPriceID,
		"MonthlyAmount":           formatBRL(monthlyAmount),
		"YearlyTotal":             formatBRL(yearlyTotal),
		"YearlyMonthlyEquivalent": formatBRL(yearlyMonthlyEquivalent),
		"SavingsPercent":          fmt.Sprintf("%.0f", savingsPercent),
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
