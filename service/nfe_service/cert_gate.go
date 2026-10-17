package nfe_service

import (
	entity_public "armazenda/entity/public"
)

// nfeCertGate returns nil when the farm config carries certificate bytes,
// otherwise a pt-br warning toast pointing the user at the config screen. It is
// the single source for the certificate-presence check used by both emission
// paths (departure and detached).
func nfeCertGate(cfg *entity_public.FarmConfig) *entity_public.Toast {
	if cfg != nil && len(cfg.CertificateData) > 0 {
		return nil
	}
	toast := entity_public.GetWarningToast(
		"Certificado digital não configurado",
		"Envie o arquivo do certificado (.pfx, .p12) e a senha em NF-e em Configurações",
	)
	return &toast
}
