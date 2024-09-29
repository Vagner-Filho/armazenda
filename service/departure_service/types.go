package departure_service

type ReadableDeparture struct {
	Manifest      uint32
	DepartureDate string
	Product       string
	VehiclePlate  string
	Weight        float64
}
