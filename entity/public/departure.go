package entity_public

type Departure struct {
	Manifest      uint32  `form:"manifest"`
	DepartureDate int64   `form:"departureDate" binding:"required"`
	Product       Grain   `form:"product"  binding:"gte=0"`
	VehiclePlate  string  `form:"vehiclePlate" binding:"required"`
	Weight        float64 `form:"weight" binding:"required"`
	Address       int8    `form:"address" binding:"required"`
}
