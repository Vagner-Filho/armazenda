package service

import (
	"fmt"
	"strings"
	"time"

	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/sefaz"
	"armazenda/pkg/nfe/sign"
	"armazenda/pkg/nfe/xml"
)

// InvoiceService orchestrates the NF-e lifecycle.
// It never connects to a database — persistence is handled by nfe_model.
type InvoiceService struct {
	config config.SefazConfig
}

// NewInvoiceService creates a new invoice service.
func NewInvoiceService(cfg config.SefazConfig) *InvoiceService {
	return &InvoiceService{
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

	emitterDoc := strings.ReplaceAll(input.Emitter.Document, ".", "")

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

// BuildAndSignCancellationEvent builds and signs a cancellation event
// (tpEvento=110111) for an authorized NF-e, returning the serialized
// <envEvento> document ready to be sent to RecepcaoEvento4.
func (s *InvoiceService) BuildAndSignCancellationEvent(input xml.CancelEventInput, certData []byte, certPassword string) (string, error) {
	doc, err := xml.BuildCancellationEvent(input)
	if err != nil {
		return "", fmt.Errorf("failed to build cancellation event XML: %w", err)
	}

	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return "", fmt.Errorf("failed to load certificate: %w", err)
	}

	signer := sign.NewSigner(cert)
	if err := signer.SignEventDocument(doc); err != nil {
		return "", fmt.Errorf("failed to sign cancellation event: %w", err)
	}

	str, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize cancellation event XML: %w", err)
	}

	return str, nil
}

// SendCancellationEvent sends a signed cancellation event (<envEvento>) to the
// RecepcaoEvento4 endpoint matching the emission type (normal SEFAZ or SVC)
// and returns the parsed response. Per the MOC contingency annex, a
// cancellation must be registered in the same environment that authorized the
// NF-e, so the caller must pass the invoice's tpEmis.
func (s *InvoiceService) SendCancellationEvent(signedEventXML string, certData []byte, certPassword string, tpEmis defaults.TpEmis) (*sefaz.EventoResponse, error) {
	cert, err := sign.LoadCertificateFromBytes(certData, certPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client, err := sefaz.NewClient(s.config, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create SEFAZ client: %w", err)
	}

	url, ns, action, err := sefaz.GetEndpointWithSOAPActionAndEmission(s.config.StateUF, "RecepcaoEvento4", s.config.Environment == config.EnvironmentProduction, tpEmis)
	if err != nil {
		return nil, err
	}

	cUF := defaults.UFCode(s.config.StateUF)
	soapBody := xml.BuildSOAPEnvelopeWithCabecMsg(ns, signedEventXML, cUF, "1.00")
	resp, err := client.Post(url, action, []byte(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send cancellation event to SEFAZ: %w", err)
	}

	parsed, parseErr := sefaz.ParseEventoResponse(resp)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse SEFAZ event response: %w", parseErr)
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
		return padLeftZeros(emit.Document, 14)
	}
	return padLeftZeros(emit.Document, 14)
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
