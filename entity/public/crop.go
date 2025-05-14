package entity_public

import "time"

type Crop struct {
	Id        uint8
	Name      string
	Product   uint8
	StartDate time.Time
	Farm      uint32
}
