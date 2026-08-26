package xml

import (
	"fmt"
	"time"

	"armazenda/pkg/nfe/defaults"

	"github.com/beevik/etree"
)

// Event types and versions (MOC 7.0, NT 2014.002).
const (
	// TpEventoCancelamento is the NF-e cancellation event type.
	TpEventoCancelamento = "110111"
	// VersaoEvento is the layout version of the event envelope.
	VersaoEvento = "2v2"
	// Justification bounds per MOC (xJust: 15-256 characters).
	justificativaMinLen = 15
	justificativaMaxLen = 256
)

// CancelEventInput holds the data needed to build a cancellation event (110111).
type CancelEventInput struct {
	AccessKey     string    // 44-digit chNFe of the NF-e being cancelled
	Protocol      string    // nProt of the authorization being cancelled
	Justification string    // xJust (15-256 characters)
	EmitterDoc    string    // CNPJ or CPF of the emitter (author of the event)
	EmitterType   int       // 1=CNPJ, 2=CPF (matches entity.EmitterData.Type)
	EmitterUF     string    // UF used for cOrgao
	Environment   int       // 1=production, 2=homologation
	DhEvento      time.Time // event timestamp
	SeqEvento     int       // nSeqEvento (1 for a first cancellation)
}

// BuildCancellationEvent builds an unsigned <envEvento> document containing a
// single cancellation event (tpEvento=110111), ready to be signed. The caller
// is expected to sign the <infEvento> element and then serialize the document.
func BuildCancellationEvent(input CancelEventInput) (*etree.Document, error) {
	if len(input.AccessKey) != 44 {
		return nil, fmt.Errorf("access key must have 44 digits, got %d", len(input.AccessKey))
	}
	if input.Protocol == "" {
		return nil, fmt.Errorf("authorization protocol (nProt) is required for cancellation")
	}
	// Sanitize before validating: the check must run on the value that
	// actually goes on the wire (the justification comes from a free-text
	// textarea and may contain newlines or characters outside Latin-1).
	justification := SanitizeSchemaString(input.Justification)
	justLen := len([]rune(justification))
	if justLen < justificativaMinLen || justLen > justificativaMaxLen {
		return nil, fmt.Errorf("justificativa deve ter entre %d e %d caracteres", justificativaMinLen, justificativaMaxLen)
	}
	cUF := defaults.UFCode(input.EmitterUF)
	if cUF == "" {
		return nil, fmt.Errorf("UF do emitente inválida: %s", input.EmitterUF)
	}
	seqEvento := input.SeqEvento
	if seqEvento < 1 {
		seqEvento = 1
	}

	tpAmb := "2"
	if input.Environment == 1 {
		tpAmb = "1"
	}

	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	env := etree.NewElement("envEvento")
	env.CreateAttr("xmlns", nfeNamespace)
	env.CreateAttr("versao", VersaoEvento)
	env.CreateElement("idLote").SetText("1")

	evento := env.CreateElement("evento")
	evento.CreateAttr("versao", VersaoEvento)

	infEvento := evento.CreateElement("infEvento")
	// Id = "ID" + tpEvento + chNFe + nSeqEvento (2 digits, zero-padded)
	infEvento.CreateAttr("Id", fmt.Sprintf("ID%s%s%02d", TpEventoCancelamento, input.AccessKey, seqEvento))

	infEvento.CreateElement("cOrgao").SetText(cUF)
	infEvento.CreateElement("tpAmb").SetText(tpAmb)
	if input.EmitterType == 2 {
		infEvento.CreateElement("CPF").SetText(input.EmitterDoc)
	} else {
		infEvento.CreateElement("CNPJ").SetText(input.EmitterDoc)
	}
	infEvento.CreateElement("chNFe").SetText(input.AccessKey)
	infEvento.CreateElement("dhEvento").SetText(input.DhEvento.Format(time.RFC3339))
	infEvento.CreateElement("tpEvento").SetText(TpEventoCancelamento)
	infEvento.CreateElement("nSeqEvento").SetText(fmt.Sprintf("%d", seqEvento))
	infEvento.CreateElement("verEvento").SetText(VersaoEvento)

	detEvento := infEvento.CreateElement("detEvento")
	detEvento.CreateAttr("versao", VersaoEvento)
	detEvento.CreateElement("descEvento").SetText("Cancelamento")
	detEvento.CreateElement("nProt").SetText(input.Protocol)
	detEvento.CreateElement("xJust").SetText(justification)

	doc.SetRoot(env)
	return doc, nil
}
