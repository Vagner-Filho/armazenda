package nfe_service

import (
	"testing"

	entity_public "armazenda/entity/public"
)

func TestNFeCertGate_Helper(t *testing.T) {
	cert := []byte("fake-cert-bytes")

	cases := []struct {
		name    string
		cfg     *entity_public.FarmConfig
		wantNil bool
	}{
		{"nil config", nil, false},
		{"nil cert bytes", &entity_public.FarmConfig{}, false},
		{"empty cert bytes", &entity_public.FarmConfig{CertificateData: []byte{}}, false},
		{"cert present", &entity_public.FarmConfig{CertificateData: cert}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := nfeCertGate(c.cfg)
			if c.wantNil && got != nil {
				t.Fatalf("expected nil toast, got %+v", got)
			}
			if !c.wantNil && got == nil {
				t.Fatalf("expected warning toast, got nil")
			}
			if !c.wantNil && got != nil {
				if got.Message != "Certificado digital não configurado" {
					t.Errorf("Message = %q, want %q", got.Message, "Certificado digital não configurado")
				}
				if got.Hint != "Envie o arquivo do certificado (.pfx, .p12) e a senha em NF-e em Configurações" {
					t.Errorf("Hint = %q, want %q", got.Hint, "Envie o arquivo do certificado (.pfx, .p12) e a senha em NF-e em Configurações")
				}
			}
		})
	}
}
