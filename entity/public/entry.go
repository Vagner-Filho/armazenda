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
	ArrivalDate string  `form:"arrivalDate" binding:"required"`
}

type EntryFilter struct {
	Manifest        uint32  `form:"manifest"`
	Product         Grain   `form:"product"`
	Field           uint32  `form:"field"`
	Harvest         string  `form:"harvest"`
	Vehicle         string  `form:"vehiclePlate"`
	GrossWeightFrom float64 `form:"grossWeightFrom"`
	GrossWeightTo   float64 `form:"grossWeightTo"`
	TareFrom        float64 `form:"tareFrom"`
	TareTo          float64 `form:"tareTo"`
	NetWeightFrom   float64 `form:"netWeightFrom"`
	NetWeightTo     float64 `form:"netWeightTo"`
	HumidityFrom    string  `form:"humidityFrom"`
	HumidityTo      string  `form:"humidityTo"`
	ArrivalDateFrom string  `form:"arrivalDateFrom"`
	ArrivalDateTo   string  `form:"arrivalDateTo"`
}
