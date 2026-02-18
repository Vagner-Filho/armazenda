package departure_service

import (
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
)

func GetDeparture(id uint32) (entity_public.Departure, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.GetDeparture(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a saída", "")
			return entity_public.Departure{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.Departure{}, &toast
	}
	return departure, nil
}

func GetDisplayDepartures(farm uint32, page int) ([]entity_public.DisplayDeparture, int, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, total, err := dModel.GetDisplayDepartures(farm, page)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a saída", "")
			return []entity_public.DisplayDeparture{}, 0, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return []entity_public.DisplayDeparture{}, 0, &toast
	}
	return departure, total, nil
}

func AddDeparture(bd entity_public.Departure) (entity_public.DisplayDeparture, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.AddDeparture(bd)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao adicionar a saída", "")
			return entity_public.DisplayDeparture{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.DisplayDeparture{}, &toast
	}
	toast := entity_public.GetSuccessToast("Saída cadastrada", "")
	return departure, &toast
}

func PutDeparture(d entity_public.Departure) (entity_public.DisplayDeparture, entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()
	departure, error := dModel.PutDeparture(d)
	if error != nil {
		if error.IsServerErr == true {
			return entity_public.DisplayDeparture{}, entity_public.GetErrorToast("Houve um erro interno ao editar a saída", "")
		}
		return entity_public.DisplayDeparture{}, entity_public.GetWarningToast(error.Message, "")
	}
	return departure, entity_public.GetSuccessToast("Saída editada", "")
}

func PutDepartureDraft(d entity_public.DepartureDraft) (entity_public.DisplayDepartureDraft, entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	departure, err := dModel.UpdateDepartureDraft(d)
	if err != nil {
		if err.IsServerErr == true {
			return entity_public.DisplayDepartureDraft{}, entity_public.GetErrorToast("Houve um erro interno ao editar o rascunho", "")
		}
		return entity_public.DisplayDepartureDraft{}, entity_public.GetWarningToast(err.Message, "")
	}
	return departure, entity_public.GetSuccessToast("Rascunho editado", "")
}

func DeleteDeparture(id uint32) *entity_public.Toast {
	dModel := departure_model.GetDepartureModel()
	err := dModel.DeleteDeparture(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Saída deletada", "")
	return &toast
}

func FilterDepartures(df entity_public.DepartureFilter, page int, farm uint32) ([]entity_public.DisplayDeparture, int, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()
	departures, total, err := dModel.FilterDepartures(df, page, farm)

	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar saídas", "")
		return departures, 0, &toast
	}

	return departures, total, nil
}

func GetDeparturePdf(id uint32) (*entity_public.DeparturePdf, *entity_public.Toast) {
	eModel := departure_model.GetDepartureModel()

	departure, err := eModel.GetDeparturePdf(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}

	return &departure, nil
}

func CreateDepartureDraft(d entity_public.DepartureDraft) (entity_public.DisplayDepartureDraft, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	draft, err := dModel.CreateDepartureDraft(d)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao criar o rascunho", "")
			return entity_public.DisplayDepartureDraft{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.DisplayDepartureDraft{}, &toast
	}
	toast := entity_public.GetSuccessToast("Rascunho de saída criado", "")
	return draft, &toast
}

func GetDepartureDraft(id uint32) (entity_public.DepartureDraft, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	draft, err := dModel.GetDepartureDraft(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar o rascunho", "")
			return entity_public.DepartureDraft{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.DepartureDraft{}, &toast
	}
	return draft, nil
}

func GetAllDepartureDrafts(farmId uint32) ([]entity_public.DisplayDepartureDraft, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	drafts, err := dModel.GetAllDepartureDrafts(farmId)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar os rascunhos", "")
			return []entity_public.DisplayDepartureDraft{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return []entity_public.DisplayDepartureDraft{}, &toast
	}
	return drafts, nil
}

func UpdateDepartureDraft(d entity_public.DepartureDraft) (entity_public.DisplayDepartureDraft, *entity_public.Toast) {
	dModel := departure_model.GetDepartureModel()

	draft, err := dModel.UpdateDepartureDraft(d)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao atualizar o rascunho", "")
			return entity_public.DisplayDepartureDraft{}, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return entity_public.DisplayDepartureDraft{}, &toast
	}
	toast := entity_public.GetSuccessToast("Rascunho de saída atualizado", "")
	return draft, &toast
}

