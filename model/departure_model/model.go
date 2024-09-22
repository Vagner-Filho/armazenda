package departure_model

import "armazenda/model/entry_model"

type Departure struct {
    Manifest uint32
    DepartureDate int64
    Product entry_model.Grain
    VehiclePlate string
    Weight float64
}

var departures = []Departure {
    {
        Manifest: 1,
        DepartureDate: 1726967334411,
        Product: 0,
        VehiclePlate: "OPA 2312",
        Weight: 20392,
    },
    {
        Manifest: 2,
        DepartureDate: 1726967334411,
        Product: 1,
        VehiclePlate: "OPA 2312",
        Weight: 12398,
    },
    {
        Manifest: 3,
        DepartureDate: 1726967334411,
        Product: 0,
        VehiclePlate: "OPA 2312",
        Weight: 40242,
    },
}

func GetDepartures() []Departure {
    return departures
}
