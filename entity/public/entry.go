package entity_public

type Entry struct {
	Manifest    uint32  `form:"manifest"`
	Product     Grain   `form:"product" binding:"gte=0"`
	Field       uint32  `form:"field" binding:"required"`
	Harvest     string  `form:"harvest" binding:"required"`
	Vehicle     string  `form:"vehiclePlate"`
	GrossWeight float64 `form:"grossWeight" binding:"required"`
	Tare        float64 `form:"tare" binding:"required"`
	NetWeight   float64 `form:"netWeight"`
	Humidity    string  `form:"humidity" binding:"required"`
	ArrivalDate int64   `form:"arrivalDate" binding:"required"`
}
