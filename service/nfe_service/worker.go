package nfe_service

import (
	"fmt"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	model_error "armazenda/model/error"
	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"

	"github.com/shopspring/decimal"
)

// StartRetryWorker starts the background worker that retries pending invoices.
// On startup it immediately resets backoff for all pending invoices and runs
// a first pass, then continues with a periodic ticker every 5 minutes.
func StartRetryWorker() {
	go func() {
		nfeModel := nfe_model.GetNFeModel()

		// On startup: reset backoff so all pending invoices are eligible immediately.
		if err := nfeModel.ResetPendingBackoff(); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker startup: failed to reset backoff: %v", err))
		}

		// Run immediately on startup — don't wait for first ticker tick.
		processPendingInvoices()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			processPendingInvoices()
		}
	}()
}

func processPendingInvoices() {
	nfeModel := nfe_model.GetNFeModel()

	// Process pending invoices (status polling for already-sent NF-e)
	pendingInvoices, err := nfeModel.GetPendingInvoicesForRetry()
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to get pending invoices: %v", err))
		return
	}
	for _, inv := range pendingInvoices {
		if err := processInvoice(inv); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to process pending invoice %d: %v", inv.ID, err))
		}
	}

	// Process draft invoices (auto-retry via SVC when available, max 24h old)
	draftInvoices, err := nfeModel.GetDraftInvoicesForRetry()
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to get draft invoices: %v", err))
		return
	}
	for _, inv := range draftInvoices {
		if err := processDraftInvoice(inv); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to process draft invoice %d: %v", inv.ID, err))
		}
	}
}

// isProcessingStatusCode returns true for SEFAZ codes that indicate the NF-e is still being processed.
func isProcessingStatusCode(code string) bool {
	switch code {
	case "101", "102", "103", "104", "105", "106":
		return true
	}
	return false
}

func processInvoice(inv nfe_model.InvoiceForRetry) error {
	nfeModel := nfe_model.GetNFeModel()

	// Get departure to find farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(inv.DepartureID)
	if depErr != nil {
		return fmt.Errorf("failed to get departure: %w", depErr)
	}

	// Get farm NFe config
	farmConfig, cfgErr := nfeModel.GetFarmConfig(departure.Farm)
	if cfgErr != nil {
		return fmt.Errorf("failed to get farm config: %w", cfgErr)
	}
	if farmConfig == nil {
		return fmt.Errorf("farm has no NFe config")
	}

	certPassword := decryptPassword(farmConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmConfig.Environment),
		StateUF:     farmConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(sefazCfg)

	// Determine which endpoint to query based on the invoice's tpEmis
	tpEmis := defaults.TpEmis(inv.TpEmis)
	if tpEmis == defaults.EmissaoNormal {
		tpEmis = defaults.EmissaoNormal
	}

	resp, queryErr := invService.QueryInvoiceStatusWithEmission(inv.AccessKey, farmConfig.CertificateData, certPassword, tpEmis)
	if queryErr != nil {
		// Network error — keep pending, do NOT increment retry count
		return queryErr
	}

	if resp == nil {
		return fmt.Errorf("empty query response from SEFAZ")
	}

	// We got a response from SEFAZ — now count this as a real retry attempt
	if err := nfeModel.IncrementRetryCount(inv.ID); err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to increment retry count for invoice %d: %v", inv.ID, err))
	}

	if resp.IsAuthorized() {
		updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "authorized", resp.Protocol, resp.StatusCode, resp.StatusMotive)
		if updErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
		}
		// Build and store the <nfeProc> wrapper so the DANFE parser can
		// extract nProt/dhRecbto from the stored XML.
		if inv.XMLSigned != "" {
			if authXML, buildErr := nfe_xml.BuildAuthorizedXML(inv.XMLSigned, inv.AccessKey, resp.Protocol, resp.DhRecbto, resp.StatusCode, resp.StatusMotive); buildErr == nil {
				if xmlErr := nfeModel.UpdateInvoiceAuthorizedXML(inv.ID, authXML); xmlErr != nil {
					model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: UpdateInvoiceAuthorizedXML error: %v", xmlErr.Error()))
				}
			} else {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: BuildAuthorizedXML error: %v", buildErr.Error()))
			}
		}
	} else if isProcessingStatusCode(resp.StatusCode) {
		// Still processing — keep pending
		updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "pending", resp.Protocol, resp.StatusCode, resp.StatusMotive)
		if updErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
		}
	} else {
		// Definitive non-success response (rejection, not found, etc.) — mark as denied
		updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "denied", resp.Protocol, resp.StatusCode, resp.StatusMotive)
		if updErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
		}
	}

	return nil
}

