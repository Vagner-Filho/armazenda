package entity_public

type Farm struct {
	Id                uint32  `form:"id"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
	Name              string  `form:"name" binding:"required"`
	HumidityDiscount  float64 `form:"humidityDiscount" binding:"required"`
	*Address
}
