package entity_public

type ProductiveField struct {
	Name         string  `json:"name"`
	Productivity float64 `json:"productivity"`
}

type ProductiveFields struct {
	Nominal  *ProductiveField `json:"nominal"`
	Relative *ProductiveField `json:"relative"`
}
