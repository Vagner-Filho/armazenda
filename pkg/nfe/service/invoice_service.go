package service

import (
	"context"
	"fmt"
	"time"

	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/sefaz"
	"armazenda/pkg/nfe/sign"
	"armazenda/pkg/nfe/xml"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// InvoiceService orchestrates the NF-e lifecycle.
type InvoiceService struct {
	pool   *pgxpool.Pool
	config config.SefazConfig
}

// NewInvoiceService creates a new invoice service.
func NewInvoiceService(pool *pgxpool.Pool, cfg config.SefazConfig) *InvoiceService {
	return &InvoiceService{
		pool:   pool,
		config: cfg,
	}
}

// BuildAndSign builds and signs an NF-e from departure data.
func (s *InvoiceService) BuildAndSign(input entity.InvoiceInput, certData []byte, certPassword string) (string, error) {
	// 1. Build XML
	builder := xml.NewBuilder()
	doc, err := builder.Build(input)
	if err != nil {
		return "", fmt.Errorf("failed to build NF-e XML: %w", err)
	}

	// 2. Load certificate from bytes
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return "", fmt.Errorf("failed to load certificate: %w", err)
	}

	// 2a. Validate certificate belongs to the emitter
	certDoc, docErr := cert.GetDocument()
	if docErr != nil {
		return "", fmt.Errorf("failed to extract document from certificate: %w", docErr)
	}
	var emitterDoc string
	if input.Emitter.Type == 2 {
		emitterDoc = input.Emitter.CPF
	} else {
		emitterDoc = input.Emitter.CNPJ
	}
	if certDoc != emitterDoc {
		return "", fmt.Errorf("certificado digital não pertence ao emitente: certificado=%s, emitente=%s", certDoc, emitterDoc)
	}

	// 3. Sign XML
	signer := sign.NewSigner(cert)
	if err := signer.SignDocument(doc); err != nil {
		return "", fmt.Errorf("failed to sign NF-e: %w", err)
	}

	// 4. Serialize to string
	str, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}

	return str, nil
}

// SendToSefaz sends a signed NF-e to SEFAZ (normal endpoint) and returns the parsed response.
func (s *InvoiceService) SendToSefaz(signedXML string, certData []byte, certPassword string) (*sefaz.AutorizacaoResponse, error) {
	return s.SendToSefazWithEmission(signedXML, certData, certPassword, defaults.EmissaoNormal)
}

// SendToSefazWithEmission sends a signed NF-e to the endpoint matching the emission type
// (normal SEFAZ or SVC) and returns the parsed response.
func (s *InvoiceService) SendToSefazWithEmission(signedXML string, certData []byte, certPassword string, tpEmis defaults.TpEmis) (*sefaz.AutorizacaoResponse, error) {
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client, err := sefaz.NewClient(s.config, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEFAZ client: %w", err)
	}

	url, ns, action, err := sefaz.GetEndpointWithSOAPActionAndEmission(s.config.StateUF, "NFeAutorizacao4", s.config.Environment == config.EnvironmentProduction, tpEmis)
	if err != nil {
		return nil, err
	}

	// Wrap the signed NF-e in <enviNFe> batch envelope as required by SEFAZ
	batchXML := fmt.Sprintf(`<enviNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><idLote>1</idLote><indSinc>1</indSinc>%s</enviNFe>`, signedXML)

	soapBody := xml.BuildSOAPEnvelope(ns, batchXML)
	resp, err := client.Post(url, action, []byte(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send to SEFAZ: %w", err)
	}

	parsed, parseErr := sefaz.ParseAutorizacaoResponse(resp)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse SEFAZ response: %w", parseErr)
	}

	return parsed, nil
}

// CheckSVCStatus checks if the SVC for the configured state is operational.
func (s *InvoiceService) CheckSVCStatus(certData []byte, certPassword string) (*sefaz.StatusResponse, error) {
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client, err := sefaz.NewClient(s.config, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEFAZ client: %w", err)
	}

	return client.CheckSVCStatus()
}

// CheckSefazStatus checks if the configured SEFAZ is operational.
func (s *InvoiceService) CheckSefazStatus(certData []byte, certPassword string) (*sefaz.StatusResponse, error) {
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client, err := sefaz.NewClient(s.config, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEFAZ client: %w", err)
	}

	return client.CheckStatus()
}

// SaveInvoice saves a signed NF-e to the database with status 'signed'.
func (s *InvoiceService) SaveInvoice(ctx context.Context, departureID uint32, accessKey string, serie, number int, xmlSigned string) error {
	query := `
		INSERT INTO nfe_invoice (departure_id, access_key, serie, number, status, xml_signed, signed_at)
		VALUES ($1, $2, $3, $4, 'signed', $5, $6)
	`
	_, err := s.pool.Exec(ctx, query, departureID, accessKey, serie, number, xmlSigned, time.Now())
	return err
}

