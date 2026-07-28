package entity_public

import "time"

type UserFarmAccess struct {
	UserID    uint32
	FarmID    uint32
	IsAllowed bool
}

type OwnerSubscription struct {
	Id                           uint32
	OwnerDocument                string
	OwnerDocumentType            int
	StripeCustomerId             *string
	StripeSubscriptionId         *string
	SubscriptionStatus           string
	SubscriptionCurrentPeriodEnd *time.Time
	CreatedAt                    time.Time
	TierKey                      *string
}
