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
		data.TpEmis = textOrEmpty(ide.SelectElement("tpEmis"))
		data.TpAmb = textOrEmpty(ide.SelectElement("tpAmb"))
		data.TpNF = textOrEmpty(ide.SelectElement("tpNF"))
		data.DhSaiEnt = formatDateTime(textOrEmpty(ide.SelectElement("dhSaiEnt")))
		data.DhCont = formatDateTime(textOrEmpty(ide.SelectElement("dhCont")))
		data.XJust = textOrEmpty(ide.SelectElement("xJust"))
		data.VerProc = textOrEmpty(ide.SelectElement("verProc"))
	}

	// emit
	if emit := infNFe.SelectElement("emit"); emit != nil {
		data.EmitterName = textOrEmpty(emit.SelectElement("xNome"))
		data.EmitterCNPJ = coalesce(
			textOrEmpty(emit.SelectElement("CNPJ")),
			textOrEmpty(emit.SelectElement("CPF")),
		)
		data.EmitterIE = textOrEmpty(emit.SelectElement("IE"))
		data.EmitterCRT = textOrEmpty(emit.SelectElement("CRT"))
		if ender := emit.SelectElement("enderEmit"); ender != nil {
			data.EmitterAddress = textOrEmpty(ender.SelectElement("xLgr"))
			data.EmitterNumber = textOrEmpty(ender.SelectElement("nro"))
			data.EmitterComplement = textOrEmpty(ender.SelectElement("xCpl"))
			data.EmitterNeighborhood = textOrEmpty(ender.SelectElement("xBairro"))
			data.EmitterCEP = textOrEmpty(ender.SelectElement("CEP"))
			data.EmitterCity = textOrEmpty(ender.SelectElement("xMun"))
			data.EmitterUF = textOrEmpty(ender.SelectElement("UF"))
			data.EmitterPhone = textOrEmpty(ender.SelectElement("fone"))
		}
	}

	// dest
	if dest := infNFe.SelectElement("dest"); dest != nil {
		data.DestName = textOrEmpty(dest.SelectElement("xNome"))
		data.DestCNPJ = coalesce(
			textOrEmpty(dest.SelectElement("CNPJ")),
			textOrEmpty(dest.SelectElement("CPF")),
		)
		data.DestIE = textOrEmpty(dest.SelectElement("IE"))
		data.DestIndIEDest = textOrEmpty(dest.SelectElement("indIEDest"))
		if ender := dest.SelectElement("enderDest"); ender != nil {
			data.DestAddress = textOrEmpty(ender.SelectElement("xLgr"))
			data.DestNumber = textOrEmpty(ender.SelectElement("nro"))
			data.DestComplement = textOrEmpty(ender.SelectElement("xCpl"))
			data.DestNeighborhood = textOrEmpty(ender.SelectElement("xBairro"))
			data.DestCEP = textOrEmpty(ender.SelectElement("CEP"))
			data.DestCity = textOrEmpty(ender.SelectElement("xMun"))
			data.DestUF = textOrEmpty(ender.SelectElement("UF"))
			data.DestPhone = textOrEmpty(ender.SelectElement("fone"))
		}
	}

	// det / prod + imposto
	for _, det := range infNFe.SelectElements("det") {
		var p entity.DANFEProduct
		if prod := det.SelectElement("prod"); prod != nil {
			p = entity.DANFEProduct{
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
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vFrete"))); err == nil {
				p.VFrete = v
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vSeg"))); err == nil {
				p.VSeg = v
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vDesc"))); err == nil {
				p.VDesc = v
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vOutro"))); err == nil {
				p.VOutro = v
			}
			// Tributable unit (Grupo I)
			p.UTrib = textOrEmpty(prod.SelectElement("uTrib"))
			if q, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("qTrib"))); err == nil {
				p.QTrib = q
			}
			if v, err := decimal.NewFromString(textOrEmpty(prod.SelectElement("vUnTrib"))); err == nil {
				p.VUnTrib = v
			}
		}

		// infAdProd is a sibling of prod inside det (MOC Anexo I, id V01)
		p.InfAdProd = textOrEmpty(det.SelectElement("infAdProd"))

		// imposto
		if imposto := det.SelectElement("imposto"); imposto != nil {
			// ICMS (try common variants)
			if icms := imposto.SelectElement("ICMS"); icms != nil {
				for _, child := range icms.ChildElements() {
					if strings.HasPrefix(child.Tag, "ICMS") {
						p.CST = textOrEmpty(child.SelectElement("CST"))
						p.VBC = parseDecimalText(child.SelectElement("vBC"))
						p.PICMS = parseDecimalText(child.SelectElement("pICMS"))
						p.VICMS = parseDecimalText(child.SelectElement("vICMS"))
						break
					}
				}
			}
			// IPI
			if ipi := imposto.SelectElement("IPI"); ipi != nil {
				if ipiTrib := ipi.SelectElement("IPITrib"); ipiTrib != nil {
					p.PIPI = parseDecimalText(ipiTrib.SelectElement("pIPI"))
					p.VIPI = parseDecimalText(ipiTrib.SelectElement("vIPI"))
				}
			}
			// PIS
			if pis := imposto.SelectElement("PIS"); pis != nil {
				for _, child := range pis.ChildElements() {
					if strings.HasPrefix(child.Tag, "PIS") {
						p.PPIS = parseDecimalText(child.SelectElement("pPIS"))
						p.VPIS = parseDecimalText(child.SelectElement("vPIS"))
						break
					}
				}
			}
			// COFINS
			if cofins := imposto.SelectElement("COFINS"); cofins != nil {
				for _, child := range cofins.ChildElements() {
					if strings.HasPrefix(child.Tag, "COFINS") {
						p.PCOFINS = parseDecimalText(child.SelectElement("pCOFINS"))
						p.VCOFINS = parseDecimalText(child.SelectElement("vCOFINS"))
						break
					}
				}
			}
		}
		data.Products = append(data.Products, p)
	}

	// total
	if total := infNFe.SelectElement("total"); total != nil {
		if icmsTot := total.SelectElement("ICMSTot"); icmsTot != nil {
			data.TotalValue = parseDecimalText(icmsTot.SelectElement("vNF"))
			data.VBC = parseDecimalText(icmsTot.SelectElement("vBC"))
			data.VICMS = parseDecimalText(icmsTot.SelectElement("vICMS"))
			data.VICMSDeson = parseDecimalText(icmsTot.SelectElement("vICMSDeson"))
			data.VBCST = parseDecimalText(icmsTot.SelectElement("vBCST"))
			data.VST = parseDecimalText(icmsTot.SelectElement("vST"))
			data.VII = parseDecimalText(icmsTot.SelectElement("vII"))
			data.VIPI = parseDecimalText(icmsTot.SelectElement("vIPI"))
			data.VPIS = parseDecimalText(icmsTot.SelectElement("vPIS"))
			data.VCOFINS = parseDecimalText(icmsTot.SelectElement("vCOFINS"))
			data.VFrete = parseDecimalText(icmsTot.SelectElement("vFrete"))
			data.VSeg = parseDecimalText(icmsTot.SelectElement("vSeg"))
			data.VDesc = parseDecimalText(icmsTot.SelectElement("vDesc"))
			data.VOutro = parseDecimalText(icmsTot.SelectElement("vOutro"))
			data.VTotTrib = parseDecimalText(icmsTot.SelectElement("vTotTrib"))
		}
		if issqnTot := total.SelectElement("ISSQNtot"); issqnTot != nil {
			data.VBCISSQN = parseDecimalText(issqnTot.SelectElement("vBC"))
			data.VISSQN = parseDecimalText(issqnTot.SelectElement("vISS"))
			data.VPISISSQN = parseDecimalText(issqnTot.SelectElement("vPIS"))
			data.VCOFINSSISSQN = parseDecimalText(issqnTot.SelectElement("vCOFINS"))
		}
	}

	// transp
	if transp := infNFe.SelectElement("transp"); transp != nil {
		data.ModFrete = textOrEmpty(transp.SelectElement("modFrete"))
		if transporta := transp.SelectElement("transporta"); transporta != nil {
			data.TranspName = textOrEmpty(transporta.SelectElement("xNome"))
			data.TranspCNPJ = coalesce(
				textOrEmpty(transporta.SelectElement("CNPJ")),
				textOrEmpty(transporta.SelectElement("CPF")),
			)
			data.TranspIE = textOrEmpty(transporta.SelectElement("IE"))
			data.TranspAddress = textOrEmpty(transporta.SelectElement("xEnder"))
			data.TranspCity = textOrEmpty(transporta.SelectElement("xMun"))
			data.TranspUF = textOrEmpty(transporta.SelectElement("UF"))
		}
		if vol := transp.SelectElement("vol"); vol != nil {
			data.QVol = textOrEmpty(vol.SelectElement("qVol"))
			data.Esp = textOrEmpty(vol.SelectElement("esp"))
			data.Marca = textOrEmpty(vol.SelectElement("marca"))
			data.NVol = textOrEmpty(vol.SelectElement("nVol"))
			data.PesoL = parseDecimalText(vol.SelectElement("pesoL"))
			data.PesoB = parseDecimalText(vol.SelectElement("pesoB"))
		}
		if veic := transp.SelectElement("veicTransp"); veic != nil {
			data.VeicPlate = textOrEmpty(veic.SelectElement("placa"))
			data.VeicUF = textOrEmpty(veic.SelectElement("UF"))
		}
	}

	// infAdic
	if infAdic := infNFe.SelectElement("infAdic"); infAdic != nil {
		data.InfCpl = textOrEmpty(infAdic.SelectElement("infCpl"))
		data.InfAdFisco = textOrEmpty(infAdic.SelectElement("infAdFisco"))
	}

	// protNFe (authorized only)
	if root.Tag == "nfeProc" {
		if protNFe := root.SelectElement("protNFe"); protNFe != nil {
			if infProt := protNFe.SelectElement("infProt"); infProt != nil {
				data.Protocol = textOrEmpty(infProt.SelectElement("nProt"))
				data.ProtocolDate = formatDateTime(textOrEmpty(infProt.SelectElement("dhRecbto")))
				data.CStat = textOrEmpty(infProt.SelectElement("cStat"))
				data.XMotivo = textOrEmpty(infProt.SelectElement("xMotivo"))
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
