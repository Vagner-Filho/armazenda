package entity_public

import "github.com/shopspring/decimal"

type Field struct {
	Id       uint16
	Name     string
	Selected bool `db:"-"`
	Farm     uint32
	Hectares decimal.Decimal
}
