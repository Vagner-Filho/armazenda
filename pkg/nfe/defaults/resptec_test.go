package defaults_test

import (
	"testing"

	"armazenda/pkg/nfe/defaults"
)

// TestRespTec verifies the re-enabled system-scoped responsável técnico
// block: the gate is on while all four 1-1 fields are present. The
// 60-char cap on xContato is enforced at the builder (truncation, no
// warning) — see pkg/nfe/xml/builder.go buildRespTec; the source constant
// may hold the full razão social.
func TestRespTec(t *testing.T) {
	if !defaults.RespTecConfigured() {
		t.Error("RespTecConfigured() = false with all four fields present, want true")
	}

	// Values kept for re-enable.
	if got, want := defaults.RespTec.CNPJ, "47962538000140"; got != want {
		t.Errorf("CNPJ = %q, want %q", got, want)
	}
	if len(defaults.RespTec.CNPJ) != 14 {
		t.Errorf("CNPJ must be 14 digits, got %d chars", len(defaults.RespTec.CNPJ))
	}
	for _, r := range defaults.RespTec.CNPJ {
		if r < '0' || r > '9' {
			t.Errorf("CNPJ contains non-digit %q", r)
			break
		}
	}
	// The source constant may exceed the 60-char TString limit: the
	// builder caps the emitted <xContato> at 60 runes (truncation), so
	// nothing here enforces the limit on the stored value.
	if got, want := defaults.RespTec.Email, "vagner.filhopro@gmail.com"; got != want {
		t.Errorf("Email = %q, want %q", got, want)
	}
	if got, want := defaults.RespTec.Fone, "66996615930"; got != want {
		t.Errorf("Fone = %q, want %q", got, want)
	}
}

// TestRespTecUnconfigured documents the gate contract for when the block is
// re-enabled: an incomplete vendor identity (missing any of the four 1-1
// fields) must keep the gate off, so the emission paths omit the block
// instead of emitting one SEFAZ rejects (e.g. blank email → 215 "One of
// 'email' is expected").
func TestRespTecUnconfigured(t *testing.T) {
	if !defaults.RespTecConfigured() {
		t.Skip("gate is globally disabled (see TestRespTec); nothing to assert until re-enabled")
	}

	saved := defaults.RespTec
	t.Cleanup(func() { defaults.RespTec = saved })

	defaults.RespTec.CNPJ = ""
	if defaults.RespTecConfigured() {
		t.Error("RespTecConfigured() = true with empty CNPJ, want false")
	}

	defaults.RespTec = defaults.SystemRespTecData{CNPJ: "47962538000140"}
	if defaults.RespTecConfigured() {
		t.Error("RespTecConfigured() = true with empty fone, want false")
	}

	defaults.RespTec = defaults.SystemRespTecData{
		CNPJ:     "47962538000140",
		XContato: "X",
		Email:    "",
		Fone:     "66996615930",
	}
	if defaults.RespTecConfigured() {
		t.Error("RespTecConfigured() = true with empty email, want false")
	}
}
