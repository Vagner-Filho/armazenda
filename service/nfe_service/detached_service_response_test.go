package nfe_service

import (
	"errors"
	"strings"
	"testing"

	// Internal - Entities
	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/sefaz"
	nfe_xml "armazenda/pkg/nfe/xml"
)

const testSignedNFe = `<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00"><ide><serie>1</serie><nNF>123</nNF><tpEmis>1</tpEmis><tpAmb>2</tpAmb></ide></infNFe></NFe>`

const testAccessKey = "51250312345678000190550010000001231234567890"

// detachedStatusCall records one UpdateDetachedInvoiceStatus invocation.
type detachedStatusCall struct {
	id          int
	status      string
	protocol    string
	sefazCode   string
	sefazMotive string
}

// detachedAuthXMLCall records one successful UpdateDetachedInvoiceAuthorizedXML invocation.
type detachedAuthXMLCall struct {
	id  int
	xml string
}

// stubDetachedResponseModel satisfies detachedResponseModel and records calls
// so handleDetachedSefazResponse can be tested without a database.
type stubDetachedResponseModel struct {
	statusCalls  []detachedStatusCall
	authXMLCalls []detachedAuthXMLCall
	authXMLErr   error
}

func (s *stubDetachedResponseModel) UpdateDetachedInvoiceStatus(id int, status, protocol, sefazCode, sefazMotive string) error {
	s.statusCalls = append(s.statusCalls, detachedStatusCall{id: id, status: status, protocol: protocol, sefazCode: sefazCode, sefazMotive: sefazMotive})
	return nil
}

func (s *stubDetachedResponseModel) UpdateDetachedInvoiceAuthorizedXML(id int, xmlAuthorized string) error {
	if s.authXMLErr != nil {
		return s.authXMLErr
	}
	s.authXMLCalls = append(s.authXMLCalls, detachedAuthXMLCall{id: id, xml: xmlAuthorized})
	return nil
}

func TestHandleDetachedSefazResponse_Authorized_StoresNfeProcXML(t *testing.T) {
	stub := &stubDetachedResponseModel{}
	svc := NewNFeService()
	resp := &sefaz.AutorizacaoResponse{
		StatusCode:   "100",
		StatusMotive: "Autorizado o uso da NF-e",
		Protocol:     "351250123456789",
		DhRecbto:     "2025-03-15T10:31:00-03:00",
		AccessKey:    testAccessKey,
	}

	result, toast := svc.handleDetachedSefazResponse(resp, 7, testSignedNFe, testAccessKey, stub)

	if len(stub.statusCalls) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(stub.statusCalls))
	}
	call := stub.statusCalls[0]
	if call.id != 7 || call.status != "authorized" || call.protocol != "351250123456789" || call.sefazCode != "100" {
		t.Errorf("unexpected status call: %+v", call)
	}

	if len(stub.authXMLCalls) != 1 {
		t.Fatalf("expected 1 authorized XML update, got %d", len(stub.authXMLCalls))
	}
	if stub.authXMLCalls[0].id != 7 {
		t.Errorf("authorized XML updated for id %d, want 7", stub.authXMLCalls[0].id)
	}

	// The stored wrapper must round-trip through the DANFE parser with the
	// protocol fields intact (AC1/AC4).
	data, parseErr := nfe_xml.ParseDANFEData(stub.authXMLCalls[0].xml)
	if parseErr != nil {
		t.Fatalf("ParseDANFEData on stored xml_authorized: %v", parseErr)
	}
	if data.Protocol != "351250123456789" {
		t.Errorf("Protocol = %s, want '351250123456789'", data.Protocol)
	}
	if data.CStat != "100" {
		t.Errorf("CStat = %s, want '100'", data.CStat)
	}
	if data.XMotivo != "Autorizado o uso da NF-e" {
		t.Errorf("XMotivo = %s, want 'Autorizado o uso da NF-e'", data.XMotivo)
	}
	if !strings.Contains(data.ProtocolDate, "15/03/2025") {
		t.Errorf("ProtocolDate = %s, want to contain '15/03/2025'", data.ProtocolDate)
	}

	// The emitted result keeps the signed XML, not the wrapper.
	if result.XML != testSignedNFe {
		t.Error("result XML should remain the signed XML")
	}
	if result.AccessKey != testAccessKey {
		t.Errorf("result AccessKey = %s, want %s", result.AccessKey, testAccessKey)
	}
	if toast.Type != entity_public.SuccessToast {
		t.Errorf("toast type = %v, want SuccessToast", toast.Type)
	}
}

