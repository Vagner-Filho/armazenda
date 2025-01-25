package entity_public

type Vehicle struct {
	Plate    string
	Name     string
	Selected bool `db:"-"`
}
