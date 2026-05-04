package entry_service

import (
	"fmt"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/model/entry_model"
	model_error "armazenda/model/error"
	"armazenda/model/farm_config_model"
	"armazenda/model/humidity_progression_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/pkg/calculator"

	"github.com/shopspring/decimal"
)

func AddEntryDraft(ge entity_public.EntryDraft) (entity_public.DisplayEntryDraft, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	newEntry, addErr := eModel.AddEntryDraft(ge)
	if addErr != nil {
		if addErr.IsServerErr == true {
			return entity_public.DisplayEntryDraft{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar o rascunho", "")
		}
		return entity_public.DisplayEntryDraft{}, entity_public.GetWarningToast(addErr.Message, "")
	}
	return newEntry, entity_public.GetSuccessToast("Rascunho adicionado", "")
}

func GetEntryDraft(id uint32) (entity_public.EntryDraft, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	newEntry, getErr := eModel.GetEntryDraft(id)
	if getErr != nil {
		if getErr.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao recuperar o rascunho", "")
			return entity_public.EntryDraft{}, &toast
		}
		toast := entity_public.GetWarningToast(getErr.Message, "")
		return entity_public.EntryDraft{}, &toast
	}
	return newEntry, nil
}

func getStorageTaxModifier(e entity_public.Entry, cm CropModelInterface, prod_m ProductModelInterface, pm PersonModelInterface) (decimal.Decimal, error) {
	var storageTaxModifier decimal.Decimal
	if e.Origin != nil {
		crop, err := cm.GetCropById(e.Crop)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(storage tax) GetCropById: %v", err.Error()))
			return storageTaxModifier, err
		}

		product, err := prod_m.GetProductById(crop.Product)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(storage tax) GetProductById: %v", err.Error()))
			return storageTaxModifier, err
		}

		personConfig, err := pm.GetPersonConfig(*e.Origin)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(storage tax) GetPersonConfig: %v", err.Error()))
			return storageTaxModifier, err
		}

		storageTaxModifier = personConfig.GetProductEntryDiscount(product.Id)
	}
	return storageTaxModifier, nil
}

func getProgressionId(e entity_public.Entry, pm PersonModelInterface, fcm FarmConfigModelInterface) (*uint32, error) {
	var progressionId *uint32
	if e.Origin != nil {
		config, err := pm.GetPersonConfig(*e.Origin)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(progression id) GetPersonConfig: %v", err.Error()))
			return progressionId, err
		}
		progressionId = config.HumidityProgressionId
	} else {
		fc, err := fcm.GetFarmConfig(e.Farm)
		if err == nil {
			progressionId = fc.FarmUsedHumidityProgressionId
		} else {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(progression id) GetFarmConfig: %v", err.Error()))
		}
	}
	return progressionId, nil
}

func getEntryCalculationInput(e entity_public.Entry, pm PersonModelInterface, fcm FarmConfigModelInterface, hpm HumidityProgressionModelInterface, cm CropModelInterface, prod_m ProductModelInterface) (calculator.EntryCalculationInput, *entity_public.Toast) {
	var progressionId *uint32
	var storageTaxModifier decimal.Decimal
	var discountModifier decimal.Decimal
	var threshold decimal.Decimal
	var err error

	if e.Humidity != nil && e.Humidity.GreaterThan(decimal.Zero) {
		progressionId, err = getProgressionId(e, pm, fcm)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(entry calculation input) progressionId: %v", err.Error()))
			toast := entity_public.GetErrorToast("Falha ao calcular desconto de humidade", "entrada não adicionada")
			return calculator.EntryCalculationInput{}, &toast
		}
		threshold, err = hpm.GetFirstTierThreshold(progressionId)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(entry calculation input) threshold: %v", err.Error()))
			toast := entity_public.GetErrorToast("Falha ao calcular desconto de humidade", "entrada não adicionada")
			return calculator.EntryCalculationInput{}, &toast
		}

		if e.Humidity.GreaterThan(threshold) {
			discountModifier, err = pm.GetHumidityDiscount(e.Origin, e.Farm, *e.Humidity)
			if err != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("(entry calculation input) discountModifier: %v", err.Error()))
				toast := entity_public.GetErrorToast("Falha ao calcular desconto de humidade", "entrada não adicionada")
				return calculator.EntryCalculationInput{}, &toast
			}
		}

		e.HumidityDiscountModifier = &discountModifier
	}

	if e.Origin != nil {
		storageTaxModifier, err = getStorageTaxModifier(e, cm, prod_m, pm)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(entry calculation input) storageTaxModifier: %v", err.Error()))
			toast := entity_public.GetErrorToast("Falha ao calcular taxa de serviço", "entrada não adicionada")
			return calculator.EntryCalculationInput{}, &toast
		}
	}

	return calculator.EntryCalculationInput{
		GrossWeight:        e.GrossWeight,
		Tare:               e.Tare,
		Humidity:           e.Humidity,
		Damage:             e.Damage,
		Impurity:           e.Impurity,
		HumidityModifier:   &discountModifier,
		StorageTaxModifier: &storageTaxModifier,
		HumidityThreshold:  &threshold,
	}, nil
}

