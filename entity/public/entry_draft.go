package entity_public

import "github.com/shopspring/decimal"

type EntryDraft struct {
	Id      uint32           `form:"id"`
	Name    string           `form:"name" binding:"required"`
	Field   uint16           `form:"field" binding:"required"`
	Crop    uint8            `form:"crop" binding:"required"`
	Vehicle string           `form:"vehiclePlate"`
	Tare    *decimal.Decimal `form:"tare, omitempty" binding:"gte=0"`
	Farm    uint32           `form:"farm" binding:"gte=0"`
	Origin  *uint32          `form:"origin,omitempty"`
}

type DisplayEntryDraft struct {
	Id      uint32           `form:"id"`
	Name    string           `form:"name"`
	Field   string           `form:"field"`
	Crop    string           `form:"crop"`
	Vehicle string           `form:"vehiclePlate"`
	Tare    *decimal.Decimal `form:"tare"`
	Origin  string           `form:"origin"`
}
