package sefaz

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/sign"
)

// Client handles HTTP communication with SEFAZ.
type Client struct {
	cfg        config.SefazConfig
	httpClient *http.Client
}

// NewClient creates a new SEFAZ HTTP client.
func NewClient(cfg config.SefazConfig, cert *sign.Certificate) (*Client, error) {
	tlsConfig := cfg.TLSConfig()

	if cert != nil {
		// mTLS: add client certificate
		certPool := x509.NewCertPool()
		for _, c := range cert.Chain {
			certPool.AddCert(c)
		}
		certPool.AddCert(cert.Leaf)

		// Build certificate chain for TLS
		certChain := [][]byte{cert.Leaf.Raw}
		for _, c := range cert.Chain {
			certChain = append(certChain, c.Raw)
		}

		tlsConfig.Certificates = []tls.Certificate{
			{
				Certificate: certChain,
				PrivateKey:  cert.PrivateKey,
				Leaf:        cert.Leaf,
			},
		}
		tlsConfig.RootCAs = certPool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
	}, nil
}

// Post sends a SOAP request to the given endpoint.
func (c *Client) Post(endpointURL string, soapBody []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("SEFAZ returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// CheckStatus checks if the SEFAZ service is available.
func (c *Client) CheckStatus() error {
	url, ns, err := GetEndpointWithNamespace(c.cfg.StateUF, "NfeStatusServico4", c.cfg.Environment == config.EnvironmentProduction)
	if err != nil {
		return err
	}

	cUF := defaults.UFCode(c.cfg.StateUF)
	tpAmb := "2"
	if c.cfg.Environment == config.EnvironmentProduction {
		tpAmb = "1"
	}

	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap12:Header>
    <nfeCabecMsg xmlns="%s">
      <cUF>%s</cUF>
      <versaoDados>4.00</versaoDados>
    </nfeCabecMsg>
  </soap12:Header>
  <soap12:Body>
    <nfeDadosMsg xmlns="%s">
      <consStatServ xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00">
        <tpAmb>%s</tpAmb>
        <cUF>%s</cUF>
        <xServ>STATUS</xServ>
      </consStatServ>
    </nfeDadosMsg>
  </soap12:Body>
</soap12:Envelope>`, ns, cUF, ns, tpAmb, cUF)

	resp, err := c.Post(url, []byte(soapBody))
	if err != nil {
		return fmt.Errorf("SEFAZ status check failed: %w", err)
	}

	// TODO: Parse response to check if service is operational
	_ = resp
	return nil
}
