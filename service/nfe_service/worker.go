package nfe_service

import (
	"fmt"
	"time"

	"armazenda/model/departure_model"
	model_error "armazenda/model/error"
	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/service"
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
	invoices, err := nfeModel.GetPendingInvoicesForRetry()
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to get pending invoices: %v", err))
		return
	}

	if len(invoices) == 0 {
		return
	}

	for _, inv := range invoices {
		if err := processInvoice(inv); err != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to process invoice %d: %v", inv.ID, err))
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
	invService := service.NewInvoiceService(nil, sefazCfg)

	switch inv.Status {
	case "signed":
		// Never sent to SEFAZ — send now
		if inv.XMLSigned == "" {
			return fmt.Errorf("invoice %d has no signed XML", inv.ID)
		}

		resp, sendErr := invService.SendToSefaz(inv.XMLSigned, farmConfig.CertificateData, certPassword)
		if sendErr != nil {
			// Network error — keep pending, do NOT increment retry count
			updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "pending", "", "", "SEFAZ unavailable: "+sendErr.Error())
			if updErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
			}
			return sendErr
		}

		if resp == nil {
			updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "pending", "", "", "Empty response from SEFAZ")
			if updErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
			}
			return fmt.Errorf("empty response from SEFAZ")
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
		} else if resp.IsProcessing() || resp.IsAccepted() {
			updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "pending", resp.Protocol, resp.StatusCode, resp.StatusMotive)
			if updErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
			}
		} else if resp.IsRejected() {
			updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "denied", "", resp.StatusCode, resp.StatusMotive)
			if updErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
			}
		} else {
			updErr := nfeModel.UpdateInvoiceStatus(inv.ID, "pending", "", resp.StatusCode, resp.StatusMotive)
			if updErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("Retry worker: failed to update status: %v", updErr))
			}
		}

	case "pending":
		// Already sent, query status
		resp, queryErr := invService.QueryInvoiceStatus(inv.AccessKey, farmConfig.CertificateData, certPassword)
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
	}

	return nil
}
