package person_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/person_model"
)

func GetPeopleByFarm(farm uint32) ([]entity_public.PersonOption, *entity_public.Toast) {
	bmodel := person_model.GetpersonModel()

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
	bmodel := person_model.GetpersonModel()

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
	bmodel := person_model.GetpersonModel()

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

func FilterPerson(filters entity_public.PersonFilter, farm uint32) ([]entity_public.PersonDisplay, *entity_public.Toast) {
	bModel := person_model.GetpersonModel()
	people, err := bModel.FilterPerson(filters, farm)
	if err != nil {
		toast := entity_public.GetErrorToast(err.Error(), "")
		return []entity_public.PersonDisplay{}, &toast
	}
	return people, nil
}
