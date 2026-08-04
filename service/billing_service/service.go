package billing_service

import (
	"armazenda/entity/public"
	"armazenda/model/owner_subscription_model"
	"armazenda/model/subscription_model"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v85"
	portal "github.com/stripe/stripe-go/v85/billingportal/session"
	"github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/price"
	"github.com/stripe/stripe-go/v85/subscription"
	"github.com/stripe/stripe-go/v85/webhook"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

type tempOwnerSession struct {
	ownerDocument string
	expiresAt     time.Time
}

var (
	tempSessions   = make(map[string]tempOwnerSession)
	tempSessionsMu sync.RWMutex
)

func CreateTempOwnerSession(ownerDocument string) string {
	tempSessionsMu.Lock()
	defer tempSessionsMu.Unlock()

	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	tempSessions[token] = tempOwnerSession{
		ownerDocument: ownerDocument,
		expiresAt:     time.Now().Add(10 * time.Minute),
	}
	return token
}

func ValidateTempOwnerSession(token string) (string, bool) {
	tempSessionsMu.Lock()
	defer tempSessionsMu.Unlock()

	session, ok := tempSessions[token]
	if !ok {
		return "", false
	}
	if time.Now().After(session.expiresAt) {
		delete(tempSessions, token)
		return "", false
	}
	return session.ownerDocument, true
}

func getPriceID() string {
	return os.Getenv("STRIPE_PRICE_ID")
}

func GetPublishableKey() string {
	return os.Getenv("STRIPE_PUBLISHABLE_KEY")
}

func formatBRL(value float64) string {
	s := fmt.Sprintf("%.2f", value)
	return strings.ReplaceAll(s, ".", ",")
}

type PricingTier struct {
	TierKey                 string
	ProductName             string
	ProductDescription      string
	MonthlyPriceID          string
	YearlyPriceID           string
	MonthlyAmount           string
	YearlyTotal             string
	YearlyMonthlyEquivalent string
	SavingsPercent          string
	Features                []string
}

func ListStripePrices() ([]*stripe.Price, error) {
	params := &stripe.PriceListParams{
		Active: stripe.Bool(true),
		Type:   stripe.String("recurring"),
	}
	params.AddExpand("data.product")
	iter := price.List(params)
	var prices []*stripe.Price
	for iter.Next() {
		prices = append(prices, iter.Price())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return prices, nil
}

func GetPricingTiers() ([]PricingTier, error) {
	prices, err := ListStripePrices()
	if err != nil {
		return nil, err
	}

	type tierGroup struct {
		tierKey string
		monthly *stripe.Price
		yearly  *stripe.Price
		product *stripe.Product
	}

	groups := make(map[string]*tierGroup)
	for _, p := range prices {
		if p.Product == nil {
			continue
		}
		tierKey := p.Product.ID

		g, ok := groups[tierKey]
		if !ok {
			g = &tierGroup{tierKey: tierKey, product: p.Product}
			groups[tierKey] = g
		}

		if p.Recurring != nil {
			switch p.Recurring.Interval {
			case stripe.PriceRecurringIntervalMonth:
				g.monthly = p
			case stripe.PriceRecurringIntervalYear:
				g.yearly = p
			}
		}
	}

	var tiers []PricingTier
	for _, g := range groups {
		if g.monthly == nil || g.yearly == nil {
			continue
		}

		monthlyAmount := float64(g.monthly.UnitAmount) / 100.0
		yearlyTotal := float64(g.yearly.UnitAmount) / 100.0
		yearlyMonthlyEquivalent := yearlyTotal / 12.0
		savingsPercent := math.Round((1.0 - yearlyMonthlyEquivalent/monthlyAmount) * 100)

		var features []string
		if g.product.MarketingFeatures != nil {
			for _, f := range g.product.MarketingFeatures {
				if f.Name != "" {
					features = append(features, f.Name)
				}
			}
		}

		tiers = append(tiers, PricingTier{
			TierKey:                 g.tierKey,
			ProductName:             g.product.Name,
			ProductDescription:      g.product.Description,
			MonthlyPriceID:          g.monthly.ID,
			YearlyPriceID:           g.yearly.ID,
			MonthlyAmount:           formatBRL(monthlyAmount),
			YearlyTotal:             formatBRL(yearlyTotal),
			YearlyMonthlyEquivalent: formatBRL(yearlyMonthlyEquivalent),
			SavingsPercent:          fmt.Sprintf("%.0f", savingsPercent),
			Features:                features,
		})
	}

	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].TierKey < tiers[j].TierKey
	})

	return tiers, nil
}

func ResolveTierKey(priceID string) (string, error) {
	if priceID == "" {
		return "", fmt.Errorf("priceID is empty")
	}

	stripePrice, err := GetStripePrice(priceID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch stripe price: %w", err)
	}

	if stripePrice.Product == nil || stripePrice.Product.ID == "" {
		return "", fmt.Errorf("stripe price has no product")
	}

	osm := owner_subscription_model.GetOwnerSubscriptionModel()
	tierKey, err := osm.GetTierKeyByStripeProductID(stripePrice.Product.ID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve tier key for product %s: %w", stripePrice.Product.ID, err)
	}

	return tierKey, nil
}

