package sefaz

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// AutorizacaoResponse holds the parsed response from NFeAutorizacao4.
type AutorizacaoResponse struct {
	StatusCode    string // cStat
	StatusMotive  string // xMotivo
	Protocol      string // nProt (if available)
	ReceiptNumber string // nRec (for batch tracking)
	AccessKey     string // chNFe (if available)
	DhRecbto      string // dhRecbto (date/time of receipt by SEFAZ)
}

// IsAuthorized returns true if the NF-e was authorized.
func (r *AutorizacaoResponse) IsAuthorized() bool {
	return r.StatusCode == "100"
}

// IsProcessing returns true if the batch is being processed.
func (r *AutorizacaoResponse) IsProcessing() bool {
	return r.StatusCode == "103" || r.StatusCode == "105"
}

// IsRejected returns true if the NF-e was rejected.
func (r *AutorizacaoResponse) IsRejected() bool {
	code := r.StatusCode
	// Rejection codes are >= 200 or specific non-success codes
	return code != "" && code != "100" && code != "103" && code != "104" && code != "105" && code != "107"
}

// IsAccepted returns true if the batch was received successfully.
func (r *AutorizacaoResponse) IsAccepted() bool {
	return r.StatusCode == "103" || r.StatusCode == "104"
}

// ParseAutorizacaoResponse parses the SOAP response from NFeAutorizacao4.
func ParseAutorizacaoResponse(soapBody []byte) (*AutorizacaoResponse, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(soapBody); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP response: %w", err)
	}

	resp := &AutorizacaoResponse{}

	// Try to find retEnviNFe
	retEnvi := doc.FindElement("//retEnviNFe")
	if retEnvi == nil {
		// Fallback: look directly in Body
		retEnvi = doc.FindElement("//Body/retEnviNFe")
	}

	if retEnvi != nil {
		// For indSinc=1 responses, the actual per-NF-e result is inside <protNFe><infProt>
		protNFe := retEnvi.FindElement("protNFe")
		if protNFe != nil {
			infProt := protNFe.FindElement("infProt")
			if infProt != nil {
				resp.StatusCode = getElementText(infProt, "cStat")
				resp.StatusMotive = getElementText(infProt, "xMotivo")
				resp.Protocol = getElementText(infProt, "nProt")
				resp.AccessKey = getElementText(infProt, "chNFe")
				resp.DhRecbto = getElementText(infProt, "dhRecbto")
			}
		}

		// If no <protNFe> found (indSinc=0 or async), use top-level cStat
		if resp.StatusCode == "" {
			resp.StatusCode = getElementText(retEnvi, "cStat")
			resp.StatusMotive = getElementText(retEnvi, "xMotivo")
		}

		// Try to find receipt number
		infRec := retEnvi.FindElement("infRec")
		if infRec != nil {
			resp.ReceiptNumber = getElementText(infRec, "nRec")
		}
	} else {
		// Fallback: try retConsReciNFe (sometimes the response comes as query result)
		retCons := doc.FindElement("//retConsReciNFe")
		if retCons != nil {
			resp.StatusCode = getElementText(retCons, "cStat")
			resp.StatusMotive = getElementText(retCons, "xMotivo")

			// Look for protocol inside protNFe
			protNFe := retCons.FindElement("protNFe")
			if protNFe != nil {
				infProt := protNFe.FindElement("infProt")
				if infProt != nil {
					resp.StatusCode = getElementText(infProt, "cStat")
					resp.StatusMotive = getElementText(infProt, "xMotivo")
					resp.Protocol = getElementText(infProt, "nProt")
					resp.AccessKey = getElementText(infProt, "chNFe")
					resp.DhRecbto = getElementText(infProt, "dhRecbto")
				}
			}
		}
	}

	return resp, nil
}

// StatusResponse holds the parsed response from NFeStatusServico4.
type StatusResponse struct {
	StatusCode   string
	StatusMotive string
}

// IsOperational returns true if the SEFAZ service is operational.
func (r *StatusResponse) IsOperational() bool {
	return r.StatusCode == "107"
}

// IsSVCOperational returns true if the SVC service is active (code 107 from SVC status check).
func (r *StatusResponse) IsSVCOperational() bool {
	return r.StatusCode == "107"
}

// IsSVCDeactivating returns true if the SVC is being deactivated (code 113).
func (r *StatusResponse) IsSVCDeactivating() bool {
	return r.StatusCode == "113"
}

// IsSVCDisabled returns true if the SVC has been disabled by the origin SEFAZ (code 114).
func (r *StatusResponse) IsSVCDisabled() bool {
	return r.StatusCode == "114"
}