func processDraftInvoice(inv nfe_model.InvoiceForRetry) error {
	nfeModel := nfe_model.GetNFeModel()

	// Get departure to find farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(inv.DepartureID)
	if depErr != nil {
		return fmt.Errorf("failed to get departure: %w", depErr)
	}

	// Get farm NFe config
	farmConfig, cfgErr := nfeModel.GetFarmConfig(departure.Farm)
	if cfgErr != nil {
		return fmt.Errorf("failed to get farm config: %w", cfgErr)
	}
	if farmConfig == nil {
		return fmt.Errorf("farm has no NFe config")
	}

	certPassword := decryptPassword(farmConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmConfig.Environment),
		StateUF:     farmConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(sefazCfg)

	// Check if SVC is now operational
	svcResp, svcErr := invService.CheckSVCStatus(farmConfig.CertificateData, certPassword)
	if svcErr != nil || svcResp == nil || !svcResp.IsSVCOperational() {
		// SVC still not available — keep as draft, increment retry count
		if err := nfeModel.IncrementRetryCount(inv.ID); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to increment retry count for draft invoice %d: %v", inv.ID, err))
		}
		return nil
	}

	// SVC is active — attempt to rebuild and send via SVC
	// We need the original unitPrice from the invoice record
	invoiceRecord, invErr := nfeModel.GetInvoiceByAccessKey(inv.AccessKey)
	if invErr != nil || invoiceRecord == nil {
		return fmt.Errorf("failed to get invoice record for draft rebuild: %w", invErr)
	}

	tpEmis := defaults.SVCForState(farmConfig.EmitterUF)
	now := time.Now()
	reason := "Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"

	// Get the original user-typed tax rates from the persisted row.
	// For legacy invoices (no row) or partially-set rows, individual rates will
	// be nil and MergeRates will fall back to the farm config.
	 taxRateOverrides := invoiceRecord.TaxRates
	if taxRateOverrides == nil {
		taxRateOverrides = &entity.TaxRates{}
	}
	rates := MergeRates(*taxRateOverrides, farmConfig)

	// Rebuild the per-emission overrides from the invoice record so the
	// contingency NF-e is identical to the original attempt.
	invoiceOverrides := invoiceOverridesFromRecord(invoiceRecord)

	// Get the original full invoice input data
	nfeSvc := NewNFeService()
	input, _, _, _, toast := nfeSvc.prepareInvoiceBuildData(departure.Id, invoiceRecord.UnitPrice, uint32(departure.Farm), "", rates, invoiceOverrides)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return fmt.Errorf("failed to prepare rebuild data: %s", toast.Message)
	}

	// Allocate new number for the contingency invoice
	newNumber, allocErr := nfeModel.AllocateNumber(uint32(departure.Farm), farmConfig.Serie)
	if allocErr != nil {
		return fmt.Errorf("failed to allocate contingency number: %w", allocErr)
	}

	newSignedXML, newAccessKey, rebuildErr := invService.RebuildForContingency(input, newNumber, generateRandomCNF(), tpEmis, reason, farmConfig.CertificateData, certPassword)
	if rebuildErr != nil {
		return fmt.Errorf("failed to rebuild for SVC: %w", rebuildErr)
	}

	// Persist the new contingency invoice. Carry the same user-typed rates and
	// overrides so the worker's next retry uses the same values.
	var ratesToPersist *entity.TaxRates
	if hasAnyUserRate(*taxRateOverrides) {
		ratesToPersist = taxRateOverrides
	}
	// Sum the per-item IBS/CBS totals from the rebuilt input for the totals
	// columns. input.Items was populated by prepareInvoiceBuildData above.
	var totalIBSValue, totalCBSValue decimal.Decimal
	for _, item := range input.Items {
		totalIBSValue = totalIBSValue.Add(item.Imposto.IBSCBS.VIBS)
		totalCBSValue = totalCBSValue.Add(item.Imposto.IBSCBS.VCBS)
	}
	newInvoiceID, createErr := nfeModel.CreateInvoiceWithEmission(
		departure.Id, newAccessKey, farmConfig.Serie, newNumber,
		invoiceRecord.CFOP, invoiceRecord.NCM, invoiceRecord.QuantityKG, invoiceRecord.UnitPrice, invoiceRecord.TotalValue,
		totalIBSValue, totalCBSValue,
		int(tpEmis), now, reason, ratesToPersist, invoiceOverrides,
	)
	if createErr != nil {
		return fmt.Errorf("failed to create contingency invoice: %w", createErr)
	}

	if err := nfeModel.UpdateInvoiceSignedXML(newInvoiceID, newSignedXML); err != nil {
		return fmt.Errorf("failed to save contingency XML: %w", err)
	}

	// Send to SVC
	sefazResp, sendErr := invService.SendToSefazWithEmission(newSignedXML, farmConfig.CertificateData, certPassword, tpEmis)
	if sendErr != nil {
		// SVC failed unexpectedly — mark new as draft, supersede old
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: "+sendErr.Error())
		_ = nfeModel.SupersedeInvoice(inv.ID, newInvoiceID)
		if err := nfeModel.IncrementRetryCount(inv.ID); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to increment retry count for draft invoice %d: %v", inv.ID, err))
		}
		return sendErr
	}

	if sefazResp == nil {
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: resposta vazia")
		_ = nfeModel.SupersedeInvoice(inv.ID, newInvoiceID)
		if err := nfeModel.IncrementRetryCount(inv.ID); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to increment retry count for draft invoice %d: %v", inv.ID, err))
		}
		return fmt.Errorf("empty response from SVC")
	}

	// SVC responded — handle the response and supersede the old draft
	if sefazResp.IsAuthorized() {
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "authorized", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
	} else if sefazResp.IsProcessing() || sefazResp.IsAccepted() {
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "pending", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
	} else if sefazResp.IsRejected() {
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "denied", "", sefazResp.StatusCode, sefazResp.StatusMotive)
	} else {
		_ = nfeModel.UpdateInvoiceStatus(newInvoiceID, "pending", "", sefazResp.StatusCode, sefazResp.StatusMotive)
	}

	_ = nfeModel.SupersedeInvoice(inv.ID, newInvoiceID)
	return nil
}

