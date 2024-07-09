package entity

type Grain int

const (
	Corn Grain = iota
	Soy  Grain = iota
)

var GrainMap = make(map[Grain]string)

type GrainEntry struct {
	Product      Grain
	Field        string
	HarvestYear  int64
	Vehicle      string
	VehiclePlate string
	GrossWeight  float64
	Tare         float64
	Humidity     string
	ArrivalDate  int64
}
