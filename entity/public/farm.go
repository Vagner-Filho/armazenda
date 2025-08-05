package entity_public

type Farm struct {
	Id   uint32 `form:"id"`
	Name string `form:"name" binding:"required"`
	*Address
	HumidityDiscount float64 `form:"humidityDiscount" binding:"required"`
}
