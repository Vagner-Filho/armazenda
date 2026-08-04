package xml_test

import (
	"regexp"
	"testing"

	"armazenda/pkg/nfe/xml"
)

// tStringPattern mirrors the schema TString type restriction:
// printable Latin-1 only, no leading/trailing space.
var tStringPattern = regexp.MustCompile(`^([!-ÿ]|[!-ÿ][ -ÿ]*[!-ÿ])$`)

func TestSanitizeSchemaString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text unchanged", "Soja em grão Nº 123", "Soja em grão Nº 123"},
		{"accents preserved", "válido até çÇ áéíóú", "válido até çÇ áéíóú"},
		{"newline becomes separator", "linha1\nlinha2", "linha1; linha2"},
		{"CRLF becomes separator", "linha1\r\nlinha2", "linha1; linha2"},
		{"multiple newlines collapse", "a\n\n\nb", "a; b"},
		{"blank lines dropped", "a\n   \nb", "a; b"},
		{"leading trailing trimmed", "  texto  ", "texto"},
		{"per-line trimmed", "  a  \n  b  ", "a; b"},
		{"tab becomes space", "a\tb", "a b"},
		{"control chars dropped", "a\x00\x07b", "ab"},
		{"emoji dropped", "nota fiscal 🚀 emitida", "nota fiscal  emitida"},
		{"char above latin1 dropped", "texto – com travessão", "texto  com travessão"},
		{"only newlines", "\n\n\n", ""},
		{"only forbidden chars", "\x00\x01", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xml.SanitizeSchemaString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeSchemaString(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if got != "" && !tStringPattern.MatchString(got) {
				t.Errorf("SanitizeSchemaString(%q) = %q violates the TString pattern", tt.input, got)
			}
		})
	}
}
