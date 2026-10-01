package entity_public

type Product struct {
	Id   uint8  `form:"id" json:"id"`
	Name string `form:"name" binding:"required" json:"name"`
	NCM  string `form:"ncm" binding:"required" json:"ncm"`
}

// FarmProduct is a user-created product scoped to a single farm, used by
// the detached NF-e feature. It is independent from the global catalog.
type FarmProduct struct {
	Product
	FarmId uint32 `form:"-" json:"farmId"`
}

// ProductDisplay is one row of the products management page: either a
// global catalog product (read-only) or a farm-created one (editable).
type ProductDisplay struct {
	Id     uint8
	Name   string
	NCM    string
	IsFarm bool
}
