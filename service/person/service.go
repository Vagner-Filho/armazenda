package person_service

import (
	"time"

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

func AddLegalPerson(lp entity_public.LegalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	var t entity_public.Toast
	if len(lp.Cnpj) < 14 {
		t = entity_public.GetWarningToast("CNPJ inválido", "verifique tamanho e formato")
		return entity_public.PersonDisplay{}, &t
	}
	exists, err := bmodel.CnpjExistsInFarm(lp.Cnpj, lp.Person.Farm)
	if err != nil {
		t = entity_public.GetErrorToast(err.Message, "")
		return entity_public.PersonDisplay{}, &t
	}
	if exists {
		t = entity_public.GetWarningToast("CNPJ já cadastrado para esta fazenda.", "")
		return entity_public.PersonDisplay{}, &t
	}

	person, err := bmodel.AddLegalPerson(lp)
	if err != nil {
		if err.IsServerErr == true {
			t = entity_public.GetErrorToast(err.Error(), "")
			return entity_public.PersonDisplay{}, &t
		}
		t = entity_public.GetWarningToast(err.Error(), "")
		return entity_public.PersonDisplay{}, &t
	}

	t = entity_public.GetSuccessToast("Pessoa cadastrada!", "")
	return person, &t
}

func AddNaturalPerson(np entity_public.NaturalPerson) (entity_public.PersonDisplay, *entity_public.Toast) {
	bmodel := person_model.GetPersonModel()

	var t entity_public.Toast
	if len(np.Cpf) < 11 {
		t = entity_public.GetWarningToast("CPF inválido", "verifique tamanho e formato")
		return entity_public.PersonDisplay{}, &t
	}
	exists, err := bmodel.CpfExistsInFarm(np.Cpf, np.Person.Farm)
	if err != nil {
		toast := entity_public.GetErrorToast(err.Message, "")
		return entity_public.PersonDisplay{}, &toast
	}
	if exists {
		toast := entity_public.GetWarningToast("CPF já cadastrado para esta fazenda.", "")
		return entity_public.PersonDisplay{}, &toast
	}

	person, err := bmodel.AddNaturalPerson(np)
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

// SyncPerson represents a person for synchronization
type SyncPerson struct {
	Id                    uint32  `json:"id"`
	Name                  string  `json:"name"`
	Document              string  `json:"document"`
	IE                    string  `json:"ie"`
	Farm                  uint32  `json:"farm"`
	HumidityProgressionId *uint32 `json:"humidityProgressionId"`
	EntrySoyDiscount      float64 `json:"entrySoyDiscount"`
	EntryCornDiscount     float64 `json:"entryCornDiscount"`
	ModifiedAt            int64   `json:"modifiedAt"`
	Deleted               bool    `json:"deleted,omitempty"`
}

// GetPeopleForSync retrieves people modified since a specific time
func GetPeopleForSync(since time.Time, farm uint32) ([]SyncPerson, error) {
	pmodel := person_model.GetPersonModel()
	people, err := pmodel.GetPeopleModifiedSince(since, farm)
	if err != nil {
		return nil, err
	}

	syncPeople := make([]SyncPerson, len(people))
	for i, person := range people {
		syncPeople[i] = convertToSyncPerson(person)
	}

	return syncPeople, nil
}

// GetModifiedPersonCount returns the count of people modified since a specific time
func GetModifiedPersonCount(since time.Time, farm uint32) (int, error) {
	pmodel := person_model.GetPersonModel()
	return pmodel.GetModifiedCount(since, farm)
}

func convertToSyncPerson(person entity_public.Person) SyncPerson {
	syncPerson := SyncPerson{
		Id:                    person.Id,
		IE:                    person.Ie,
		Farm:                  person.Farm,
		HumidityProgressionId: person.PersonConfig.HumidityProgressionId,
		EntrySoyDiscount:      0,
		EntryCornDiscount:     0,
		ModifiedAt:            person.ModifiedAt.Unix(),
	}

	// Convert PersonConfig
	esd, _ := person.PersonConfig.EntrySoyDiscount.Float64()
	syncPerson.EntrySoyDiscount = esd
	ecd, _ := person.PersonConfig.EntryCornDiscount.Float64()
	syncPerson.EntryCornDiscount = ecd

	return syncPerson
}
