package config

import (
	"crypto/tls"
	"time"

	"armazenda/pkg/nfe/defaults"
)

// Environment represents the SEFAZ environment type.
type Environment int

const (
	EnvironmentProduction   Environment = 1
	EnvironmentHomologation Environment = 2
)

func (e Environment) String() string {
	switch e {
	case EnvironmentProduction:
		return "producao"
	case EnvironmentHomologation:
		return "homologacao"
	default:
		return "homologacao"
	}
}

// SefazConfig holds the configuration for SEFAZ communication.
type SefazConfig struct {
	Environment Environment
	StateUF     string // e.g., "MT"
	Timeout     time.Duration
	Certificate CertificateConfig

	// Contingency fields (populated when entering SVC mode).
	ContingencyMode   *defaults.TpEmis // nil = normal emission
	ContingencyReason string
	ContingencyStart  time.Time
}

// CertificateConfig holds the digital certificate (A1) settings.
type CertificateConfig struct {
	Path     string
	Password string
}

// TLSConfig returns the TLS configuration for mTLS communication with SEFAZ.
func (c SefazConfig) TLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.Environment == EnvironmentHomologation,
		Renegotiation:      tls.RenegotiateFreelyAsClient,
	}
}
