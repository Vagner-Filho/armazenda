package entity

type Grain int

const (
	Corn Grain = iota
	Soy  Grain = iota
)

var GrainMap = make(map[Grain]string)

type GrainEntry struct {
	Waybill      uint32
	Product      Grain
	Field        string
	Harvest      string
	Vehicle      string
	VehiclePlate string
	GrossWeight  float64
	Tare         float64
	Humidity     string
	ArrivalDate  int64
}