// ParseStatusResponse parses the SOAP response from NFeStatusServico4.
func ParseStatusResponse(soapBody []byte) (*StatusResponse, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(soapBody); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP response: %w", err)
	}

	retStat := doc.FindElement("//retConsStatServ")
	if retStat == nil {
		retStat = doc.FindElement("//Body/retConsStatServ")
	}

	resp := &StatusResponse{}
	if retStat != nil {
		resp.StatusCode = getElementText(retStat, "cStat")
		resp.StatusMotive = getElementText(retStat, "xMotivo")
	}

	return resp, nil
}

// ConsultaResponse holds the parsed response from NFeConsultaProtocolo4.
type ConsultaResponse struct {
	StatusCode   string
	StatusMotive string
	Protocol     string
	AccessKey    string
	DhRecbto     string
}

// IsAuthorized returns true if the NF-e was authorized.
func (r *ConsultaResponse) IsAuthorized() bool {
	return r.StatusCode == "100"
}

// ParseConsultaResponse parses the SOAP response from NFeConsultaProtocolo4.
func ParseConsultaResponse(soapBody []byte) (*ConsultaResponse, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(soapBody); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP response: %w", err)
	}

	retCons := doc.FindElement("//retConsSitNFe")
	if retCons == nil {
		retCons = doc.FindElement("//Body/retConsSitNFe")
	}

	resp := &ConsultaResponse{}
	if retCons != nil {
		resp.StatusCode = getElementText(retCons, "cStat")
		resp.StatusMotive = getElementText(retCons, "xMotivo")

		protNFe := retCons.FindElement("protNFe")
		if protNFe != nil {
			infProt := protNFe.FindElement("infProt")
			if infProt != nil {
				resp.StatusCode = getElementText(infProt, "cStat")
				resp.StatusMotive = getElementText(infProt, "xMotivo")
				resp.Protocol = getElementText(infProt, "nProt")
				resp.AccessKey = getElementText(infProt, "chNFe")
				resp.DhRecbto = getElementText(infProt, "dhRecbto")
			}
		}
	}

	return resp, nil
}

// EventoResponse holds the parsed response from RecepcaoEvento4.
type EventoResponse struct {
	StatusCode   string // cStat (from retEvento/infEvento when available)
	StatusMotive string // xMotivo
	Protocol     string // nProt of the registered event
	AccessKey    string // chNFe
	DhRegEvento  string // dhRegEvento (date/time the event was registered)
}

// IsRegistered returns true if the event was registered and linked to the NF-e.
func (r *EventoResponse) IsRegistered() bool {
	return r.StatusCode == "135"
}

// IsAlreadyCancelled returns true if SEFAZ reports the NF-e is already
// cancelled (rejection 218). Callers may treat this as an idempotent success
// to reconcile local state with SEFAZ.
func (r *EventoResponse) IsAlreadyCancelled() bool {
	return r.StatusCode == "218"
}

// ParseEventoResponse parses the SOAP response from RecepcaoEvento4.
// The top-level retEnvEvento cStat is 128 (lote processado); the per-event
// result lives in retEvento/infEvento (135 = registered and linked).
func ParseEventoResponse(soapBody []byte) (*EventoResponse, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(soapBody); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP response: %w", err)
	}

	resp := &EventoResponse{}

	retEvento := doc.FindElement("//retEvento")
	if retEvento != nil {
		infEvento := retEvento.FindElement("infEvento")
		if infEvento != nil {
			resp.StatusCode = getElementText(infEvento, "cStat")
			resp.StatusMotive = getElementText(infEvento, "xMotivo")
			resp.Protocol = getElementText(infEvento, "nProt")
			resp.AccessKey = getElementText(infEvento, "chNFe")
			resp.DhRegEvento = getElementText(infEvento, "dhRegEvento")
		}
	} else {
		retEvento := doc.FindElement("//retInutNFe")
		if retEvento != nil {
			infEvento := retEvento.FindElement("infInut")
			if infEvento != nil {
				resp.StatusCode = getElementText(infEvento, "cStat")
				resp.StatusMotive = getElementText(infEvento, "xMotivo")
				resp.Protocol = getElementText(infEvento, "nProt")
				resp.AccessKey = getElementText(infEvento, "chNFe")
				resp.DhRegEvento = getElementText(infEvento, "dhRegEvento")
			}

		}
	}

	if resp.StatusCode == "" {
		// Fallback: top-level retEnvEvento cStat (e.g., 128, or a batch-level rejection)
		retEnv := doc.FindElement("//retEnvEvento")
		if retEnv == nil {
			retEnv = doc.FindElement("//Body/retEnvEvento")
		}
		if retEnv != nil {
			resp.StatusCode = getElementText(retEnv, "cStat")
			resp.StatusMotive = getElementText(retEnv, "xMotivo")
		}
	}

	return resp, nil
}

// getElementText returns the trimmed text of a child element, or empty string if not found.
func getElementText(parent *etree.Element, tag string) string {
	el := parent.FindElement(tag)
	if el == nil {
		return ""
	}
	return strings.TrimSpace(el.Text())
}
