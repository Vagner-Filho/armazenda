package entity_public

type Departure struct {
	Manifest      uint32  `form:"manifest"`
	DepartureDate string  `form:"departureDate" binding:"required"`
	Product       Grain   `form:"product"  binding:"gte=0"`
	VehiclePlate  string  `form:"vehiclePlate" binding:"required"`
	Weight        float64 `form:"weight" binding:"required"`
	Buyer         string  `form:"buyer" binding:"required"`
}