func CreateCheckoutSession(pendingRegistrationID uint32, priceID string, quantity int64) (string, error) {
	if priceID == "" {
		priceID = getPriceID()
	}

	successURL := os.Getenv("STRIPE_SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:8100/payment/success"
	}
	cancelURL := os.Getenv("STRIPE_CANCEL_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:8100/payment/cancel"
	}

	tierKey, tierErr := ResolveTierKey(priceID)
	if tierErr != nil {
		fmt.Printf("warning: failed to resolve tier key for checkout: %v\n", tierErr)
		tierKey = ""
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(quantity),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"pending_registration_id": fmt.Sprintf("%d", pendingRegistrationID),
			"tier_key":                tierKey,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"pending_registration_id": fmt.Sprintf("%d", pendingRegistrationID),
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	// Update pending registration with the session ID
	sm := subscription_model.GetSubscriptionModel()
	updateErr := sm.UpdatePendingRegistrationSessionID(pendingRegistrationID, s.ID)
	if updateErr != nil {
		fmt.Printf("failed to update pending registration with session ID: %v\n", updateErr)
	}

	return s.URL, nil
}

func CreateCheckoutSessionForOwner(ownerSubscriptionID uint32, priceID string, quantity int64, customerEmail string) (string, error) {
	if priceID == "" {
		return "", fmt.Errorf("priceID is required")
	}

	successURL := os.Getenv("STRIPE_SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:8100/payment/success"
	}
	cancelURL := os.Getenv("STRIPE_CANCEL_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:8100/payment/cancel"
	}

	tierKey, tierErr := ResolveTierKey(priceID)
	if tierErr != nil {
		fmt.Printf("warning: failed to resolve tier key for checkout: %v\n", tierErr)
		tierKey = ""
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(quantity),
			},
		},
		SuccessURL:    stripe.String(successURL),
		CancelURL:     stripe.String(cancelURL),
		CustomerEmail: stripe.String(customerEmail),
		Metadata: map[string]string{
			"owner_subscription_id": fmt.Sprintf("%d", ownerSubscriptionID),
			"tier_key":              tierKey,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"owner_subscription_id": fmt.Sprintf("%d", ownerSubscriptionID),
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}

func HandleWebhook(payload []byte, sigHeader string) error {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET not configured")
	}

	event, err := webhook.ConstructEvent(payload, sigHeader, secret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	sm := subscription_model.GetSubscriptionModel()

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return fmt.Errorf("failed to unmarshal checkout session: %w", err)
		}

		// Check for existing owner subscription checkout first
		ownerSubIDStr, hasOwnerSub := session.Metadata["owner_subscription_id"]
		if hasOwnerSub {
			var ownerSubID uint32
			fmt.Sscanf(ownerSubIDStr, "%d", &ownerSubID)

			customerID := ""
			if session.Customer != nil {
				customerID = session.Customer.ID
			}
			subscriptionID := ""
			if session.Subscription != nil {
				subscriptionID = session.Subscription.ID
			}

			if subscriptionID != "" {
				osm := owner_subscription_model.GetOwnerSubscriptionModel()
				var periodEnd time.Time
				var status string
				sub, subErr := getSubscription(subscriptionID)
				if subErr != nil {
					fmt.Printf("failed to fetch subscription details after checkout: %v\n", subErr)
					periodEnd = time.Now()
					status = "active"
				} else if sub != nil {
					periodEnd = subscriptionPeriodEnd(sub)
					status = string(sub.Status)
				}

				tierKey := session.Metadata["tier_key"]
				if tierKey == "" {
					resolvedTier, tierErr := ResolveTierKey(session.LineItems.Data[0].Price.ID)
					if tierErr == nil {
						tierKey = resolvedTier
					}
				}

				updateErr := osm.UpdateFromCheckout(ownerSubID, customerID, subscriptionID, status, periodEnd, tierKey)
				if updateErr != nil {
					fmt.Printf("failed to update owner subscription from checkout: %v\n", updateErr)
				}
			}
			return nil
		}

		pendingIDStr, ok := session.Metadata["pending_registration_id"]
		if !ok {
			return fmt.Errorf("missing pending_registration_id in checkout session metadata")
		}

		var pendingID uint32
		fmt.Sscanf(pendingIDStr, "%d", &pendingID)

		pending, getErr := sm.GetPendingRegistrationBySessionID(session.ID)
		if getErr != nil || pending == nil {
			// Fallback: try to find by ID
			pending, getErr = sm.GetPendingRegistrationByID(pendingID)
			if getErr != nil || pending == nil {
				return fmt.Errorf("pending registration not found for session %s: %w", session.ID, getErr)
			}
		}

		_, farmIds, createErr := sm.CreateFarmAndUserFromPending(*pending)
		if createErr != nil {
			return fmt.Errorf("failed to create farm and user from pending registration: %w", createErr)
		}

		// Create owner_subscription record
		customerID := ""
		if session.Customer != nil {
			customerID = session.Customer.ID
		}
		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}

		ownerDoc := pending.Cpf
		ownerDocType := 1
		if pending.OwnerDocument != nil && *pending.OwnerDocument != "" {
			ownerDoc = *pending.OwnerDocument
			if pending.OwnerDocumentType != nil {
				ownerDocType = *pending.OwnerDocumentType
			}
		}

		if subscriptionID != "" {
			osm := owner_subscription_model.GetOwnerSubscriptionModel()
			var periodEnd time.Time
			var status string
			sub, subErr := getSubscription(subscriptionID)
			if subErr != nil {
				fmt.Printf("failed to fetch subscription details after checkout: %v\n", subErr)
				periodEnd = time.Now()
				status = "active"
			} else if sub != nil {
				periodEnd = subscriptionPeriodEnd(sub)
				status = string(sub.Status)
			}

			// Resolve tier key from pending registration's selected price
			tierKey := ""
			if pending.StripePriceID != nil && *pending.StripePriceID != "" {
				resolvedTier, tierErr := ResolveTierKey(*pending.StripePriceID)
				if tierErr != nil {
					fmt.Printf("failed to resolve tier key from price %s: %v\n", *pending.StripePriceID, tierErr)
				} else {
					tierKey = resolvedTier
				}
			}
			// Fallback to session metadata if pending registration has no price
			if tierKey == "" {
				tierKey = session.Metadata["tier_key"]
			}

			for _, farmId := range farmIds {
				_, osErr := osm.Create(farmId, ownerDoc, ownerDocType, customerID, subscriptionID, status, periodEnd, tierKey)
				if osErr != nil {
					fmt.Printf("failed to create owner subscription: %v\n", osErr)
				}
			}
		}

		delErr := sm.DeletePendingRegistration(pending.Id)
		if delErr != nil {
			fmt.Printf("failed to delete pending registration: %v\n", delErr)
		}

	case "customer.subscription.created":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("failed to unmarshal subscription: %w", err)
		}

		osm := owner_subscription_model.GetOwnerSubscriptionModel()
		periodEnd := subscriptionPeriodEnd(&sub)
		status := string(sub.Status)
		updateErr := osm.UpdateStatus(sub.ID, status, periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update owner subscription: %w", updateErr)
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("failed to unmarshal subscription: %w", err)
		}

		osm := owner_subscription_model.GetOwnerSubscriptionModel()
		periodEnd := subscriptionPeriodEnd(&sub)
		status := string(sub.Status)
		updateErr := osm.UpdateStatus(sub.ID, status, periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update owner subscription: %w", updateErr)
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("failed to unmarshal subscription: %w", err)
		}

		osm := owner_subscription_model.GetOwnerSubscriptionModel()
		periodEnd := subscriptionPeriodEnd(&sub)
		updateErr := osm.UpdateStatus(sub.ID, "canceled", periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update owner subscription: %w", updateErr)
		}

	}

	return nil
}

