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