func AddEntry(ge entity_public.Entry, em EntryModelInterface, pm PersonModelInterface, prod_m ProductModelInterface, cm CropModelInterface, hpm HumidityProgressionModelInterface, fcm FarmConfigModelInterface) (entity_public.DisplayEntry, entity_public.Toast) {
	calcInput, toast := getEntryCalculationInput(ge, pm, fcm, hpm, cm, prod_m)

	if toast != nil {
		return entity_public.DisplayEntry{}, *toast
	}

	result := calculator.CalculateEntry(calcInput)

	if result.IsValid == false {
		return entity_public.DisplayEntry{}, entity_public.GetErrorToast(result.ErrorMessage, "")
	}

	ge.NetWeight = result.NetWeight
	newEntry, addErr := em.AddEntry(ge)

	if addErr != nil {
		if addErr.IsServerErr == true {
			return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.DisplayEntry{}, entity_public.GetWarningToast(addErr.Message, "")
	}

	if ge.Origin != nil {
		err := em.AddEntryTax(newEntry.Id, result.StorageTax, *calcInput.StorageTaxModifier)
		if err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("(entry tax) error from AddEntryTax: %v", err.Error()))
		}
	}
	return newEntry, entity_public.GetSuccessToast("Entrada adicionada", "")
}

func GetEntry(id uint32) (*entity_public.Entry, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, err := eModel.GetEntry(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}
	return &entry, nil
}

func GetEntryPdf(id uint32) (*entity_public.EntryPdf, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entryPdf, err := eModel.GetEntryPdf(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}

	entry, err := eModel.GetEntry(id)
	if err != nil {
		if err.IsServerErr == true {
			toast := entity_public.GetErrorToast("Houve um erro interno ao buscar a entrada :(", "")
			return nil, &toast
		}
		toast := entity_public.GetWarningToast(err.Message, "")
		return nil, &toast
	}

	// Fetch the humidity threshold from the progression
	hpm := humidity_progression_model.GetHumidityProgressionModel()
	pm := person_model.GetPersonModel()
	var progressionId *uint32
	if entry.Origin != nil {
		config, _ := pm.GetPersonConfig(*entry.Origin)
		progressionId = config.HumidityProgressionId
	}
	threshold, _ := hpm.GetFirstTierThreshold(progressionId)

	discountedHumidity := DiscountHumidity(entry.Humidity, entry.GrossWeight.Sub(entry.Tare), entry.HumidityDiscountModifier, &threshold)
	discountedDamage := DiscountDamage(entry.Damage, entry.GrossWeight.Sub(entry.Tare))
	discountedImpurity := DiscountImpurity(entry.Impurity, entry.GrossWeight.Sub(entry.Tare))

	entryPdf.DiscountedHumidity = discountedHumidity
	entryPdf.DiscountedDamage = discountedDamage
	entryPdf.DiscountedImpurity = discountedImpurity

	return &entryPdf, nil
}

func PutEntry(ge entity_public.Entry) (entity_public.DisplayEntry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()
	pm := person_model.GetPersonModel()
	fcm := farm_config_model.GetFarmConfigModel()
	hpm := humidity_progression_model.GetHumidityProgressionModel()
	cm := crop_model.GetCropModel()
	prod_m := product_model.GetProductModel()

	calcInput, toast := getEntryCalculationInput(ge, pm, fcm, hpm, cm, prod_m)

	if toast != nil {
		return entity_public.DisplayEntry{}, *toast
	}

	result := calculator.CalculateEntry(calcInput)

	if result.IsValid == false {
		return entity_public.DisplayEntry{}, entity_public.GetErrorToast(result.ErrorMessage, "")
	}

	ge.NetWeight = result.NetWeight
	entry, putErr := eModel.PutEntry(ge)
	if putErr != nil {
		if putErr.IsServerErr == true {
			return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Houve um erro interno ao editar a entrada", "")
		}
		return entity_public.DisplayEntry{}, entity_public.GetWarningToast(putErr.Message, "")
	}
	return entry, entity_public.GetSuccessToast("Entrada editada", "")
}

