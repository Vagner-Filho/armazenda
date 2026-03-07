package entity_public

import "github.com/shopspring/decimal"

type HumidityProgression struct {
	Id              uint32                    `json:"id"`
	Name            string                    `json:"name"`
	FarmId          *uint32                   `json:"farmId,omitempty"` // NULL for system default
	IsSystemDefault bool                      `json:"isSystemDefault"`
	Tiers           []HumidityProgressionTier `json:"tiers"`
}

type HumidityProgressionTier struct {
	Id                uint32          `json:"id"`
	ProgressionId     uint32          `json:"progressionId"`
	ThresholdHumidity decimal.Decimal `json:"thresholdHumidity"`
	DiscountValue     decimal.Decimal `json:"discountValue"`
}
