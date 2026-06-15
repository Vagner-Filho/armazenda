package sign

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"software.sslmate.com/src/go-pkcs12"
)

// Certificate holds the loaded A1 certificate data.
type Certificate struct {
	PrivateKey crypto.Signer
	Leaf       *x509.Certificate
	Chain      []*x509.Certificate
}

// LoadCertificate loads an A1 certificate from a .pfx/.p12 (PKCS12) file path.
func LoadCertificate(pfxPath, password string) (*Certificate, error) {
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return LoadCertificateFromBytes(pfxData, password)
}

// LoadCertificateFromBytes loads an A1 certificate from raw PKCS#12 bytes.
func LoadCertificateFromBytes(pfxData []byte, password string) (*Certificate, error) {
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

// GetDocument extracts the CPF or CNPJ embedded in the certificate.
// Brazilian A1 certificates encode the holder's document in Subject fields.
// Returns the document string (11 digits for CPF, 14 for CNPJ) and an error
// if no valid document is found.
func (c *Certificate) GetDocument() (string, error) {
	if c.Leaf == nil {
		return "", fmt.Errorf("certificate leaf is nil")
	}

	subject := c.Leaf.Subject

	// Primary: CommonName often contains "NAME:CPF" for e-CPF A1 certificates
	if doc := extractFromCommonName(subject.CommonName); doc != "" {
		return doc, nil
	}

	// Secondary: SerialNumber may contain the document directly
	if doc := cleanDigits(subject.SerialNumber); len(doc) == 11 || len(doc) == 14 {
		return doc, nil
	}

	// Tertiary: Check OU (OrganizationalUnit) fields in Names
	for _, name := range subject.Names {
		raw := fmt.Sprintf("%v", name.Value)
		doc := cleanDigits(raw)
		if len(doc) == 11 || len(doc) == 14 {
			return doc, nil
		}
	}

	return "", fmt.Errorf("no CPF or CNPJ found in certificate subject")
}

// extractFromCommonName parses Brazilian A1 CommonName formats.
// Typical e-CPF A1 format: "FULL NAME:12345678901"
// Returns the 11-digit CPF suffix if found, or empty string.
func extractFromCommonName(cn string) string {
	parts := strings.Split(cn, ":")
	if len(parts) < 2 {
		return ""
	}
	suffix := strings.TrimSpace(parts[len(parts)-1])
	cleaned := cleanDigits(suffix)
	if len(cleaned) == 11 {
		return cleaned
	}
	return ""
}

// cleanDigits strips any non-numeric characters from a string.
func cleanDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}