func DeleteDepartureDraft(id uint32) *entity_public.Toast {
	dModel := departure_model.GetDepartureModel()
	err := dModel.DeleteDepartureDraft(id)

	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao deletar o rascunho", "")
			return &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return &toast
	}

	toast := entity_public.GetSuccessToast("Rascunho de saída deletado", "")
	return &toast
}

// SyncDeparture represents a departure for synchronization
type SyncDeparture struct {
	Id            uint32    `json:"id"`
	DepartureDate time.Time `json:"departureDate"`
	Vehicle       uint16    `json:"vehicle"`
	Crop          uint8     `json:"crop"`
	GrossWeight   float64   `json:"grossWeight"`
	Tare          float64   `json:"tare"`
	NetWeight     float64   `json:"netWeight"`
	Humidity      *float64  `json:"humidity,omitempty"`
	Damage        *float64  `json:"damage,omitempty"`
	Impurity      *float64  `json:"impurity,omitempty"`
	Farm          uint32    `json:"farm"`
	Recipient     *uint32   `json:"recipient,omitempty"`
	Origin        *uint32   `json:"origin,omitempty"`
	ModifiedAt    time.Time `json:"modifiedAt"`
	Deleted       bool      `json:"deleted,omitempty"`
}

// GetDeparturesForSync retrieves departures modified since a specific time
func GetDeparturesForSync(since time.Time, farm uint32) ([]SyncDeparture, error) {
	dModel := departure_model.GetDepartureModel()
	departures, err := dModel.GetDeparturesModifiedSince(since, farm)
	if err != nil {
		return nil, err
	}

	syncDepartures := make([]SyncDeparture, len(departures))
	for i, dep := range departures {
		syncDepartures[i] = convertToSyncDeparture(dep)
	}

	return syncDepartures, nil
}

// GetModifiedDepartureCount returns the count of departures modified since a specific time
func GetModifiedDepartureCount(since time.Time, farm uint32) (int, error) {
	dModel := departure_model.GetDepartureModel()
	return dModel.GetModifiedCount(since, farm)
}

func convertToSyncDeparture(dep entity_public.Departure) SyncDeparture {
	syncDep := SyncDeparture{
		Id:            dep.Id,
		DepartureDate: dep.DepartureDate,
		Vehicle:       dep.Vehicle,
		Crop:          dep.Crop,
		GrossWeight:   0,
		Tare:          0,
		NetWeight:     0,
		Farm:          dep.Farm,
		Recipient:     dep.Recipient,
		Origin:        dep.Origin,
		ModifiedAt:    dep.ModifiedAt,
	}

	// Convert CargoWeight
	gw, _ := dep.CargoWeight.GrossWeight.Float64()
	syncDep.GrossWeight = gw
	t, _ := dep.CargoWeight.Tare.Float64()
	syncDep.Tare = t
	nw, _ := dep.CargoWeight.NetWeight.Float64()
	syncDep.NetWeight = nw

	// Convert Analysis
	if dep.Analysis.Humidity != nil {
		h, _ := dep.Analysis.Humidity.Float64()
		syncDep.Humidity = &h
	}
	if dep.Analysis.Damage != nil {
		d, _ := dep.Analysis.Damage.Float64()
		syncDep.Damage = &d
	}
	if dep.Analysis.Impurity != nil {
		i, _ := dep.Analysis.Impurity.Float64()
		syncDep.Impurity = &i
	}

	return syncDep
}
