package entity_public

type BuyerCompany struct {
	Id                uint8   `form:"id"`
	CompanyName       string  `form:"companyName" binding:"required"`
	FantasyName       string  `form:"fantasyName"`
	Cnpj              string  `form:"cnpj" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
}

type BuyerPerson struct {
	Id                uint8   `form:"id"`
	Name              string  `form:"name" binding:"required"`
	Cpf               string  `form:"cpf" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
}

type Buyer struct {
	Ie string
	Id string
}

type BuyerDisplay struct {
	Id   uint8
	Name string
}
