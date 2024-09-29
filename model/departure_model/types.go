package departure_model

import "armazenda/model/entry_model"

type BaseDeparture struct {
    DepartureDate int64
    Product entry_model.Grain
    VehiclePlate string
    Weight float64
}

type Departure struct {
    Manifest uint32
    BaseDeparture
}
