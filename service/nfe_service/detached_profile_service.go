package nfe_service

import (
	"errors"
	"fmt"
	"strings"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// DetachedProfileItemInput is one item of a Rascunho de NF-e: the product
// fields collected by the detached form plus the per-item CFOP and unit price.
// Quantity and gross weight are not rascunho values.
type DetachedProfileItemInput struct {
	FarmProductID *uint16
	ProductName   string
	NCM           string
	CEST          *string
	CFOP          string
	Unit          string
	UnitPrice     decimal.Decimal
}

// DetachedProfileInput carries every field needed to create or update a
// Rascunho de NF-e. Recipient is optional; when PersonID is set the recipient
// must belong to the farm, and when inline data is provided the save action
// creates a farm contact and links the rascunho to it.
type DetachedProfileInput struct {
	ID         int // 0 = create, >0 = update
	FarmID     uint32
	Name       string
	Recipient  DetachedRecipient
	NaturezaOp *string
	ModFrete   *int
	InfCpl     *string
	Items      []DetachedProfileItemInput
	TaxRates   entity.TaxRates
	Overrides  *entity.InvoiceOverrides
}

// profileRecipientID resolves the optional recipient of a rascunho. An
// explicit person id is validated against the current farm. Inline recipient
// data is only saved as a contact when explicitly requested by the save
// action (allowCreate), never during a read/apply. Returns nil when no
// recipient was provided at all.
func (s *NFeService) profileRecipientID(input DetachedProfileInput, nfeModel *nfe_model.NFeModel, allowCreate bool) (*uint32, entity_public.Toast) {
	if input.Recipient.PersonID != nil {
		recipientID := *input.Recipient.PersonID
		if err := nfeModel.ValidateDetachedProfileReferences(input.FarmID, &recipientID, nil); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("profileRecipientID validation error: %v", err))
			return nil, entity_public.GetWarningToast("Destinatário inválido", "O destinatário selecionado não pertence à sua fazenda")
		}
		return &recipientID, entity_public.Toast{}
	}

	name := strings.TrimSpace(input.Recipient.Name)
	document := strings.TrimSpace(input.Recipient.Document)
	if name == "" && document == "" {
		// No recipient is a valid rascunho state.
		return nil, entity_public.Toast{}
	}
	if name == "" || document == "" {
		return nil, entity_public.GetWarningToast(
			"Destinatário incompleto",
			"Informe o nome e o documento do destinatário")
	}
	if !allowCreate {
		return nil, entity_public.GetWarningToast(
			"Dados do destinatário não salvos",
			"Salve o destinatário como contato da fazenda para usá-lo no rascunho")
	}

	pModel := person_model.GetPersonModel()
	recipientInput := person_model.DetachedRecipientInput{
		Name:         name,
		Document:     document,
		IE:           input.Recipient.IE,
		Street:       input.Recipient.Street,
		Number:       input.Recipient.Number,
		Neighborhood: input.Recipient.Neighborhood,
		City:         input.Recipient.City,
		State:        input.Recipient.State,
		CEP:          input.Recipient.CEP,
		PhoneNumber:  input.Recipient.PhoneNumber,
		Email:        input.Recipient.Email,
	}
	personID, createErr := pModel.CreatePersonForDetachedNFe(input.FarmID, detachedDocumentType(document), recipientInput)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreatePersonForDetachedNFe (profile) error: %v", createErr.Error()))
		return nil, entity_public.GetErrorToast("Falha ao salvar destinatário", "")
	}
	return &personID, entity_public.Toast{}
}

// validateProfileItems normalizes and validates the item defaults of a
// rascunho: at least one item, four-digit CFOPs and non-negative exact unit
// prices.
func validateProfileItems(items []DetachedProfileItemInput) (*entity_public.Toast, []nfe_model.DetachedProfileItem) {
	if len(items) == 0 {
		toast := entity_public.GetWarningToast("Nenhum item informado", "Adicione pelo menos um item ao rascunho")
		return &toast, nil
	}

	normalized := make([]nfe_model.DetachedProfileItem, 0, len(items))
	for i, item := range items {
		if !IsFourDigitCFOP(item.CFOP) {
			toast := entity_public.GetWarningToast(
				"CFOP inválido",
				fmt.Sprintf("Item %d: informe um CFOP com exatamente 4 dígitos", i+1))
			return &toast, nil
		}
		if item.UnitPrice.IsNegative() {
			toast := entity_public.GetWarningToast(
				"Preço unitário inválido",
				fmt.Sprintf("Item %d: informe um preço unitário válido", i+1))
			return &toast, nil
		}
		unit := strings.TrimSpace(item.Unit)
		if unit == "" {
			unit = "KG"
		}
		normalized = append(normalized, nfe_model.DetachedProfileItem{
			FarmProductID: item.FarmProductID,
			ProductName:   strings.TrimSpace(item.ProductName),
			NCM:           strings.TrimSpace(item.NCM),
			CEST:          item.CEST,
			CFOP:          item.CFOP,
			Unit:          unit,
			UnitPrice:     item.UnitPrice,
		})
	}
	return nil, normalized
}

