package entity_public

// ProductVolumeStat holds the farm-level totals of one product:
// weight received (entries) and weight sent (departures).
type ProductVolumeStat struct {
	ProductName    string
	EntryTotal     float64
	DepartureTotal float64
}

// FieldProductTotal is one (field, product) aggregate row used to build the
// field charts.
type FieldProductTotal struct {
	FieldName    string
	ProductName  string
	TotalWeight  float64
	Productivity float64 // kg/ha (only filled by the relative query)
}

// ProductDataset is one chart series: all values of a single product,
// aligned with the chart labels (fields).
type ProductDataset struct {
	Product string    `json:"product"`
	Values  []float64 `json:"values"`
}

// FieldProductSeries is the grouped-bar chart payload for one chart:
// labels are fields, and each dataset is a product.
type FieldProductSeries struct {
	Labels   []string
	Datasets []ProductDataset
}

// FieldProductCharts carries both charts: nominal volume and relative
// productivity (kg/ha).
type FieldProductCharts struct {
	Nominal  []FieldProductTotal
	Relative []FieldProductTotal
}
