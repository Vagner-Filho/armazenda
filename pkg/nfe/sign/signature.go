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
	// Use default namespace (no prefix) for XMLDSIG elements to avoid
	// SEFAZ rejection 587: "Usar somente o namespace padrao da NF-e"
	ctx.Prefix = ""

	signedElement, err := ctx.SignEnveloped(element)
	if err != nil {
		return nil, fmt.Errorf("failed to sign element: %w", err)
	}

	return signedElement, nil
}

// SignDocument signs the <infNFe> element and moves the resulting <Signature>
// to be a sibling of <infNFe> under <NFe>, as required by the NF-e 4.00 schema.
func (s *Signer) SignDocument(doc *etree.Document) error {
	// Find the <NFe> root and <infNFe> to sign
	root := doc.Root()
	if root == nil {
		return fmt.Errorf("document has no root element")
	}

	infNFe := root.FindElement("//infNFe")
	if infNFe == nil {
		return fmt.Errorf("infNFe element not found")
	}

	// Sign enveloped — this places <ds:Signature> inside <infNFe>
	signedElement, err := s.SignEnveloped(infNFe)
	if err != nil {
		return err
	}

	// Extract the <Signature> from inside the signed <infNFe>
	// The NF-e 4.00 schema requires Signature as a direct child of <NFe>, not inside <infNFe>
	// We search without prefix since XMLDSIG now uses default namespace (Prefix="")
	signature := signedElement.FindElement("//Signature")

	if signature != nil {
		// Create a detached copy of the signature to place outside <infNFe>
		// We must copy because RemoveChild+AddChild with the same element
		// can cause etree internal issues.
		sigCopy := signature.Copy()
		// Remove the signature from inside <infNFe>.
		// goxmldsig's SignEnveloped appends the signature directly to ret.Child
		// without calling AddChild(), so Parent() is nil. RemoveChild checks
		// Parent() and silently fails; use RemoveChildAt(index) instead.
		for i, token := range signedElement.Child {
			if el, ok := token.(*etree.Element); ok {
				if el.Tag == "Signature" {
					signedElement.RemoveChildAt(i)
					break
				}
			}
		}
		// Replace the original <infNFe> with the cleaned one
		root.RemoveChild(infNFe)
		root.AddChild(signedElement)
		// Add the signature copy as a sibling after <infNFe>
		root.AddChild(sigCopy)
	} else {
		// No signature found — replace with signed element as-is
		root.RemoveChild(infNFe)
		root.AddChild(signedElement)
	}

	return nil
}
