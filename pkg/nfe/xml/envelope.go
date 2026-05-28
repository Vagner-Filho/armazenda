package xml

import (
	"github.com/beevik/etree"
)

const (
	envelopeNS = "http://www.w3.org/2003/05/soap-envelope"
)

// BuildSOAPEnvelope wraps a raw XML body inside a SOAP envelope.
// cUF is the 2-digit IBGE state code (e.g., "51" for MT, "35" for SP).
// serviceNamespace is the SOAP namespace for the specific SEFAZ service.
func BuildSOAPEnvelope(cUF, serviceNamespace, bodyContent string) string {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	env := etree.NewElement("soap12:Envelope")
	env.CreateAttr("xmlns:soap12", envelopeNS)
	env.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	env.CreateAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")

	head := env.CreateElement("soap12:Header")
	nfeCabMsg := head.CreateElement("nfeCabecMsg")
	nfeCabMsg.CreateAttr("xmlns", serviceNamespace)
	versaoDados := nfeCabMsg.CreateElement("versaoDados")
	versaoDados.SetText("4.00")
	cufElem := nfeCabMsg.CreateElement("cUF")
	cufElem.SetText(cUF)

	body := env.CreateElement("soap12:Body")

	// Parse the body content and add it
	bodyDoc := etree.NewDocument()
	if err := bodyDoc.ReadFromString(bodyContent); err == nil {
		body.AddChild(bodyDoc.Root())
	} else {
		// Fallback: create raw element
		nfeDadosMsg := body.CreateElement("nfeDadosMsg")
		nfeDadosMsg.CreateAttr("xmlns", serviceNamespace)
		nfeDadosMsg.SetText(bodyContent)
	}

	doc.SetRoot(env)
	str, _ := doc.WriteToString()
	return str
}
