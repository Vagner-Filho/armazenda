package entity_public

import "github.com/shopspring/decimal"

type EntryDraft struct {
	Id      uint32           `form:"id" json:"id"`
	Name    string           `form:"name" binding:"required" json:"name"`
	Field   uint16           `form:"field" binding:"required" json:"field"`
	Crop    uint8            `form:"crop" binding:"required" json:"crop"`
	Vehicle uint16           `form:"vehiclePlate" json:"vehiclePlate"`
	Tare    *decimal.Decimal `form:"tare,omitempty" json:"tare,omitempty"`
	Farm    uint32           `form:"farm" binding:"gte=0" json:"farm"`
	Origin  *uint32          `form:"origin,omitempty" json:"origin,omitempty"`
}

type DisplayEntryDraft struct {
	Id      uint32           `form:"id" json:"id"`
	Name    string           `form:"name" json:"name"`
	Field   string           `form:"field" json:"field"`
	Crop    string           `form:"crop" json:"crop"`
	Vehicle string           `form:"vehiclePlate" json:"vehiclePlate"`
	Tare    *decimal.Decimal `form:"tare" json:"tare"`
	Origin  string           `form:"origin" json:"origin"`
}
