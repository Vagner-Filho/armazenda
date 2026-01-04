package person_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/person_model"
	"math"
)

type PersonPageData struct {
	People      []entity_public.PersonDisplay
	CurrentPage int
	TotalPages  int
	NextPage    int
	PrevPage    int
	HasNextPage bool
	HasPrevPage bool
	NoContent   bool
}

func GetPeopleByFarm(farm uint32) ([]entity_public.PersonOption, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	people, err := bmodel.GetPeopleByFarm(farm)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Erro ao buscar pessoas", "")
			return []entity_public.PersonOption{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.PersonOption{}, &toast
	}
	return people, nil
}

func AddLegalPerson(bc entity_public.LegalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	exists, err := bmodel.CnpjExistsInFarm(bc.Cnpj, bc.Person.Farm)
	if err != nil {
		toast := entity_public.GetErrorToast(err.Message, "")
		return entity_public.PersonDisplay{}, &toast
	}
	if exists {
		toast := entity_public.GetWarningToast("CNPJ já cadastrado para esta fazenda.", "")
		return entity_public.PersonDisplay{}, &toast
	}

	person, err := bmodel.AddLegalPerson(bc)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.PersonDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.PersonDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Pessoa cadastrada!", "")
	return person, &toast
}

func AddNaturalPerson(bp entity_public.NaturalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	exists, err := bmodel.CpfExistsInFarm(bp.Cpf, bp.Person.Farm)
	if err != nil {
		toast := entity_public.GetErrorToast(err.Message, "")
		return entity_public.PersonDisplay{}, &toast
	}
	if exists {
		toast := entity_public.GetWarningToast("CPF já cadastrado para esta fazenda.", "")
		return entity_public.PersonDisplay{}, &toast
	}

	person, err := bmodel.AddNaturalPerson(bp)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.PersonDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.PersonDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Pessoa cadastrada!", "")
	return person, &toast
}

func FilterPerson(filters entity_public.PersonFilter, farm uint32, page, limit int) (PersonPageData, *entity_public.Toast) {
	bModel := person_model.GetPersonModel()
	people, total, err := bModel.FilterPerson(filters, farm, page, limit)
	if err != nil {
		if err.IsServerErr {
			toast := entity_public.GetErrorToast("Erro ao filtrar pessoas", err.Message)
			return PersonPageData{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return PersonPageData{}, &toast
	}

	if total == 0 {
		return PersonPageData{
			People:    []entity_public.PersonDisplay{},
			NoContent: true,
		}, nil
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return PersonPageData{
		People:      people,
		CurrentPage: page,
		TotalPages:  totalPages,
		NextPage:    page + 1,
		PrevPage:    page - 1,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
		NoContent:   page == 0 && len(people) == 0,
	}, nil
}

func GetNaturalPerson(id uint32) (entity_public.NaturalPerson, *entity_public.Toast) {
	pmodel := person_model.GetPersonModel()

	var person entity_public.NaturalPerson
	person, err := pmodel.GetNaturalPersonById(id)

	if err != nil {
		toast := entity_public.GetWarningToast(err.Message, "")
		return person, &toast
	}

	return person, nil
}

func GetLegalPerson(id uint32) (entity_public.LegalPerson, *entity_public.Toast) {
	pmodel := person_model.GetPersonModel()

	var person entity_public.LegalPerson
	person, err := pmodel.GetLegalPersonById(id)

	if err != nil {
		toast := entity_public.GetWarningToast(err.Message, "")
		return person, &toast
	}

	return person, nil
}

func UpdateNaturalPerson(bp entity_public.NaturalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	person, err := bmodel.UpdateNaturalPerson(bp)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.PersonDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.PersonDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Pessoa atualizada!", "")
	return person, &toast
}

func UpdateLegalPerson(bc entity_public.LegalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	person, err := bmodel.UpdateLegalPerson(bc)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return entity_public.PersonDisplay{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Error(), "")
		return entity_public.PersonDisplay{}, &toast
	}

	toast := entity_public.GetSuccessToast("Pessoa atualizada!", "")
	return person, &toast
}
