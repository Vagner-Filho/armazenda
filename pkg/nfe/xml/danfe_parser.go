package xml

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"armazenda/pkg/nfe/entity"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

// ParseDANFEData extracts DANFE-relevant fields from an NF-e XML string.
// It handles both signed NF-e (<NFe>) and authorized NF-e (<nfeProc>) documents.
func ParseDANFEData(xmlContent string) (*entity.DANFEData, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}

	var nfeRoot *etree.Element
	if root.Tag == "nfeProc" {
		nfeRoot = root.SelectElement("NFe")
	} else if root.Tag == "NFe" {
		nfeRoot = root
	} else {
		// Fallback: search children for NFe element
		nfeRoot = root.SelectElement("NFe")
		if nfeRoot == nil {
			return nil, fmt.Errorf("NFe element not found in XML root <%s>", root.Tag)
		}
	}

	if nfeRoot == nil {
		return nil, fmt.Errorf("NFe element not found in XML")
	}

	infNFe := nfeRoot.SelectElement("infNFe")
	if infNFe == nil {
		return nil, fmt.Errorf("infNFe element not found")
	}

	data := &entity.DANFEData{}

	// ide
	if ide := infNFe.SelectElement("ide"); ide != nil {
		data.Serie = parseIntText(ide.SelectElement("serie"))
		data.Numero = parseIntText(ide.SelectElement("nNF"))
		data.NaturezaOp = textOrEmpty(ide.SelectElement("natOp"))
		data.EmissionDate = formatDateTime(textOrEmpty(ide.SelectElement("dhEmi")))
	}

	// emit
	if emit := infNFe.SelectElement("emit"); emit != nil {
		data.EmitterName = textOrEmpty(emit.SelectElement("xNome"))
		data.EmitterCNPJ = coalesce(
			textOrEmpty(emit.SelectElement("CNPJ")),
			textOrEmpty(emit.SelectElement("CPF")),
		)
		if ender := emit.SelectElement("enderEmit"); ender != nil {
			data.EmitterAddress = formatAddress(ender)
			data.EmitterCity = textOrEmpty(ender.SelectElement("xMun"))
			data.EmitterUF = textOrEmpty(ender.SelectElement("UF"))
		}
	}

	// dest
	if dest := infNFe.SelectElement("dest"); dest != nil {
		data.DestName = textOrEmpty(dest.SelectElement("xNome"))
		data.DestCNPJ = coalesce(
			textOrEmpty(dest.SelectElement("CNPJ")),
			textOrEmpty(dest.SelectElement("CPF")),
		)
		if ender := dest.SelectElement("enderDest"); ender != nil {
			data.DestAddress = formatAddress(ender)
			data.DestCity = textOrEmpty(ender.SelectElement("xMun"))
			data.DestUF = textOrEmpty(ender.SelectElement("UF"))
		}
	}

	// det / prod
	for _, det := range infNFe.SelectElements("det") {
		if prod := det.SelectElement("prod"); prod != nil {
			p := entity.DANFEProduct{
				Code: textOrEmpty(prod.SelectElement("cProd")),
				Desc: textOrEmpty(prod.SelectElement("xProd")),
				NCM:  textOrEmpty(prod.SelectElement("NCM")),
				CFOP: textOrEmpty(prod.SelectElement("CFOP")),
				Unit: textOrEmpty(prod.SelectElement("uCom")),
			}
			if q, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("qCom"))); err == nil {
				p.Quantity = q
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vUnCom"))); err == nil {
				p.UnitPrice = v
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vProd"))); err == nil {
				p.Total = v
			}
			data.Products = append(data.Products, p)
		}
	}

	// total
	if total := infNFe.SelectElement("total"); total != nil {
		if icmsTot := total.SelectElement("ICMSTot"); icmsTot != nil {
			data.TotalValue = parseDecimalText(icmsTot.SelectElement("vNF"))
			data.ICMSValue = parseDecimalText(icmsTot.SelectElement("vICMS"))
		}
	}

	// protNFe (authorized only)
	if root.Tag == "nfeProc" {
		if protNFe := root.SelectElement("protNFe"); protNFe != nil {
			if infProt := protNFe.SelectElement("infProt"); infProt != nil {
				data.Protocol = textOrEmpty(infProt.SelectElement("nProt"))
				data.ProtocolDate = formatDateTime(textOrEmpty(infProt.SelectElement("dhRecbto")))
			}
		}
	}

	// Access key from infNFe Id attribute
	if idAttr := infNFe.SelectAttrValue("Id", ""); strings.HasPrefix(idAttr, "NFe") {
		data.AccessKey = idAttr[3:]
	}

	return data, nil
}

func textOrEmpty(el *etree.Element) string {
	if el == nil {
		return ""
	}
	return el.Text()
}

func parseIntText(el *etree.Element) int {
	if el == nil {
		return 0
	}
	v, _ := strconv.Atoi(el.Text())
	return v
}

func parseDecimalText(el *etree.Element) decimal.Decimal {
	if el == nil {
		return decimal.Zero
	}
	if d, err := decimal.NewFromString(el.Text()); err == nil {
		return d
	}
	return decimal.Zero
}

func formatAddress(ender *etree.Element) string {
	parts := make([]string, 0, 4)
	if v := textOrEmpty(ender.SelectElement("xLgr")); v != "" {
		parts = append(parts, v)
	}
	if v := textOrEmpty(ender.SelectElement("nro")); v != "" {
		parts = append(parts, v)
	}
	if v := textOrEmpty(ender.SelectElement("xBairro")); v != "" {
		parts = append(parts, v)
	}
	return strings.Join(parts, ", ")
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func formatDateTime(iso string) string {
	if iso == "" {
		return ""
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, iso); err == nil {
			if layout == "2006-01-02" {
				return t.Format("02/01/2006")
			}
			return t.Format("02/01/2006 15:04:05")
		}
	}
	return iso
}
