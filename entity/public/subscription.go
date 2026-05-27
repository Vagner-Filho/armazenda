package entity_public

import "time"

type PendingRegistration struct {
	Id                      uint32
	Email                   string
	Name                    string
	Passwd                  string
	Cpf                     string
	InscricaoEstadual       string
	Role                    string
	StripeCheckoutSessionId *string
	CreatedAt               time.Time
}

type FarmSubscription struct {
	StripeCustomerId             *string
	StripeSubscriptionId         *string
	Status                       string
	SubscriptionCurrentPeriodEnd *time.Time
}

type CreateUserResult struct {
	Toast       Toast
	CheckoutURL string
}
