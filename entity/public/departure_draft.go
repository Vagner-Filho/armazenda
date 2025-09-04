package entity_public

import (
	"time"

	"github.com/shopspring/decimal"
)

type DepartureDraft struct {
	Id        uint32          `form:"id"`
	Name      string          `form:"name" binding:"required"`
	Person    *uint32         `form:"origin,omitempty"`
	Crop      uint32          `form:"crop"`
	Vehicle   string          `form:"vehiclePlate"`
	Tare      decimal.Decimal `form:"tare"`
	Farm      uint32          `form:"farm"`
	StartedAt time.Time       `form:"started_at"`
}

type DisplayDepartureDraft struct {
	Id      uint32
	Name    string
	Person  *string
	Crop    string
	Vehicle string
	Tare    float64
}
