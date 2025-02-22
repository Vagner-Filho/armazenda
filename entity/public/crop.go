package entity_public

import "time"

type Crop struct {
	Id        uint8
	Name      string
	StartDate time.Time
	Product   uint8
	Selected  bool `db:"-"`
}
