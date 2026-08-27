package xml

import (
	"github.com/beevik/etree"
)

const (
	envelopeNS = "http://www.w3.org/2003/05/soap-envelope"
)

// BuildSOAPEnvelope wraps a raw XML payload inside a SOAP envelope, matching the
// structure used by production NF-e libraries (PyNFe, Zeus DFe, etc.).
//
// The envelope contains only a <soap:Body> with a single <nfeDadosMsg> element.
// The <nfeDadosMsg> carries the service namespace, and the raw payload is parsed
// and appended as its child. No <soap:Header> or <nfeCabecMsg> is included.
func BuildSOAPEnvelope(serviceNamespace, bodyContent string) string {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	env := etree.NewElement("soap12:Envelope")
	env.CreateAttr("xmlns:soap12", envelopeNS)
	env.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	env.CreateAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")

	body := env.CreateElement("soap12:Body")

	nfeDadosMsg := body.CreateElement("nfeDadosMsg")
	nfeDadosMsg.CreateAttr("xmlns", serviceNamespace)

	// Parse the body content and add it as a child of <nfeDadosMsg>
	bodyDoc := etree.NewDocument()
	if err := bodyDoc.ReadFromString(bodyContent); err == nil {
		nfeDadosMsg.AddChild(bodyDoc.Root())
	} else {
		// Fallback: wrap raw content
		nfeDadosMsg.SetText(bodyContent)
	}

	doc.SetRoot(env)
	str, _ := doc.WriteToString()
	return str
}

// BuildSOAPEnvelopeWithCabecMsg builds a SOAP 1.1 envelope with a <soap:Header>
// containing <nfeCabecMsg> per MOC 7.0 §3.3. Used for web services that
// require the SOAP header (e.g. RecepcaoEvento4 in MT homolog).
//
// cUF is the IBGE state code (e.g. "51" for MT). versaoDados is the data
// layout version ("1.00" for events, "4.00" for NF-e/NFC-e).
//
// The existing BuildSOAPEnvelope is left untouched so the other four flows
// (authorization, query, status, SVC status) are not affected by this PoC.
// Per-UF envelope differences, when discovered, should be handled by the
// caller — this function takes everything as explicit parameters and has no
// lookup tables or global state.
func BuildSOAPEnvelopeWithCabecMsg(serviceNamespace, bodyContent, cUF, versaoDados string) string {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	env := etree.NewElement("soap:Envelope")
	env.CreateAttr("xmlns:soap", envelopeNS)
	env.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	env.CreateAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")

	header := env.CreateElement("soap:Header")
	cabec := header.CreateElement("nfeCabecMsg")
	cabec.CreateAttr("xmlns", serviceNamespace)
	cabec.CreateElement("cUF").SetText(cUF)
	cabec.CreateElement("versaoDados").SetText(versaoDados)

	body := env.CreateElement("soap:Body")
	nfeDadosMsg := body.CreateElement("nfeDadosMsg")
	nfeDadosMsg.CreateAttr("xmlns", serviceNamespace)

	bodyDoc := etree.NewDocument()
	if err := bodyDoc.ReadFromString(bodyContent); err == nil {
		nfeDadosMsg.AddChild(bodyDoc.Root())
	} else {
		// Fallback: wrap raw content
		nfeDadosMsg.SetText(bodyContent)
	}

	doc.SetRoot(env)
	str, _ := doc.WriteToString()
	return str
}
