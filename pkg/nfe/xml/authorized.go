package xml

import (
	"fmt"

	"github.com/beevik/etree"
)

// BuildAuthorizedXML wraps a signed <NFe> document inside the standard SEFAZ
// <nfeProc> envelope, appending a <protNFe> with the authorization protocol
// data returned by SEFAZ. This is the canonical format for an authorized NF-e
// and is what should be stored in xml_authorized.
//
// The signedNFeXML parameter must be a complete <NFe>...</NFe> document.
// The protocol parameters come from the SEFAZ Autorizacao/Consulta response.
func BuildAuthorizedXML(signedNFeXML, chNFe, nProt, dhRecbto, cStat, xMotivo string) (string, error) {
	// Parse the signed NFe to extract it as a subtree
	nfeDoc := etree.NewDocument()
	if err := nfeDoc.ReadFromString(signedNFeXML); err != nil {
		return "", fmt.Errorf("failed to parse signed NFe XML: %w", err)
	}

	nfeRoot := nfeDoc.Root()
	if nfeRoot == nil || nfeRoot.Tag != "NFe" {
		return "", fmt.Errorf("expected <NFe> root element, got <%s>", elementTagOrEmpty(nfeRoot))
	}

	// Build the nfeProc wrapper
	proc := etree.NewElement("nfeProc")
	proc.CreateAttr("xmlns", nfeNamespace)
	proc.CreateAttr("versao", "4.00")

	// Append the entire signed NFe as a child
	proc.AddChild(nfeRoot.Copy())

	// Build protNFe > infProt
	protNFe := proc.CreateElement("protNFe")
	protNFe.CreateAttr("versao", "4.00")

	infProt := protNFe.CreateElement("infProt")
	infProt.CreateAttr("Id", "") // SEFAZ may include an Id; empty is acceptable
	if chNFe != "" {
		infProt.CreateElement("chNFe").SetText(chNFe)
	}
	if dhRecbto != "" {
		infProt.CreateElement("dhRecbto").SetText(dhRecbto)
	}
	if nProt != "" {
		infProt.CreateElement("nProt").SetText(nProt)
	}
	if cStat != "" {
		infProt.CreateElement("cStat").SetText(cStat)
	}
	if xMotivo != "" {
		infProt.CreateElement("xMotivo").SetText(xMotivo)
	}

	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true
	doc.SetRoot(proc)

	result, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize nfeProc: %w", err)
	}
	return result, nil
}

func elementTagOrEmpty(el *etree.Element) string {
	if el == nil {
		return ""
	}
	return el.Tag
}