// invoiceOverridesFromRecord reconstructs an InvoiceOverrides struct from the
// persisted invoice record so that draft retries and SVC rebuilds use the exact
// same per-emission values the user originally entered.
func invoiceOverridesFromRecord(inv *nfe_model.Invoice) *entity.InvoiceOverrides {
	if inv == nil {
		return nil
	}
	hasValue := false
	o := &entity.InvoiceOverrides{}
	if inv.NaturezaOp != nil && *inv.NaturezaOp != "" {
		o.NaturezaOp = inv.NaturezaOp
		hasValue = true
	}
	if inv.ProductDescription != nil && *inv.ProductDescription != "" {
		o.ProductDesc = inv.ProductDescription
		hasValue = true
	}
	if inv.CEST != nil && *inv.CEST != "" {
		o.CEST = inv.CEST
		hasValue = true
	}
	if inv.Unit != nil && *inv.Unit != "" {
		o.Unit = inv.Unit
		hasValue = true
	}
	if inv.ModFrete != nil {
		o.ModFrete = inv.ModFrete
		hasValue = true
	}
	if inv.ICMSCST != nil && *inv.ICMSCST != "" {
		o.ICMSCST = inv.ICMSCST
		hasValue = true
	}
	if inv.PISCST != nil && *inv.PISCST != "" {
		o.PISCST = inv.PISCST
		hasValue = true
	}
	if inv.COFINSCST != nil && *inv.COFINSCST != "" {
		o.COFINSCST = inv.COFINSCST
		hasValue = true
	}
	// infCpl is restored even when empty because the user may have explicitly
	// cleared the field. A nil column means "not set at all" (legacy invoice).
	if inv.InfCpl != nil {
		o.InfCpl = inv.InfCpl
		hasValue = true
	}
	if !hasValue {
		return nil
	}
	return o
}
