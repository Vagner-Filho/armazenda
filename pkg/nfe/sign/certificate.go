package sign

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

// Certificate holds the loaded A1 certificate data.
type Certificate struct {
	PrivateKey crypto.Signer
	Leaf       *x509.Certificate
	Chain      []*x509.Certificate
}

// LoadCertificate loads an A1 certificate from a .pfx (PKCS12) file.
func LoadCertificate(pfxPath, password string) (*Certificate, error) {
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	privateKey, certificate, caChain, err := pkcs12.DecodeChain(pfxData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PKCS12: %w", err)
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	return &Certificate{
		PrivateKey: signer,
		Leaf:       certificate,
		Chain:      caChain,
	}, nil
}

// RawChain returns the certificate chain as raw DER bytes.
func (c *Certificate) RawChain() [][]byte {
	chain := [][]byte{c.Leaf.Raw}
	for _, cert := range c.Chain {
		chain = append(chain, cert.Raw)
	}
	return chain
}

// Valid checks if the certificate is currently valid.
func (c *Certificate) Valid() bool {
	return c.Leaf != nil && c.Leaf.NotBefore.Before(time.Now()) && c.Leaf.NotAfter.After(time.Now())
}

// DaysUntilExpiry returns the number of days until the certificate expires.
func (c *Certificate) DaysUntilExpiry() int {
	if c.Leaf == nil {
		return 0
	}
	return int(c.Leaf.NotAfter.Sub(time.Now()).Hours() / 24)
}
