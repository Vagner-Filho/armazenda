package entity_public

import "github.com/shopspring/decimal"

type CargoWeight struct {
	GrossWeight decimal.Decimal `form:"grossWeight" binding:"required,gte=0"`
	Tare        decimal.Decimal `form:"tare" binding:"required,gte=0"`
	NetWeight   decimal.Decimal `form:"netWeight" binding:"required,gte=0"`
}

func (cw CargoWeight) ToDTO() CargoWeightDTO {
	gw, _ := cw.GrossWeight.Float64()
	tare, _ := cw.Tare.Float64()
	nw, _ := cw.NetWeight.Float64()
	return CargoWeightDTO{
		GrossWeight: gw,
		Tare:        tare,
		NetWeight:   nw,
	}
}

type CargoWeightDTO struct {
	GrossWeight float64 `form:"grossWeight" binding:"required,gte=0"`
	Tare        float64 `form:"tare" binding:"required,gte=0"`
	NetWeight   float64 `form:"netWeight" binding:"required,gte=0"`
}

func (cw CargoWeightDTO) ToEntity() CargoWeight {
	gw := decimal.NewFromFloat(cw.GrossWeight)
	tare := decimal.NewFromFloat(cw.Tare)
	nw := decimal.NewFromFloat(cw.NetWeight)
	return CargoWeight{
		GrossWeight: gw,
		Tare:        tare,
		NetWeight:   nw,
	}
}
