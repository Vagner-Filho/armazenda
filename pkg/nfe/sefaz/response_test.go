package sefaz_test

import (
	"testing"

	"armazenda/pkg/nfe/sefaz"
)

func TestParseEventoResponse_Registered(t *testing.T) {
	body := []byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<nfeResultMsg xmlns="http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4">
			<retEnvEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00">
				<idLote>1</idLote>
				<tpAmb>2</tpAmb>
				<verAplic>MT-1.0.0</verAplic>
				<cOrgao>51</cOrgao>
				<cStat>128</cStat>
				<xMotivo>Lote de Evento Processado</xMotivo>
				<retEvento versao="1.00">
					<infEvento>
						<tpAmb>2</tpAmb>
						<verAplic>MT-1.0.0</verAplic>
						<cOrgao>51</cOrgao>
						<cStat>135</cStat>
						<xMotivo>Evento registrado e vinculado a NF-e</xMotivo>
						<chNFe>51250312345678000190550010000001231234567890</chNFe>
						<tpEvento>110111</tpEvento>
						<xEvento>Cancelamento</xEvento>
						<nSeqEvento>1</nSeqEvento>
						<dhRegEvento>2025-03-15T10:35:00-04:00</dhRegEvento>
						<nProt>151250987654321</nProt>
					</infEvento>
				</retEvento>
			</retEnvEvento>
		</nfeResultMsg>
	</soap:Body>
</soap:Envelope>`)

	resp, err := sefaz.ParseEventoResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsRegistered() {
		t.Fatalf("expected IsRegistered, got cStat=%q motive=%q", resp.StatusCode, resp.StatusMotive)
	}
	if resp.IsAlreadyCancelled() {
		t.Fatal("did not expect IsAlreadyCancelled for cStat 135")
	}
	if resp.Protocol != "151250987654321" {
		t.Errorf("expected event protocol 151250987654321, got %q", resp.Protocol)
	}
	if resp.AccessKey != "51250312345678000190550010000001231234567890" {
		t.Errorf("unexpected access key: %q", resp.AccessKey)
	}
	if resp.DhRegEvento != "2025-03-15T10:35:00-04:00" {
		t.Errorf("unexpected dhRegEvento: %q", resp.DhRegEvento)
	}
}

func TestParseEventoResponse_AlreadyCancelled(t *testing.T) {
	body := []byte(`<retEnvEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00">
	<cStat>128</cStat>
	<xMotivo>Lote de Evento Processado</xMotivo>
	<retEvento versao="1.00">
		<infEvento>
			<cStat>218</cStat>
			<xMotivo>Rejeição: NF-e já está cancelada na base de dados da SEFAZ</xMotivo>
			<chNFe>51250312345678000190550010000001231234567890</chNFe>
		</infEvento>
	</retEvento>
</retEnvEvento>`)

	resp, err := sefaz.ParseEventoResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsAlreadyCancelled() {
		t.Fatalf("expected IsAlreadyCancelled, got cStat=%q", resp.StatusCode)
	}
	if resp.IsRegistered() {
		t.Fatal("did not expect IsRegistered for cStat 218")
	}
}

func TestParseEventoResponse_Rejection(t *testing.T) {
	body := []byte(`<retEnvEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00">
	<cStat>128</cStat>
	<xMotivo>Lote de Evento Processado</xMotivo>
	<retEvento versao="1.00">
		<infEvento>
			<cStat>501</cStat>
			<xMotivo>Rejeição: Prazo de cancelamento superior ao previsto na Legislação</xMotivo>
			<chNFe>51250312345678000190550010000001231234567890</chNFe>
		</infEvento>
	</retEvento>
</retEnvEvento>`)

	resp, err := sefaz.ParseEventoResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsRegistered() || resp.IsAlreadyCancelled() {
		t.Fatalf("expected rejection to be neither registered nor already-cancelled, got cStat=%q", resp.StatusCode)
	}
	if resp.StatusCode != "501" {
		t.Errorf("expected cStat 501, got %q", resp.StatusCode)
	}
	if resp.StatusMotive == "" {
		t.Error("expected rejection motive to be populated")
	}
}

func TestParseEventoResponse_BatchLevelRejection(t *testing.T) {
	// When the batch itself is rejected (e.g., schema error), there is no
	// retEvento — only the top-level retEnvEvento cStat.
	body := []byte(`<retEnvEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00">
	<cStat>491</cStat>
	<xMotivo>Rejeição: O tpEvento informado inválido</xMotivo>
</retEnvEvento>`)

	resp, err := sefaz.ParseEventoResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != "491" {
		t.Errorf("expected fallback to top-level cStat 491, got %q", resp.StatusCode)
	}
	if resp.IsRegistered() || resp.IsAlreadyCancelled() {
		t.Fatalf("expected no success flags for cStat %q", resp.StatusCode)
	}
}

func TestParseEventoResponse_InvalidXML(t *testing.T) {
	if _, err := sefaz.ParseEventoResponse([]byte("<retEnvEvento><unclosed")); err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
}
