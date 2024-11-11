package entity_public

type Company struct {
	Id                int8    `form:"id"`
	CompanyName       string  `form:"companyName" binding:"required"`
	FantasyName       string  `form:"fantasyName" binding:"required"`
	Cnpj              string  `form:"cnpj" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
}

func (c Company) GetName() string {
    return c.FantasyName
}

func (c Company) GetId() int8 {
    return c.Id
}

type Personal struct {
	Id      int8    `form:"id"`
	Name    string  `form:"name" binding:"required"`
	Cpf     string  `form:"cpf" binding:"required"`
	Address Address `form:"address" binding:"required"`
}

func (p Personal) GetName() string {
    return p.Name
}

func (p Personal) GetId() int8 {
    return p.Id
}

type Buyer interface {
	GetName() string
	GetId() int8
}
