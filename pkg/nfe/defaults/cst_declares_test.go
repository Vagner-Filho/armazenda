package defaults_test

import (
	"testing"

	"armazenda/pkg/nfe/defaults"
)

// Table-driven coverage for the CST/group predicates added by
// docs/sdd/nfe_cst_zero_values: which ICMS and PIS/COFINS XML groups
// actually declare vBC/percentage/value elements.
func TestICMSDeclaresValues(t *testing.T) {
	tests := []struct {
		cst  string
		want bool
	}{
		{"00", true}, {"20", true}, {"51", true}, {"90", true},
		{"40", false}, {"41", false}, {"50", false},
		{"", false}, {"99", false},
	}
	for _, tt := range tests {
		t.Run(tt.cst, func(t *testing.T) {
			if got := defaults.ICMSDeclaresValues(tt.cst); got != tt.want {
				t.Errorf("ICMSDeclaresValues(%q) = %v, want %v", tt.cst, got, tt.want)
			}
		})
	}
}

func TestPISAndCOFINSDeclaresValues(t *testing.T) {
	tests := []struct {
		cst  string
		want bool
	}{
		{"01", true}, {"02", true},
		{"04", false}, {"05", false}, {"06", false}, {"07", false}, {"08", false}, {"09", false},
		{"", false}, {"99", false},
	}
	for _, tt := range tests {
		t.Run(tt.cst, func(t *testing.T) {
			if got := defaults.PISDeclaresValues(tt.cst); got != tt.want {
				t.Errorf("PISDeclaresValues(%q) = %v, want %v", tt.cst, got, tt.want)
			}
			if got := defaults.COFINSDeclaresValues(tt.cst); got != tt.want {
				t.Errorf("COFINSDeclaresValues(%q) = %v, want %v", tt.cst, got, tt.want)
			}
		})
	}
}
