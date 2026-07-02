package defaults_test

import (
	"testing"

	"armazenda/pkg/nfe/defaults"
)

func TestTpEmis_String(t *testing.T) {
	tests := []struct {
		input    defaults.TpEmis
		expected string
	}{
		{defaults.EmissaoNormal, "1"},
		{defaults.EPEC, "4"},
		{defaults.FSDA, "5"},
		{defaults.SVCAN, "6"},
		{defaults.SVCRS, "7"},
		{defaults.TpEmis(99), "1"}, // default fallback
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTpEmis_IsContingency(t *testing.T) {
	tests := []struct {
		input    defaults.TpEmis
		expected bool
	}{
		{defaults.EmissaoNormal, false},
		{defaults.EPEC, true},
		{defaults.FSDA, true},
		{defaults.SVCAN, true},
		{defaults.SVCRS, true},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			if got := tt.input.IsContingency(); got != tt.expected {
				t.Errorf("IsContingency() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSVCForState(t *testing.T) {
	tests := []struct {
		uf       string
		expected defaults.TpEmis
	}{
		{"MT", defaults.SVCRS},     // Mato Grosso → SVC-RS
		{"MS", defaults.SVCRS},     // Mato Grosso do Sul → SVC-RS
		{"SP", defaults.SVCAN},     // São Paulo → SVC-AN
		{"RS", defaults.SVCAN},     // Rio Grande do Sul → SVC-AN
		{"XX", defaults.EmissaoNormal}, // Unknown state → normal
	}
	for _, tt := range tests {
		t.Run(tt.uf, func(t *testing.T) {
			if got := defaults.SVCForState(tt.uf); got != tt.expected {
				t.Errorf("SVCForState(%s) = %v, want %v", tt.uf, got, tt.expected)
			}
		})
	}
}
