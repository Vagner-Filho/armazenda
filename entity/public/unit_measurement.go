package entity_public

// UnitMeasurement is the canonical code for an NF-e item's unit of
// measurement (uCom/uTrib). It is the single source of truth used by every
// unit <select> in the system.
type UnitMeasurement string

const (
	UnitMeasurementKilogram UnitMeasurement = "KG"
	UnitMeasurementGram     UnitMeasurement = "G"
	UnitMeasurementTon      UnitMeasurement = "T"
	UnitMeasurementTonAlt   UnitMeasurement = "TON"
	UnitMeasurementStone    UnitMeasurement = "ST"
	UnitMeasurementSack     UnitMeasurement = "SC"
	UnitMeasurementLiter    UnitMeasurement = "L"
	UnitMeasurementMeter    UnitMeasurement = "M"
	UnitMeasurementUnit     UnitMeasurement = "UN"
	UnitMeasurementPiece    UnitMeasurement = "PC"
)

// UnitMeasurementOption pairs a unit code with its pt-br label for rendering.
type UnitMeasurementOption struct {
	Code  string
	Label string
}

// UnitMeasurementOptions is the single source of truth for unit <select>s.
var UnitMeasurementOptions = []UnitMeasurementOption{
	{Code: "KG", Label: "Quilograma"},
	{Code: "G", Label: "Grama"},
	{Code: "T", Label: "Tonelada"},
	{Code: "TON", Label: "Tonelada"},
	{Code: "ST", Label: "Stone"},
	{Code: "SC", Label: "Saca"},
	{Code: "L", Label: "Litro"},
	{Code: "M", Label: "Metro"},
	{Code: "UN", Label: "Unidade"},
	{Code: "PC", Label: "Peça"},
}

// UnitMeasurements is the list of canonical unit codes.
var UnitMeasurements = []UnitMeasurement{
	UnitMeasurementKilogram,
	UnitMeasurementGram,
	UnitMeasurementTon,
	UnitMeasurementTonAlt,
	UnitMeasurementStone,
	UnitMeasurementSack,
	UnitMeasurementLiter,
	UnitMeasurementMeter,
	UnitMeasurementUnit,
	UnitMeasurementPiece,
}

// UnitMeasurementLabel returns the pt-br label for a unit code.
func UnitMeasurementLabel(code string) string {
	for _, o := range UnitMeasurementOptions {
		if o.Code == code {
			return o.Label
		}
	}
	return code
}
