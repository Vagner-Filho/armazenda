package buyer_model

import (
	entity_public "armazenda/entity/public"
)

var Buyer_addresses = []entity_public.Address{
	{
		Street:       "Rua A",
		Cep:          "12345-678",
		Number:       10,
		Complement:   "Ap 101",
		Neighborhood: "Bairro A",
		City:         "Cidade A",
		State:        "Estado A",
		Email:        "a@example.com",
		PhoneNumber:  "1234567890",
	},
	{
		Street:       "Rua B",
		Cep:          "87654-321",
		Number:       20,
		Complement:   "",
		Neighborhood: "Bairro B",
		City:         "Cidade B",
		State:        "Estado B",
		Email:        "b@example.com",
		PhoneNumber:  "9876543210",
	},
	{
		Street:       "Rua C",
		Cep:          "54321-098",
		Number:       30,
		Complement:   "Sala 302",
		Neighborhood: "Bairro C",
		City:         "Cidade C",
		State:        "Estado C",
		Email:        "c@example.com",
		PhoneNumber:  "0123456789",
	},
}

var companies = []entity_public.Company{
	{
		Id:                0,
		CompanyName:       "Company A",
		FantasyName:       "Fantasy A",
		Cnpj:              "12345678901",
		Address:           Buyer_addresses[0],
		InscricaoEstadual: "123456789",
	},
	{
		Id:                1,
		CompanyName:       "Company B",
		FantasyName:       "Fantasy B",
		Cnpj:              "12345678901",
		Address:           Buyer_addresses[1],
		InscricaoEstadual: "987654321",
	},
	{
		Id:                2,
		CompanyName:       "Company C",
		FantasyName:       "Fantasy C",
		Cnpj:              "12345678901",
		Address:           Buyer_addresses[2],
		InscricaoEstadual: "789012345",
	},
}

var personals = []entity_public.Personal{
	{
		Id:      0,
		Name:    "macunaima",
		Cpf:     "12345678901",
		Address: Buyer_addresses[2],
	},
	{
		Id:      1,
		Name:    "joao da silva",
		Cpf:     "01234567890",
		Address: Buyer_addresses[2],
	},
	{
		Id:      2,
		Name:    "marie curie",
		Cpf:     "90123456789",
		Address: Buyer_addresses[0],
	},
}

func AddBuyerCompany(bc entity_public.Company) entity_public.Company {
	companies = append(companies, bc)
	return bc
}

func AddBuyerPersonal(bp entity_public.Personal) entity_public.Personal {
	personals = append(personals, bp)
	return bp
}

func GetBuyers() []entity_public.Buyer {
	var buyers []entity_public.Buyer
	for _, buyer := range companies {
		buyers = append(buyers, entity_public.Company{
			Id:   buyer.GetId(),
			FantasyName: buyer.GetName(),
		})
	}

	for _, buyer := range personals {
		buyers = append(buyers, entity_public.Personal{
			Id:   buyer.GetId(),
			Name: buyer.GetName(),
		})
	}

	return buyers
}
