package entity_public

// CfopOption is one selectable CFOP in the detached NF-e selector. Options
// come from the system-wide cfop catalog or from farm-registered codes
// (IsFarm). UseCount carries the farm's local use ranking; the catalog and
// farm_cfop tables themselves are never mutated by emission.
type CfopOption struct {
	Code              string
	Description       string
	OriginDestination string
	UseCount          int
	IsFarm            bool
}
