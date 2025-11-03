package entity_public

type ProductiveField struct {
	Name         string
	Productivity float64
}

type ProductiveFields struct {
	Nominal  []ProductiveField
	Relative []ProductiveField
}