// SaveDetachedProfile creates (ID == 0) or updates a farm-scoped Rascunho de
// NF-e. It never allocates an invoice number, signs XML, persists an invoice
// or calls SEFAZ; the only write beyond the rascunho itself is the explicitly
// requested creation of an inline recipient contact.
func (s *NFeService) SaveDetachedProfile(input DetachedProfileInput) (*nfe_model.DetachedProfile, entity_public.Toast) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, entity_public.GetWarningToast("Nome obrigatório", "Informe um nome para o rascunho")
	}

	nfeModel := nfe_model.GetNFeModel()

	recipientID, toast := s.profileRecipientID(input, nfeModel, true)
	if isFailureProfileToast(toast) {
		return nil, toast
	}

	itemsToast, items := validateProfileItems(input.Items)
	if itemsToast != nil {
		return nil, *itemsToast
	}

	if err := nfeModel.ValidateDetachedProfileReferences(input.FarmID, recipientID, nfe_model.FarmProductIDsFromProfileItems(items)); err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("SaveDetachedProfile references error: %v", err))
		return nil, entity_public.GetWarningToast(
			"Item ou destinatário inválido",
			"Um dos itens ou o destinatário não pertence à sua fazenda")
	}

	profile := nfe_model.DetachedProfile{
		ID:          input.ID,
		FarmID:      input.FarmID,
		Name:        name,
		RecipientID: recipientID,
		NaturezaOp:  input.NaturezaOp,
		ModFrete:    input.ModFrete,
		InfCpl:      input.InfCpl,
		Items:       items,
		TaxRates:    &input.TaxRates,
	}
	if input.Overrides != nil {
		profile.ICMSCST = input.Overrides.ICMSCST
		profile.PISCST = input.Overrides.PISCST
		profile.COFINSCST = input.Overrides.COFINSCST
		profile.IBSCST = input.Overrides.IBSCST
		profile.CBSCST = input.Overrides.CBSCST
		profile.CClassTrib = input.Overrides.CClassTrib
	}

	if input.ID > 0 {
		if err := nfeModel.UpdateDetachedProfile(profile); err != nil {
			if errors.Is(err, nfe_model.ErrDetachedProfileNameTaken) {
				return nil, entity_public.GetWarningToast("Nome de rascunho já existe", "Escolha outro nome para o rascunho")
			}
			if errors.Is(err, nfe_model.ErrDetachedProfileNotFound) {
				return nil, entity_public.GetWarningToast("Rascunho não encontrado", "")
			}
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedProfile error: %v", err))
			return nil, entity_public.GetErrorToast("Falha ao salvar rascunho", "")
		}
		saved, getToast := s.GetDetachedProfile(input.FarmID, input.ID)
		if isFailureProfileToast(getToast) {
			return nil, getToast
		}
		return saved, entity_public.GetSuccessToast("Rascunho atualizado", "")
	}

	id, err := nfeModel.CreateDetachedProfile(profile)
	if err != nil {
		if errors.Is(err, nfe_model.ErrDetachedProfileNameTaken) {
			return nil, entity_public.GetWarningToast("Nome de rascunho já existe", "Escolha outro nome para o rascunho")
		}
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateDetachedProfile error: %v", err))
		return nil, entity_public.GetErrorToast("Falha ao salvar rascunho", "")
	}

	saved, getToast := s.GetDetachedProfile(input.FarmID, id)
	if isFailureProfileToast(getToast) {
		return nil, getToast
	}
	return saved, entity_public.GetSuccessToast("Rascunho salvo", "")
}

// GetDetachedProfile loads a single farm-scoped rascunho. A rascunho from
// another farm is reported as not found.
func (s *NFeService) GetDetachedProfile(farmID uint32, id int) (*nfe_model.DetachedProfile, entity_public.Toast) {
	profile, err := nfe_model.GetNFeModel().GetDetachedProfile(farmID, id)
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetDetachedProfile error: %v", err))
		return nil, entity_public.GetErrorToast("Erro interno ao buscar o rascunho", "")
	}
	if profile == nil {
		return nil, entity_public.GetWarningToast("Rascunho não encontrado", "")
	}
	return profile, entity_public.Toast{}
}

// ListDetachedProfiles returns every rascunho of the farm.
func (s *NFeService) ListDetachedProfiles(farmID uint32) ([]nfe_model.DetachedProfile, entity_public.Toast) {
	profiles, err := nfe_model.GetNFeModel().GetDetachedProfilesByFarm(farmID)
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetDetachedProfilesByFarm error: %v", err))
		return nil, entity_public.GetErrorToast("Erro interno ao buscar os rascunhos", "")
	}
	return profiles, entity_public.Toast{}
}

// DeleteDetachedProfile hard-deletes a farm-scoped rascunho. It has no other
// side effects: issued invoices already carry their own snapshots.
func (s *NFeService) DeleteDetachedProfile(farmID uint32, id int) entity_public.Toast {
	deleted, err := nfe_model.GetNFeModel().DeleteDetachedProfile(farmID, id)
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("DeleteDetachedProfile error: %v", err))
		return entity_public.GetErrorToast("Falha ao excluir o rascunho", "")
	}
	if !deleted {
		return entity_public.GetWarningToast("Rascunho não encontrado", "")
	}
	return entity_public.GetSuccessToast("Rascunho excluído", "")
}

// isFailureProfileToast reports whether a toast represents a failure that must
// stop a rascunho operation.
func isFailureProfileToast(toast entity_public.Toast) bool {
	return toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast
}
