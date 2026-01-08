package entry_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/person_model"

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

func AddEntry(ge entity_public.Entry, em EntryModelInterface, pm PersonModelInterface, prod_m ProductModelInterface, cm CropModelInterface) (entity_public.DisplayEntry, entity_public.Toast) {
	damageThreshold := decimal.NewFromInt(8)
	impurityThreshold := decimal.NewFromInt(1)
	humidityThreshold := decimal.NewFromInt(14)
	base100 := decimal.NewFromInt(100)
	var totalDiscount decimal.Decimal

	rawNetWeight := ge.CargoWeight.GrossWeight.Sub(ge.Tare)
	if rawNetWeight.LessThan(decimal.Zero) {
		t := entity_public.GetWarningToast("O peso líquido não pode ser menor do que zero", "confira os dados")
		return entity_public.DisplayEntry{}, t
	}
	if ge.Damage != nil {
		exceedingDamage := ge.Damage.Sub(damageThreshold)
		if exceedingDamage.GreaterThan(decimal.Zero) {
			totalDiscount = totalDiscount.Add(rawNetWeight.Mul(exceedingDamage).Div(base100))
		}
	}
	if ge.Impurity != nil {
		exceedingImpurity := ge.Impurity.Sub(impurityThreshold)
		if exceedingImpurity.GreaterThan(decimal.Zero) {
			totalDiscount = totalDiscount.Add(rawNetWeight.Mul(exceedingImpurity).Div(base100))
		}
	}
	if ge.Humidity != nil {
		exceedingHumidty := ge.Humidity.Sub(humidityThreshold)
		if exceedingHumidty.GreaterThan(decimal.Zero) {
			discountModifier, humErr := pm.GetHumidityDiscount(ge.Origin, ge.Farm)
			if humErr != nil {
				toast := entity_public.GetWarningToast("Falha ao calcular desconto de humidade", "")
				return entity_public.DisplayEntry{}, toast
			} else {
				discount := exceedingHumidty.Mul(discountModifier)
				totalDiscount = totalDiscount.Add(rawNetWeight.Mul(discount).Div(base100))
				ge.HumidityDiscountModifier = &discountModifier
			}
		}
	}
	ge.NetWeight = rawNetWeight.Sub(totalDiscount)

	var storageTax decimal.Decimal
	var storageTaxModifier decimal.Decimal
	if ge.Origin != nil {
		crop, err := cm.GetCropById(ge.Crop)
		if err != nil {

		}

		product, err := prod_m.GetProductById(crop.Product)
		if err != nil {

		}

		personConfig, err := pm.GetPersonConfig(*ge.Origin)

		storageTaxModifier = personConfig.GetProductEntryDiscount(product.Id)
		weightAfterQualityDiscount := ge.NetWeight
		storageTax = weightAfterQualityDiscount.Mul(storageTaxModifier).Div(base100)

		ge.NetWeight = weightAfterQualityDiscount.Sub(storageTax)
	}

	newEntry, addErr := em.AddEntry(ge)
	if addErr != nil {
		if addErr.IsServerErr == true {
			return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Houve um erro interno ao adicionar a entrada", "")
		}
		return entity_public.DisplayEntry{}, entity_public.GetWarningToast(addErr.Message, "")
	}

	if ge.Origin != nil {
		err := em.AddEntryTax(newEntry.Id, storageTax, storageTaxModifier)
		if err != nil {

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

	discountedHumidity := DiscountHumidity(entry.Humidity, entry.GrossWeight.Sub(entry.Tare), entry.HumidityDiscountModifier)
	discountedDamage := DiscountDamage(entry.Damage, entry.GrossWeight.Sub(entry.Tare))
	discountedImpurity := DiscountImpurity(entry.Impurity, entry.GrossWeight.Sub(entry.Tare))

	entryPdf.DiscountedHumidity = discountedHumidity
	entryPdf.DiscountedDamage = discountedDamage
	entryPdf.DiscountedImpurity = discountedImpurity

	return &entryPdf, nil
}

func PutEntry(ge entity_public.Entry) (entity_public.DisplayEntry, entity_public.Toast) {
	eModel := entry_model.GetEntryModel()

	damageThreshold := decimal.NewFromInt(8)
	impurityThreshold := decimal.NewFromInt(1)
	humidityThreshold := decimal.NewFromInt(14)
	base100 := decimal.NewFromInt(100)
	var totalDiscount decimal.Decimal

	rawNetWeight := ge.CargoWeight.GrossWeight.Sub(ge.Tare)
	if rawNetWeight.LessThan(decimal.Zero) {
		t := entity_public.GetWarningToast("O peso líquido não pode ser menor do que zero", "confira o peso bruto e a tara")
		return entity_public.DisplayEntry{}, t
	}
	if ge.Damage != nil {
		exceedingDamage := ge.Damage.Sub(damageThreshold)
		if exceedingDamage.GreaterThan(decimal.Zero) {
			totalDiscount = totalDiscount.Add(rawNetWeight.Mul(exceedingDamage).Div(base100))
		}
	}
	if ge.Impurity != nil {
		exceedingImpurity := ge.Impurity.Sub(impurityThreshold)
		if exceedingImpurity.GreaterThan(decimal.Zero) {
			totalDiscount = totalDiscount.Add(rawNetWeight.Mul(exceedingImpurity).Div(base100))
		}
	}
	if ge.Humidity != nil {
		exceedingHumidty := ge.Humidity.Sub(humidityThreshold)
		if exceedingHumidty.GreaterThan(decimal.Zero) {
			pm := person_model.GetPersonModel()
			discountModifier, humErr := pm.GetHumidityDiscount(ge.Origin, ge.Farm)
			if humErr != nil {
				toast := entity_public.GetWarningToast("Falha ao calcular desconto de humidade", "")
				return entity_public.DisplayEntry{}, toast
			} else {
				discount := exceedingHumidty.Mul(discountModifier)
				totalDiscount = totalDiscount.Add(rawNetWeight.Mul(discount).Div(base100))
			}
		}
	}
	ge.NetWeight = rawNetWeight.Sub(totalDiscount)

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
