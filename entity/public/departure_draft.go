package entity_public

import (
	"github.com/shopspring/decimal"
)

type DepartureDraft struct {
	Id        uint32           `form:"id"`
	Name      string           `form:"name" binding:"required"`
	Origin    *uint32          `form:"origin,omitempty"`
	Crop      uint32           `form:"crop"`
	Vehicle   uint16           `form:"vehiclePlate"`
	Tare      *decimal.Decimal `form:"tare,omitempty"`
	Farm      uint32           `form:"farm"`
	Recipient *uint32          `form:"recipient,omitempty"`
}

type DisplayDepartureDraft struct {
	Id      uint32
	Name    string
	Person  *string
	Crop    string
	Vehicle string
	Tare    *decimal.Decimal
}
