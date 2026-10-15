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

func TestNaturezaOpForCFOP(t *testing.T) {
	tests := []struct {
		cfop     string
		expected string
	}{
		{"5101", "Venda de producao do estabelecimento"},
		{"5102", "Venda de mercadoria adquirida ou recebida de terceiros"},
		{"5901", "Remessa para industrializacao"},
		{"6202", "Devolucao de compra"},
		{"5103", "Venda de producao do estabelecimento ao contribuinte do ICMS"},
		{"5104", "Venda de producao do estabelecimento a nao contribuinte"},
		{"9999", "Venda de mercadoria"},
	}
	for _, tt := range tests {
		t.Run(tt.cfop, func(t *testing.T) {
			if got := defaults.NaturezaOpForCFOP(tt.cfop); got != tt.expected {
				t.Errorf("NaturezaOpForCFOP(%s) = %q, want %q", tt.cfop, got, tt.expected)
			}
		})
	}
}

// TestFinNFeForCFOP verifies the curated devolução (finNFe=4) CFOP set:
// corpus anchors and the x202/x555 families flip to "4", everything else
// stays "1" (normal emission).
func TestFinNFeForCFOP(t *testing.T) {
	tests := []struct {
		cfop     string
		expected string
	}{
		// Corpus G3 anchor: devolução de compra de material de uso e consumo.
		{"5556", "4"},
		// Full curated return set.
		{"1202", "4"}, {"2202", "4"}, {"5202", "4"}, {"6202", "4"},
		{"5555", "4"}, {"6555", "4"}, {"6556", "4"},
		{"5557", "4"}, {"6557", "4"}, {"5558", "4"}, {"6558", "4"},
		{"5559", "4"}, {"6559", "4"},
		// Corpus non-return anchors — these must stay finNFe 1.
		{"5101", "1"}, // G1 venda de produção
		{"5118", "1"}, // G2 vendas a ordem
		{"5913", "1"}, // G4 retorno de demonstração (corpus: finNFe 1)
		{"5915", "1"}, // G5 remessa para conserto
		{"5551", "1"}, // G6 venda de imobilizado
		{"5949", "1"}, // G7 outras saídas
		{"5901", "1"}, // remessa para industrialização
	}
	for _, tt := range tests {
		t.Run(tt.cfop, func(t *testing.T) {
			if got := defaults.FinNFeForCFOP(tt.cfop); got != tt.expected {
				t.Errorf("FinNFeForCFOP(%s) = %q, want %q", tt.cfop, got, tt.expected)
			}
			if got := defaults.IsDevolucaoCFOP(tt.cfop); got != (tt.expected == "4") {
				t.Errorf("IsDevolucaoCFOP(%s) = %v, inconsistent with FinNFeForCFOP", tt.cfop, got)
			}
		})
	}
}
