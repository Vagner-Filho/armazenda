package xml_test

import (
	"strings"
	"testing"

	"armazenda/pkg/nfe/xml"
)

func TestBuildAuthorizedXML(t *testing.T) {
	signedNFe := `<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00"><ide><serie>1</serie></ide></infNFe></NFe>`

	result, err := xml.BuildAuthorizedXML(signedNFe, "51250312345678000190550010000001231234567890", "351250123456789", "2025-03-15T10:31:00-03:00", "100", "Autorizado o uso da NF-e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be wrapped in <nfeProc>
	if !strings.HasPrefix(result, `<?xml`) && !strings.Contains(result, "<nfeProc") {
		t.Fatal("expected <nfeProc> root element")
	}
	if !strings.Contains(result, "<nfeProc") {
		t.Fatal("result must contain <nfeProc>")
	}
	if !strings.Contains(result, "<NFe") {
		t.Fatal("result must contain the signed <NFe>")
	}
	if !strings.Contains(result, "<protNFe") {
		t.Fatal("result must contain <protNFe>")
	}
	if !strings.Contains(result, "351250123456789") {
		t.Fatal("result must contain the protocol number")
	}
	if !strings.Contains(result, "2025-03-15T10:31:00-03:00") {
		t.Fatal("result must contain dhRecbto")
	}
	if !strings.Contains(result, "<cStat>100</cStat>") {
		t.Fatal("result must contain cStat")
	}
	if !strings.Contains(result, "Autorizado o uso da NF-e") {
		t.Fatal("result must contain xMotivo")
	}
}

func TestBuildAuthorizedXML_InvalidInput(t *testing.T) {
	// Not a valid NFe document
	_, err := xml.BuildAuthorizedXML("not xml", "", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid XML input")
	}
}

func TestBuildAuthorizedXML_WrongRoot(t *testing.T) {
	// Root element is not <NFe>
	wrongRoot := `<nfeProc xmlns="http://www.portalfiscal.inf.br/nfe"></nfeProc>`
	_, err := xml.BuildAuthorizedXML(wrongRoot, "", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for non-NFe root element")
	}
}

func TestBuildAuthorizedXML_RoundTripWithParser(t *testing.T) {
	// Build an authorized XML and verify the DANFE parser can extract the
	// protocol fields from it.
	signedNFe := `<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00"><ide><serie>1</serie><nNF>123</nNF><tpEmis>1</tpEmis><tpAmb>2</tpAmb></ide></infNFe></NFe>`

	authXML, err := xml.BuildAuthorizedXML(signedNFe, "51250312345678000190550010000001231234567890", "351250123456789", "2025-03-15T10:31:00-03:00", "100", "Autorizado")
	if err != nil {
		t.Fatalf("BuildAuthorizedXML error: %v", err)
	}

	data, err := xml.ParseDANFEData(authXML)
	if err != nil {
		t.Fatalf("ParseDANFEData error: %v", err)
	}

	if data.Protocol != "351250123456789" {
		t.Errorf("Protocol = %s, want '351250123456789'", data.Protocol)
	}
	if data.CStat != "100" {
		t.Errorf("CStat = %s, want '100'", data.CStat)
	}
	if data.XMotivo != "Autorizado" {
		t.Errorf("XMotivo = %s, want 'Autorizado'", data.XMotivo)
	}
	if !strings.Contains(data.ProtocolDate, "15/03/2025") {
		t.Errorf("ProtocolDate = %s, want to contain '15/03/2025'", data.ProtocolDate)
	}
}