// GetPendingInvoices returns all invoices with status 'pending'.
func (s *InvoiceService) GetPendingInvoices(ctx context.Context) ([]PendingInvoice, error) {
	query := `
		SELECT id, departure_id, access_key, xml_signed, retry_count
		FROM nfe_invoice
		WHERE status IN ('pending', 'signed')
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []PendingInvoice
	for rows.Next() {
		var inv PendingInvoice
		if err := rows.Scan(&inv.ID, &inv.DepartureID, &inv.AccessKey, &inv.XMLSigned, &inv.RetryCount); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

// UpdateInvoiceStatus updates the status of an invoice.
func (s *InvoiceService) UpdateInvoiceStatus(ctx context.Context, id int, status string, protocol, sefazCode, sefazMotive string) error {
	query := `
		UPDATE nfe_invoice
		SET status = $2, protocol = $3, sefaz_status_code = $4, sefaz_motive = $5,
		    authorized_at = CASE WHEN $2 = 'authorized' THEN $6 ELSE authorized_at END
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query, id, status, protocol, sefazCode, sefazMotive, time.Now())
	return err
}

// PendingInvoice represents an invoice waiting to be sent to SEFAZ.
type PendingInvoice struct {
	ID          int
	DepartureID uint32
	AccessKey   string
	XMLSigned   string
	RetryCount  int
}

// QueryInvoiceStatus queries the SEFAZ for the current status of an invoice by access key.
func (s *InvoiceService) QueryInvoiceStatus(accessKey string, certData []byte, certPassword string) (*sefaz.ConsultaResponse, error) {
	return s.QueryInvoiceStatusWithEmission(accessKey, certData, certPassword, defaults.EmissaoNormal)
}

// QueryInvoiceStatusWithEmission queries the endpoint matching the emission type for the current status.
func (s *InvoiceService) QueryInvoiceStatusWithEmission(accessKey string, certData []byte, certPassword string, tpEmis defaults.TpEmis) (*sefaz.ConsultaResponse, error) {
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client, err := sefaz.NewClient(s.config, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEFAZ client: %w", err)
	}

	url, ns, action, err := sefaz.GetEndpointWithSOAPActionAndEmission(s.config.StateUF, "NFeConsultaProtocolo4", s.config.Environment == config.EnvironmentProduction, tpEmis)
	if err != nil {
		return nil, err
	}

	tpAmb := "2"
	if s.config.Environment == config.EnvironmentProduction {
		tpAmb = "1"
	}

	payload := fmt.Sprintf(`<consSitNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>%s</tpAmb><xServ>CONSULTAR</xServ><chNFe>%s</chNFe></consSitNFe>`, tpAmb, accessKey)

	soapBody := xml.BuildSOAPEnvelope(ns, payload)
	resp, err := client.Post(url, action, []byte(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to query SEFAZ: %w", err)
	}

	parsed, parseErr := sefaz.ParseConsultaResponse(resp)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse SEFAZ query response: %w", parseErr)
	}

	return parsed, nil
}

// RebuildForContingency rebuilds the NF-e XML for a contingency emission mode.
// It returns the newly signed XML and the new access key.
func (s *InvoiceService) RebuildForContingency(input entity.InvoiceInput, newNumber int, newCNF string, tpEmis defaults.TpEmis, reason string, certData []byte, certPassword string) (string, string, error) {
	now := time.Now()
	input.Numero = newNumber
	input.CNF = newCNF
	input.TpEmis = tpEmis
	input.DhCont = &now
	input.XJust = reason

	signedXML, err := s.BuildAndSign(input, certData, certPassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to rebuild NF-e for contingency: %w", err)
	}

	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:      defaults.UFCode(input.Emitter.UF),
		AAMM:     now.Format("0601"),
		Document: documentForAccessKey(input.Emitter),
		Mod:      defaults.ModeloNFe,
		Serie:    input.Serie,
		NNF:      input.Numero,
		TpEmis:   tpEmis.String(),
		CNF:      input.CNF,
	})

	return signedXML, accessKey, nil
}

func documentForAccessKey(emit entity.EmitterData) string {
	if emit.Type == 2 {
		return padLeftZeros(emit.CPF, 14)
	}
	return padLeftZeros(emit.CNPJ, 14)
}

func padLeftZeros(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}
	if len(s) > length {
		return s[:length]
	}
	return s
}

// CalculateTaxes calculates the default taxes for a grain sale.
func CalculateTaxes(unitPrice, quantity decimal.Decimal, regime defaults.TaxRegime) (icms, pis, cofins decimal.Decimal) {
	// Default ICMS rate for MT: 17%
	icmsRate := decimal.NewFromFloat(0.17)
	// Default PIS rate: 1.65%
	pisRate := decimal.NewFromFloat(0.0165)
	// Default COFINS rate: 7.6%
	cofinsRate := decimal.NewFromFloat(0.076)

	baseValue := unitPrice.Mul(quantity)

	icms = baseValue.Mul(icmsRate)
	pis = baseValue.Mul(pisRate)
	cofins = baseValue.Mul(cofinsRate)

	// Simples Nacional: different calculation (simplified)
	if regime == defaults.TaxRegimeSimplesNacional {
		// For SN, ICMS is calculated but may have different rates
		// This is a simplified placeholder
		icms = baseValue.Mul(decimal.NewFromFloat(0.12))
	}

	return icms, pis, cofins
}
