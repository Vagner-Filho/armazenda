package xml_test

import (
	"strings"
	"testing"
	"time"

	"armazenda/pkg/nfe/xml"
)

func validCancelInput() xml.CancelEventInput {
	return xml.CancelEventInput{
		AccessKey:     "51250312345678000190550010000001231234567890",
		Protocol:      "151250123456789",
		Justification: "Emissão com dados incorretos do destinatário",
		EmitterDoc:    "12345678000190",
		EmitterType:   1,
		EmitterUF:     "MT",
		Environment:   2,
		DhEvento:      time.Date(2025, 3, 15, 10, 31, 0, 0, time.FixedZone("AMT", -4*3600)),
		SeqEvento:     1,
	}
}

func TestBuildCancellationEvent(t *testing.T) {
	doc, err := xml.BuildCancellationEvent(validCancelInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	str, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	checks := []string{
		`<envEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"`,
		`<idLote>1</idLote>`,
		`Id="ID1101115125031234567800019055001000000123123456789001"`,
		`<cOrgao>51</cOrgao>`,
		`<tpAmb>2</tpAmb>`,
		`<CNPJ>12345678000190</CNPJ>`,
		`<chNFe>51250312345678000190550010000001231234567890</chNFe>`,
		`<tpEvento>110111</tpEvento>`,
		`<nSeqEvento>1</nSeqEvento>`,
		`<verEvento>1.00</verEvento>`,
		`<descEvento>Cancelamento</descEvento>`,
		`<nProt>151250123456789</nProt>`,
		`<xJust>Emissão com dados incorretos do destinatário</xJust>`,
	}
	for _, want := range checks {
		if !strings.Contains(str, want) {
			t.Errorf("expected XML to contain %q\nGot: %s", want, str)
		}
	}

	// dhEvento must carry the timezone offset
	if !strings.Contains(str, "<dhEvento>2025-03-15T10:31:00-04:00</dhEvento>") {
		t.Errorf("expected dhEvento with timezone, got: %s", str)
	}
}

func TestBuildCancellationEvent_CPF(t *testing.T) {
	input := validCancelInput()
	input.EmitterType = 2
	input.EmitterDoc = "12345678901"

	doc, err := xml.BuildCancellationEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, _ := doc.WriteToString()

	if !strings.Contains(str, "<CPF>12345678901</CPF>") {
		t.Errorf("expected CPF element for CPF emitter, got: %s", str)
	}
	if strings.Contains(str, "<CNPJ>") {
		t.Errorf("expected no CNPJ element for CPF emitter, got: %s", str)
	}
}

func TestBuildCancellationEvent_Production(t *testing.T) {
	input := validCancelInput()
	input.Environment = 1

	doc, err := xml.BuildCancellationEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, _ := doc.WriteToString()

	if !strings.Contains(str, "<tpAmb>1</tpAmb>") {
		t.Errorf("expected tpAmb=1 for production, got: %s", str)
	}
}

func TestBuildCancellationEvent_Validation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*xml.CancelEventInput)
	}{
		{"short access key", func(i *xml.CancelEventInput) { i.AccessKey = "123" }},
		{"empty protocol", func(i *xml.CancelEventInput) { i.Protocol = "" }},
		{"short justification", func(i *xml.CancelEventInput) { i.Justification = "curto" }},
		{"long justification", func(i *xml.CancelEventInput) { i.Justification = strings.Repeat("a", 257) }},
		{"invalid UF", func(i *xml.CancelEventInput) { i.EmitterUF = "XX" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCancelInput()
			tt.mutate(&input)
			if _, err := xml.BuildCancellationEvent(input); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestBuildCancellationEvent_SanitizesJustification(t *testing.T) {
	input := validCancelInput()
	input.Justification = "Emissão com dados incorretos\ndo destinatário 🚀"

	doc, err := xml.BuildCancellationEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, _ := doc.WriteToString()

	want := "<xJust>Emissão com dados incorretos; do destinatário</xJust>"
	if !strings.Contains(str, want) {
		t.Errorf("expected sanitized xJust %q, got: %s", want, str)
	}
	if strings.Contains(str, "🚀") {
		t.Error("characters above U+00FF must be dropped from xJust")
	}
}

func TestBuildCancellationEvent_JustificationTooShortAfterSanitize(t *testing.T) {
	input := validCancelInput()
	// 15+ runes raw, but the emoji-only second line is dropped, leaving < 15.
	input.Justification = "motivo curto\n🚀🚀🚀🚀🚀🚀🚀🚀"

	if _, err := xml.BuildCancellationEvent(input); err == nil {
		t.Fatal("expected validation error for justification too short after sanitize, got nil")
	}
}

func TestBuildCancellationEvent_DefaultSeqEvento(t *testing.T) {
	input := validCancelInput()
	input.SeqEvento = 0

	doc, err := xml.BuildCancellationEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, _ := doc.WriteToString()

	if !strings.Contains(str, `Id="ID1101115125031234567800019055001000000123123456789001"`) {
		t.Errorf("expected nSeqEvento to default to 1, got: %s", str)
	}
}
