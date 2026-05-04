package entity_public

import (
	"github.com/shopspring/decimal"
)

type DepartureDraft struct {
	Id        uint32           `form:"id" json:"id"`
	Name      string           `form:"name" binding:"required" json:"name"`
	Origin    *uint32          `form:"origin,omitempty" json:"origin,omitempty"`
	Crop      uint32           `form:"crop" json:"crop"`
	Vehicle   uint16           `form:"vehiclePlate" json:"vehiclePlate"`
	Tare      *decimal.Decimal `form:"tare,omitempty" json:"tare,omitempty"`
	Farm      uint32           `form:"farm" json:"farm"`
	Recipient *uint32          `form:"recipient,omitempty" json:"recipient,omitempty"`
}

type DisplayDepartureDraft struct {
	Id      uint32
	Name    string
	Person  *string
	Crop    string
	Vehicle string
	Tare    *decimal.Decimal
}
