package billing_service

import (
	"armazenda/entity/public"
	"armazenda/model/subscription_model"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/price"
	"github.com/stripe/stripe-go/v85/subscription"
	"github.com/stripe/stripe-go/v85/webhook"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

func getPriceID() string {
	return os.Getenv("STRIPE_PRICE_ID")
}

func GetPublishableKey() string {
	return os.Getenv("STRIPE_PUBLISHABLE_KEY")
}

func CreateCheckoutSession(pendingRegistrationID uint32, priceID string) (string, error) {
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

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"pending_registration_id": fmt.Sprintf("%d", pendingRegistrationID),
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

		farmId, createErr := sm.CreateFarmAndUserFromPending(*pending)
		if createErr != nil {
			return fmt.Errorf("failed to create farm and user from pending registration: %w", createErr)
		}

		// Always store Stripe customer and subscription IDs on the farm.
		// This ensures downstream webhooks (subscription.updated, etc.) can locate the farm.
		customerID := ""
		if session.Customer != nil {
			customerID = session.Customer.ID
		}
		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}
		if customerID != "" || subscriptionID != "" {
			idErr := sm.SetFarmStripeIDs(farmId, customerID, subscriptionID)
			if idErr != nil {
				fmt.Printf("failed to set farm stripe IDs: %v\n", idErr)
			}
		}

		// Try to fetch full subscription details (status + current_period_end).
		// If this fails, customer.subscription.created/updated will fill them in shortly.
		if subscriptionID != "" {
			sub, subErr := getSubscription(subscriptionID)
			if subErr != nil {
				fmt.Printf("failed to fetch subscription details after checkout: %v\n", subErr)
			} else if sub != nil {
				periodEnd := subscriptionPeriodEnd(sub)
				status := string(sub.Status)
				updateErr := sm.UpdateFarmSubscription(farmId, customerID, subscriptionID, status, periodEnd)
				if updateErr != nil {
					fmt.Printf("failed to update farm subscription after checkout: %v\n", updateErr)
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

		farmId, farmErr := sm.GetFarmByStripeSubscriptionID(sub.ID)
		if farmErr != nil {
			return fmt.Errorf("farm not found for subscription %s: %w", sub.ID, farmErr)
		}

		periodEnd := subscriptionPeriodEnd(&sub)
		status := string(sub.Status)
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		updateErr := sm.UpdateFarmSubscription(farmId, customerID, sub.ID, status, periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update farm subscription: %w", updateErr)
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("failed to unmarshal subscription: %w", err)
		}

		farmId, farmErr := sm.GetFarmByStripeSubscriptionID(sub.ID)
		if farmErr != nil {
			return fmt.Errorf("farm not found for subscription %s: %w", sub.ID, farmErr)
		}

		periodEnd := subscriptionPeriodEnd(&sub)
		status := string(sub.Status)
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		updateErr := sm.UpdateFarmSubscription(farmId, customerID, sub.ID, status, periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update farm subscription: %w", updateErr)
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("failed to unmarshal subscription: %w", err)
		}

		farmId, farmErr := sm.GetFarmByStripeSubscriptionID(sub.ID)
		if farmErr != nil {
			return fmt.Errorf("farm not found for subscription %s: %w", sub.ID, farmErr)
		}

		periodEnd := subscriptionPeriodEnd(&sub)
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		updateErr := sm.UpdateFarmSubscription(farmId, customerID, sub.ID, "canceled", periodEnd)
		if updateErr != nil {
			return fmt.Errorf("failed to update farm subscription: %w", updateErr)
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

func GetSubscriptionStatus(farmId uint32) (string, error) {
	sm := subscription_model.GetSubscriptionModel()
	return sm.GetFarmSubscriptionStatus(farmId)
}

func CreatePendingAndCheckout(user entity_public.NewUser, priceID string) (string, entity_public.Toast) {
	sm := subscription_model.GetSubscriptionModel()

	// Create pending registration without session ID first
	pendingID, err := sm.CreatePendingRegistration(user, "")
	if err != nil {
		return "", entity_public.GetErrorToast(err.Error(), "")
	}

	// Create Stripe checkout session
	checkoutURL, sessionErr := CreateCheckoutSession(pendingID, priceID)
	if sessionErr != nil {
		// Clean up pending registration
		sm.DeletePendingRegistration(pendingID)
		return "", entity_public.GetErrorToast("Falha ao criar sessão de pagamento", "")
	}

	return checkoutURL, entity_public.GetSuccessToast("Redirecionando para pagamento", "")
}