func TestHandleDetachedSefazResponse_Authorized_BuildFailureNonFatal(t *testing.T) {
	t.Run("BuildAuthorizedXMLFails", func(t *testing.T) {
		stub := &stubDetachedResponseModel{}
		svc := NewNFeService()
		resp := &sefaz.AutorizacaoResponse{
			StatusCode:   "100",
			StatusMotive: "Autorizado o uso da NF-e",
			Protocol:     "351250123456789",
			DhRecbto:     "2025-03-15T10:31:00-03:00",
		}

		result, toast := svc.handleDetachedSefazResponse(resp, 7, "not xml", testAccessKey, stub)

		if len(stub.statusCalls) != 1 || stub.statusCalls[0].status != "authorized" {
			t.Errorf("status must still be updated to authorized, got %+v", stub.statusCalls)
		}
		if len(stub.authXMLCalls) != 0 {
			t.Errorf("no authorized XML should be stored when build fails, got %d calls", len(stub.authXMLCalls))
		}
		if toast.Type != entity_public.SuccessToast {
			t.Errorf("toast type = %v, want SuccessToast", toast.Type)
		}
		if result.XML != "not xml" {
			t.Error("result XML should be unchanged on build failure")
		}
	})

	t.Run("DBUpdateFails", func(t *testing.T) {
		stub := &stubDetachedResponseModel{authXMLErr: errors.New("db down")}
		svc := NewNFeService()
		resp := &sefaz.AutorizacaoResponse{
			StatusCode:   "100",
			StatusMotive: "Autorizado o uso da NF-e",
			Protocol:     "351250123456789",
			DhRecbto:     "2025-03-15T10:31:00-03:00",
		}

		_, toast := svc.handleDetachedSefazResponse(resp, 7, testSignedNFe, testAccessKey, stub)

		if len(stub.statusCalls) != 1 || stub.statusCalls[0].status != "authorized" {
			t.Errorf("status must still be updated to authorized, got %+v", stub.statusCalls)
		}
		if toast.Type != entity_public.SuccessToast {
			t.Errorf("toast type = %v, want SuccessToast", toast.Type)
		}
	})
}

func TestHandleDetachedSefazResponse_NonAuthorized_DoesNotStoreXML(t *testing.T) {
	cases := []struct {
		name       string
		code       string
		motive     string
		wantStatus string
	}{
		{name: "Processing", code: "105", motive: "Lote em processamento", wantStatus: "pending"},
		{name: "Rejected", code: "999", motive: "Rejeicao: Erro nao catalogado", wantStatus: "denied"},
		{name: "Unknown", code: "", motive: "Resposta vazia", wantStatus: "pending"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubDetachedResponseModel{}
			svc := NewNFeService()
			resp := &sefaz.AutorizacaoResponse{StatusCode: tc.code, StatusMotive: tc.motive}

			_, _ = svc.handleDetachedSefazResponse(resp, 7, testSignedNFe, testAccessKey, stub)

			if len(stub.authXMLCalls) != 0 {
				t.Errorf("no authorized XML should be stored for status %q, got %d calls", tc.code, len(stub.authXMLCalls))
			}
			if len(stub.statusCalls) != 1 {
				t.Fatalf("expected 1 status update, got %d", len(stub.statusCalls))
			}
			if stub.statusCalls[0].status != tc.wantStatus {
				t.Errorf("status = %s, want %s", stub.statusCalls[0].status, tc.wantStatus)
			}
		})
	}
}
