package sign

import (
	"crypto"
	"fmt"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

// Signer handles XML digital signature for NF-e.
type Signer struct {
	cert *Certificate
}

// NewSigner creates a new XML signer with the given certificate.
func NewSigner(cert *Certificate) *Signer {
	return &Signer{cert: cert}
}

// SignEnveloped signs an XML element using enveloped signature.
// This is the standard signature method for NF-e.
func (s *Signer) SignEnveloped(element *etree.Element) (*etree.Element, error) {
	ctx, err := dsig.NewSigningContext(s.cert.PrivateKey, s.cert.RawChain())
	if err != nil {
		return nil, fmt.Errorf("failed to create signing context: %w", err)
	}

	// Configure for NF-e requirements
	ctx.Hash = crypto.SHA1
	if err := ctx.SetSignatureMethod(dsig.RSASHA1SignatureMethod); err != nil {
		return nil, fmt.Errorf("failed to set signature method: %w", err)
	}
	ctx.Canonicalizer = dsig.MakeC14N10RecCanonicalizer()
	ctx.IdAttribute = "Id"

	signedElement, err := ctx.SignEnveloped(element)
	if err != nil {
		return nil, fmt.Errorf("failed to sign element: %w", err)
	}

	return signedElement, nil
}

// SignDocument signs the entire XML document (the infNFe element).
func (s *Signer) SignDocument(doc *etree.Document) error {
	// Find the infNFe element to sign
	root := doc.Root()
	if root == nil {
		return fmt.Errorf("document has no root element")
	}

	infNFe := root.FindElement("//infNFe")
	if infNFe == nil {
		return fmt.Errorf("infNFe element not found")
	}

	signedElement, err := s.SignEnveloped(infNFe)
	if err != nil {
		return err
	}

	// Replace the original infNFe with the signed one
	root.RemoveChild(infNFe)
	root.AddChild(signedElement)

	return nil
}
