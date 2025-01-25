package entity_public

import "time"

type Crop struct {
	Id        uint8
	Name      string
	StartDate time.Time
	Selected  bool `db:"-"`
}