func PutEntryDraft(ge entity_public.EntryDraft) (entity_public.DisplayEntryDraft, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	entry, putErr := eModel.PutEntryDraft(ge)
	if putErr != nil {
		if putErr.IsServerErr == true {
			return entity_public.DisplayEntryDraft{}, entity_public.GetErrorToast("Houve um erro interno ao editar o rascunho", "")
		}
		return entity_public.DisplayEntryDraft{}, entity_public.GetWarningToast(putErr.Message, "")
	}
	return entry, entity_public.GetSuccessToast("Rascunho editado", "")
}

func DeleteEntry(id uint32) *entity_public.Toast {
	dModel := entry_model.GetEntryModel()
	err := dModel.DeleteEntry(id)

	if err != nil {

	}

	toast := entity_public.GetSuccessToast("Entrada deletada", "")
	return &toast
}

func DeleteEntryDraft(id uint32) *entity_public.Toast {
	dModel := entry_model.GetEntryModel()
	err := dModel.DeleteEntryDraft(id)

	if err != nil {
		toast := entity_public.GetErrorToast("Houve um erro ao deletar o rascunho", "")
		return &toast
	}

	toast := entity_public.GetSuccessToast("Rascunho deletado", "")
	return &toast
}

func FilterEntries(ef entity_public.EntryFilter, page int, farm uint32) ([]entity_public.DisplayEntry, int, *entity_public.Toast) {
	eModel := entry_model.GetEntryModel()
	entries, total, err := eModel.FilterEntries(ef, page, farm)

	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar entradas", "")
		return entries, 0, &toast
	}

	return entries, total, nil
}

// SyncEntry represents an entry for synchronization
type SyncEntry struct {
	Id                       uint32    `json:"id"`
	Field                    uint16    `json:"field"`
	Crop                     uint8     `json:"crop"`
	Vehicle                  uint16    `json:"vehicle"`
	GrossWeight              float64   `json:"grossWeight"`
	Tare                     float64   `json:"tare"`
	NetWeight                float64   `json:"netWeight"`
	Humidity                 *float64  `json:"humidity,omitempty"`
	Damage                   *float64  `json:"damage,omitempty"`
	Impurity                 *float64  `json:"impurity,omitempty"`
	HumidityDiscountModifier *float64  `json:"humidityDiscountModifier,omitempty"`
	ArrivalDate              time.Time `json:"arrivalDate"`
	Farm                     uint32    `json:"farm"`
	Origin                   *uint32   `json:"origin,omitempty"`
	ModifiedAt               time.Time `json:"modifiedAt"`
	Deleted                  bool      `json:"deleted,omitempty"`
}

// GetEntriesForSync retrieves entries modified since a specific time
func GetEntriesForSync(since time.Time, farm uint32) ([]SyncEntry, error) {
	eModel := entry_model.GetEntryModel()
	entries, err := eModel.GetEntriesModifiedSince(since, farm)
	if err != nil {
		return nil, err
	}

	syncEntries := make([]SyncEntry, len(entries))
	for i, entry := range entries {
		syncEntries[i] = convertToSyncEntry(entry)
	}

	return syncEntries, nil
}

// GetModifiedEntryCount returns the count of entries modified since a specific time
func GetModifiedEntryCount(since time.Time, farm uint32) (int, error) {
	eModel := entry_model.GetEntryModel()
	return eModel.GetModifiedCount(since, farm)
}

func convertToSyncEntry(entry entity_public.Entry) SyncEntry {
	syncEntry := SyncEntry{
		Id:          entry.Id,
		Field:       entry.Field,
		Crop:        entry.Crop,
		Vehicle:     entry.Vehicle,
		GrossWeight: 0, // Will be set from CargoWeight
		Tare:        0,
		NetWeight:   0,
		ArrivalDate: entry.ArrivalDate,
		Farm:        entry.Farm,
		Origin:      entry.Origin,
		ModifiedAt:  entry.ModifiedAt,
	}

	// Convert CargoWeight
	gw, _ := entry.CargoWeight.GrossWeight.Float64()
	syncEntry.GrossWeight = gw
	t, _ := entry.CargoWeight.Tare.Float64()
	syncEntry.Tare = t
	nw, _ := entry.CargoWeight.NetWeight.Float64()
	syncEntry.NetWeight = nw

	// Convert Analysis
	if entry.Analysis.Humidity != nil {
		h, _ := entry.Analysis.Humidity.Float64()
		syncEntry.Humidity = &h
	}
	if entry.Analysis.Damage != nil {
		d, _ := entry.Analysis.Damage.Float64()
		syncEntry.Damage = &d
	}
	if entry.Analysis.Impurity != nil {
		i, _ := entry.Analysis.Impurity.Float64()
		syncEntry.Impurity = &i
	}

	return syncEntry
}
