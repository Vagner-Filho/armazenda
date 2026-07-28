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
	OwnerDocument           *string
	OwnerDocumentType       *int
	AdditionalIEs           []string
	StripePriceID           *string
	UF                      string
}

type CreateUserResult struct {
	Toast       Toast
	CheckoutURL string
}