func getSubscription(subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	sub, err := subscription.Get(subID, params)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func GetStripeSubscription(subID string) (*stripe.Subscription, error) {
	if len(stripe.Key) == 0 {
		return nil, fmt.Errorf("chave de acesso não encontrada")
	}
	return getSubscription(subID)
}

func GetStripeCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{}
	return session.Get(sessionID, params)
}

func CancelSubscriptionAtPeriodEnd(subID string) error {
	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(false),
		Prorate:    stripe.Bool(false),
	}
	sub, err := subscription.Cancel(subID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	_ = sub
	return nil
}

func GetStripePrice(priceID string) (*stripe.Price, error) {
	params := &stripe.PriceParams{}
	return price.Get(priceID, params)
}

func subscriptionPeriodEnd(sub *stripe.Subscription) time.Time {
	if sub != nil && sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].CurrentPeriodEnd > 0 {
		return time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
	}
	return time.Now()
}

func GetSubscriptionPeriodEnd(sub *stripe.Subscription) time.Time {
	return subscriptionPeriodEnd(sub)
}

func CreateCustomerPortalSession(customerID string) (string, error) {
	if customerID == "" {
		return "", fmt.Errorf("customer_id is empty")
	}

	returnURL := os.Getenv("STRIPE_PORTAL_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:8100/"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	s, err := portal.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}

func CreatePendingAndCheckout(user entity_public.NewUser, priceID string, quantity int64) (string, entity_public.Toast) {
	sm := subscription_model.GetSubscriptionModel()

	// Create pending registration without session ID first
	pendingID, err := sm.CreatePendingRegistration(user, "")
	if err != nil {
		return "", entity_public.GetErrorToast(err.Error(), "")
	}

	// Create Stripe checkout session
	checkoutURL, sessionErr := CreateCheckoutSession(pendingID, priceID, quantity)
	if sessionErr != nil {
		// Clean up pending registration
		sm.DeletePendingRegistration(pendingID)
		return "", entity_public.GetErrorToast("Falha ao criar sessão de pagamento", "")
	}

	return checkoutURL, entity_public.GetSuccessToast("Redirecionando para pagamento", "")
}
