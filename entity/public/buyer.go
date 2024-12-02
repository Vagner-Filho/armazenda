package entity_public

import "strconv"

type Company struct {
	Id                uint8   `form:"id"`
	CompanyName       string  `form:"companyName" binding:"required"`
	FantasyName       string  `form:"fantasyName"`
	Cnpj              string  `form:"cnpj" binding:"required"`
	Address           Address `form:"address" binding:"required"`
	InscricaoEstadual string  `form:"inscricaoEstadual" binding:"required"`
}

func (c Company) GetBuyer() Buyer {
	if len(c.FantasyName) == 0 {
		return Buyer{
			Name: c.CompanyName,
			Id:   strconv.Itoa(int(c.Id)) + "-" + c.Cnpj,
		}
	}
	return Buyer{
		Name: c.FantasyName,
		Id:   strconv.Itoa(int(c.Id)) + "-" + c.Cnpj,
	}
}

type Personal struct {
	Id      uint8   `form:"id"`
	Name    string  `form:"name" binding:"required"`
	Cpf     string  `form:"cpf" binding:"required"`
	Address Address `form:"address" binding:"required"`
}

func (p Personal) GetBuyer() Buyer {
	return Buyer{
		Name: p.Name,
		Id:   strconv.Itoa(int(p.Id)) + "-" + p.Cpf,
	}
}

type Buyer struct {
	Name     string
	Id       string
	Selected bool
}
