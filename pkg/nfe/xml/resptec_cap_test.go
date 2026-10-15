package xml_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/xml"
)

// TestBuilder_RespTecXContatoCapped pins the silent truncation of xContato
// at 60 runes (TString maxLength): the full 72-char vendor razão social was
// rejected by SEFAZ with cvc-type.3.1.3 on the first homolog attempt, and
// the builder is the single enforcement point for any RespTecData source.
func TestBuilder_RespTecXContatoCapped(t *testing.T) {
	const longContact = "VAGNER SIMONE MARTINS FILHO CONSULTORIA EM TECNOLOGIA DA INFORMACAO LTDA"
	if utf8.RuneCountInString(longContact) <= 60 {
		t.Fatalf("fixture xContato must exceed 60 runes to exercise the cap")
	}

	input := minimalInvoiceInput()
	input.RespTec = &entity.RespTecData{
		CNPJ:     "47962538000140",
		XContato: longContact,
		Email:    "vagner.filhopro@gmail.com",
		Fone:     "66996615930",
	}

	doc, err := xml.NewBuilder().Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	respTec := findElement(t, doc.Root(), "./infNFe/infRespTec")
	assertChildOrder(t, respTec, "infRespTec children", []string{"CNPJ", "xContato", "email", "fone"})

	got := respTec.SelectElement("xContato").Text()
	if n := utf8.RuneCountInString(got); n != 60 {
		t.Errorf("xContato = %d runes (%q), want exactly 60", n, got)
	}
	if !strings.HasPrefix(longContact, got) {
		t.Errorf("xContato = %q, want a prefix of the source value", got)
	}
}
