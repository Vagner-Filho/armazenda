package entity_public

type Field struct {
	Id       uint16
	Name     string
	Selected bool `db:"-"`
}
