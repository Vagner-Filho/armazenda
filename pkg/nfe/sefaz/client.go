package sefaz

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/sign"
	"armazenda/pkg/nfe/xml"
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
		// Build certificate chain for mTLS
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

// Post sends a SOAP request to the given endpoint with the specified SOAPAction.
func (c *Client) Post(endpointURL, soapAction string, soapBody []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

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

// CheckStatus checks if the SEFAZ service is available and returns the parsed response.
func (c *Client) CheckStatus() (*StatusResponse, error) {
	url, ns, action, err := GetEndpointWithSOAPAction(c.cfg.StateUF, "NfeStatusServico4", c.cfg.Environment == config.EnvironmentProduction)
	if err != nil {
		return nil, err
	}

	cUF := defaults.UFCode(c.cfg.StateUF)
	tpAmb := "2"
	if c.cfg.Environment == config.EnvironmentProduction {
		tpAmb = "1"
	}

	payload := fmt.Sprintf(`<consStatServ xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>%s</tpAmb><cUF>%s</cUF><xServ>STATUS</xServ></consStatServ>`, tpAmb, cUF)

	soapBody := xml.BuildSOAPEnvelope(ns, payload)

	resp, err := c.Post(url, action, []byte(soapBody))
	if err != nil {
		return nil, fmt.Errorf("SEFAZ status check failed: %w", err)
	}

	parsed, parseErr := ParseStatusResponse(resp)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse SEFAZ status response: %w", parseErr)
	}

	return parsed, nil
}
