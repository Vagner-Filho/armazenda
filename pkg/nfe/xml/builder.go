package xml

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

const (
	nfeNamespace = "http://www.portalfiscal.inf.br/nfe"
)

// Builder builds NF-e XML documents.
type Builder struct{}

// NewBuilder creates a new NF-e XML builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build creates the NF-e XML document from the input data.
func (b *Builder) Build(input entity.InvoiceInput) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	nfe := etree.NewElement("NFe")
	nfe.CreateAttr("xmlns", nfeNamespace)

	infNFe := nfe.CreateElement("infNFe")
	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:    defaults.UFCode(input.Emitter.UF),
		AAMM:   time.Now().Format("0601"),
		CNPJ:   input.Emitter.CNPJ,
		Mod:    defaults.ModeloNFe,
		Serie:  input.Serie,
		NNF:    input.Numero,
		TpEmis: "1",
		CNF:    fmt.Sprintf("%08d", input.Numero),
	})
	infNFe.CreateAttr("Id", "NFe"+accessKey)
	infNFe.CreateAttr("versao", defaults.VersaoLayout)

	// Build sections
	b.buildIDE(infNFe, input)
	b.buildEmit(infNFe, input.Emitter)
	b.buildDest(infNFe, input.Recipient)
	for i, item := range input.Items {
		b.buildDet(infNFe, item, i+1)
	}
	b.buildTotal(infNFe, input)
	b.buildTransp(infNFe, input.Transport)
	b.buildCobr(infNFe, input)
	b.buildPag(infNFe, input.Payment)
	if input.InformacoesAdicionais != "" {
		b.buildInfAdic(infNFe, input.InformacoesAdicionais)
	}

	doc.SetRoot(nfe)
	return doc, nil
}

func (b *Builder) buildIDE(parent *etree.Element, input entity.InvoiceInput) {
	ide := parent.CreateElement("ide")
	ide.CreateElement("cUF").SetText(defaults.UFCode(input.Emitter.UF))
	ide.CreateElement("cNF").SetText(fmt.Sprintf("%08d", input.Numero))
	ide.CreateElement("natOp").SetText(input.NaturezaOp)
	ide.CreateElement("mod").SetText(defaults.ModeloNFe)
	ide.CreateElement("serie").SetText(strconv.Itoa(input.Serie))
	ide.CreateElement("nNF").SetText(strconv.Itoa(input.Numero))
	ide.CreateElement("dhEmi").SetText(time.Now().Format(time.RFC3339))
	ide.CreateElement("tpNF").SetText("1")   // 1=Saida
	ide.CreateElement("idDest").SetText("1") // 1=Operacao interna (same state)
	ide.CreateElement("cMunFG").SetText(input.Emitter.CodigoMun)
	ide.CreateElement("tpImp").SetText("1")    // 1=Retrato
	ide.CreateElement("tpEmis").SetText("1")   // 1=Emissao normal
	ide.CreateElement("cDV").SetText("0")      // Will be calculated later
	ide.CreateElement("tpAmb").SetText("2")    // 2=Homologacao (override in production)
	ide.CreateElement("finNFe").SetText("1")   // 1=NF-e normal
	ide.CreateElement("indFinal").SetText("1") // 1=Consumidor final
	ide.CreateElement("indPres").SetText("0")  // 0=Nao se aplica
	ide.CreateElement("procEmi").SetText("0")  // 0=Emissao de NF-e com aplicativo do contribuinte
	ide.CreateElement("verProc").SetText("Armazenda-1.0")
}

func (b *Builder) buildEmit(parent *etree.Element, emit entity.EmitterData) {
	e := parent.CreateElement("emit")
	if emit.Type == 2 {
		e.CreateElement("CPF").SetText(emit.CPF)
	} else {
		e.CreateElement("CNPJ").SetText(emit.CNPJ)
	}
	e.CreateElement("xNome").SetText(emit.XNome)
	if emit.XFant != "" {
		e.CreateElement("xFant").SetText(emit.XFant)
	}
	enderEmit := e.CreateElement("enderEmit")
	enderEmit.CreateElement("xLgr").SetText(emit.Logradouro)
	enderEmit.CreateElement("nro").SetText(emit.Numero)
	if emit.Bairro != "" {
		enderEmit.CreateElement("xBairro").SetText(emit.Bairro)
	}
	enderEmit.CreateElement("cMun").SetText(emit.CodigoMun)
	enderEmit.CreateElement("xMun").SetText(emit.Municipio)
	enderEmit.CreateElement("UF").SetText(emit.UF)
	if emit.CEP != "" {
		enderEmit.CreateElement("CEP").SetText(strings.ReplaceAll(emit.CEP, "-", ""))
	}
	if emit.Fone != "" {
		enderEmit.CreateElement("fone").SetText(emit.Fone)
	}
	e.CreateElement("IE").SetText(emit.IE)
	e.CreateElement("CRT").SetText(emit.CRT)
}

func (b *Builder) buildDest(parent *etree.Element, dest entity.RecipientData) {
	d := parent.CreateElement("dest")
	if dest.Type == 2 {
		d.CreateElement("CPF").SetText(dest.CPF)
	} else {
		d.CreateElement("CNPJ").SetText(dest.CNPJ)
	}
	d.CreateElement("xNome").SetText(dest.XNome)
	enderDest := d.CreateElement("enderDest")
	enderDest.CreateElement("xLgr").SetText(dest.Logradouro)
	enderDest.CreateElement("nro").SetText(dest.Numero)
	if dest.Bairro != "" {
		enderDest.CreateElement("xBairro").SetText(dest.Bairro)
	}
	enderDest.CreateElement("cMun").SetText(dest.CodigoMun)
	enderDest.CreateElement("xMun").SetText(dest.Municipio)
	enderDest.CreateElement("UF").SetText(dest.UF)
	if dest.CEP != "" {
		enderDest.CreateElement("CEP").SetText(strings.ReplaceAll(dest.CEP, "-", ""))
	}
	if dest.Fone != "" {
		enderDest.CreateElement("fone").SetText(dest.Fone)
	}
	if dest.IE != "" {
		d.CreateElement("IE").SetText(dest.IE)
	}
	d.CreateElement("indIEDest").SetText(dest.IndIEDest)
}

func (b *Builder) buildDet(parent *etree.Element, item entity.ItemData, nItem int) {
	det := parent.CreateElement("det")
	det.CreateAttr("nItem", strconv.Itoa(nItem))

	// prod
	prod := det.CreateElement("prod")
	prod.CreateElement("cProd").SetText(item.Produto.Codigo)
	prod.CreateElement("cEAN").SetText(item.Produto.CEAN)
	prod.CreateElement("xProd").SetText(item.Produto.XProd)
	prod.CreateElement("NCM").SetText(item.Produto.NCM)
	prod.CreateElement("CFOP").SetText(item.Produto.CFOP)
	prod.CreateElement("uCom").SetText(item.Produto.UCom)
	prod.CreateElement("qCom").SetText(formatDecimal(item.Produto.QCom, 4))
	prod.CreateElement("vUnCom").SetText(formatDecimal(item.Produto.VUnCom, 10))
	prod.CreateElement("vProd").SetText(formatDecimal(item.Produto.VProd, 2))
	prod.CreateElement("cEANTrib").SetText(item.Produto.CEANTrib)
	prod.CreateElement("uTrib").SetText(item.Produto.UTrib)
	prod.CreateElement("qTrib").SetText(formatDecimal(item.Produto.QTrib, 4))
	prod.CreateElement("vUnTrib").SetText(formatDecimal(item.Produto.VUnTrib, 10))
	prod.CreateElement("indTot").SetText(strconv.Itoa(item.Produto.IndTot))

	// imposto
	imp := det.CreateElement("imposto")
	b.buildICMS(imp, item.Imposto.ICMS)
	b.buildPIS(imp, item.Imposto.PIS)
	b.buildCOFINS(imp, item.Imposto.COFINS)

	if item.InfAdProd != "" {
		det.CreateElement("infAdProd").SetText(item.InfAdProd)
	}
}

func (b *Builder) buildICMS(parent *etree.Element, icms entity.ICMSData) {
	icmsElem := parent.CreateElement("ICMS")
	icms00 := icmsElem.CreateElement("ICMS00")
	icms00.CreateElement("orig").SetText(icms.Origem)
	icms00.CreateElement("CST").SetText(icms.CST)
	icms00.CreateElement("modBC").SetText(icms.ModBC)
	icms00.CreateElement("vBC").SetText(formatDecimal(icms.VBC, 2))
	icms00.CreateElement("pICMS").SetText(formatDecimal(icms.PICMS, 4))
	icms00.CreateElement("vICMS").SetText(formatDecimal(icms.VICMS, 2))
}

func (b *Builder) buildPIS(parent *etree.Element, pis entity.PISData) {
	pisElem := parent.CreateElement("PIS")
	pisAliq := pisElem.CreateElement("PISAliq")
	pisAliq.CreateElement("CST").SetText(pis.CST)
	pisAliq.CreateElement("vBC").SetText(formatDecimal(pis.VBC, 2))
	pisAliq.CreateElement("pPIS").SetText(formatDecimal(pis.PPIS, 4))
	pisAliq.CreateElement("vPIS").SetText(formatDecimal(pis.VPIS, 2))
}

func (b *Builder) buildCOFINS(parent *etree.Element, cofins entity.COFINSData) {
	cofinsElem := parent.CreateElement("COFINS")
	cofinsAliq := cofinsElem.CreateElement("COFINSAliq")
	cofinsAliq.CreateElement("CST").SetText(cofins.CST)
	cofinsAliq.CreateElement("vBC").SetText(formatDecimal(cofins.VBC, 2))
	cofinsAliq.CreateElement("pCOFINS").SetText(formatDecimal(cofins.PCOFINS, 4))
	cofinsAliq.CreateElement("vCOFINS").SetText(formatDecimal(cofins.VCOFINS, 2))
}

func (b *Builder) buildTotal(parent *etree.Element, input entity.InvoiceInput) {
	total := parent.CreateElement("total")
	icmsTot := total.CreateElement("ICMSTot")

	var vBC, vICMS, vProd decimal.Decimal
	for _, item := range input.Items {
		vBC = vBC.Add(item.Imposto.ICMS.VBC)
		vICMS = vICMS.Add(item.Imposto.ICMS.VICMS)
		vProd = vProd.Add(item.Produto.VProd)
	}
	vNF := input.TotalValue

	icmsTot.CreateElement("vBC").SetText(formatDecimal(vBC, 2))
	icmsTot.CreateElement("vICMS").SetText(formatDecimal(vICMS, 2))
	icmsTot.CreateElement("vICMSDeson").SetText("0.00")
	icmsTot.CreateElement("vFCP").SetText("0.00")
	icmsTot.CreateElement("vBCST").SetText("0.00")
	icmsTot.CreateElement("vST").SetText("0.00")
	icmsTot.CreateElement("vFCPST").SetText("0.00")
	icmsTot.CreateElement("vFCPSTRet").SetText("0.00")
	icmsTot.CreateElement("vProd").SetText(formatDecimal(vProd, 2))
	icmsTot.CreateElement("vFrete").SetText("0.00")
	icmsTot.CreateElement("vSeg").SetText("0.00")
	icmsTot.CreateElement("vDesc").SetText("0.00")
	icmsTot.CreateElement("vII").SetText("0.00")
	icmsTot.CreateElement("vIPI").SetText("0.00")
	icmsTot.CreateElement("vIPIDevol").SetText("0.00")
	icmsTot.CreateElement("vPIS").SetText("0.00")
	icmsTot.CreateElement("vCOFINS").SetText("0.00")
	icmsTot.CreateElement("vOutro").SetText("0.00")
	icmsTot.CreateElement("vNF").SetText(formatDecimal(vNF, 2))
}

func (b *Builder) buildTransp(parent *etree.Element, transp entity.TransportData) {
	transpElem := parent.CreateElement("transp")
	transpElem.CreateElement("modFrete").SetText(strconv.Itoa(transp.ModFrete))

	if transp.Transportadora != nil {
		transporta := transpElem.CreateElement("transporta")
		if transp.Transportadora.Type == 2 {
			transporta.CreateElement("CPF").SetText(transp.Transportadora.CPF)
		} else {
			transporta.CreateElement("CNPJ").SetText(transp.Transportadora.CNPJ)
		}
		transporta.CreateElement("xNome").SetText(transp.Transportadora.XNome)
		if transp.Transportadora.IE != "" {
			transporta.CreateElement("IE").SetText(transp.Transportadora.IE)
		}
		transporta.CreateElement("xEnder").SetText(transp.Transportadora.Municipio)
		transporta.CreateElement("xMun").SetText(transp.Transportadora.Municipio)
		transporta.CreateElement("UF").SetText(transp.Transportadora.UF)
	}

	if transp.Veiculo != nil {
		veicTransp := transpElem.CreateElement("veicTransp")
		veicTransp.CreateElement("placa").SetText(transp.Veiculo.Placa)
		veicTransp.CreateElement("UF").SetText(transp.Veiculo.UF)
		if transp.Veiculo.RNTC != "" {
			veicTransp.CreateElement("RNTC").SetText(transp.Veiculo.RNTC)
		}
	}

	if len(transp.Volumes) > 0 {
		vol := transpElem.CreateElement("vol")
		for _, v := range transp.Volumes {
			vol.CreateElement("qVol").SetText(strconv.Itoa(v.QVol))
			if v.Esp != "" {
				vol.CreateElement("esp").SetText(v.Esp)
			}
			if v.Marca != "" {
				vol.CreateElement("marca").SetText(v.Marca)
			}
			if v.NVol != "" {
				vol.CreateElement("nVol").SetText(v.NVol)
			}
			vol.CreateElement("pesoL").SetText(formatDecimal(v.PesoL, 3))
			vol.CreateElement("pesoB").SetText(formatDecimal(v.PesoB, 3))
		}
	}
}

func (b *Builder) buildCobr(parent *etree.Element, input entity.InvoiceInput) {
	if input.TotalValue.IsZero() {
		return
	}
	cobr := parent.CreateElement("cobr")
	fat := cobr.CreateElement("fat")
	fat.CreateElement("nFat").SetText(fmt.Sprintf("%d", input.Numero))
	fat.CreateElement("vOrig").SetText(formatDecimal(input.TotalValue, 2))
	fat.CreateElement("vLiq").SetText(formatDecimal(input.TotalValue, 2))
}

func (b *Builder) buildPag(parent *etree.Element, pag entity.PaymentData) {
	pagElem := parent.CreateElement("pag")
	detPag := pagElem.CreateElement("detPag")
	detPag.CreateElement("indPag").SetText(strconv.Itoa(pag.IndPag))
	for _, det := range pag.Detalhes {
		detPag.CreateElement("tPag").SetText(det.TPag)
		detPag.CreateElement("vPag").SetText(formatDecimal(det.VPag, 2))
	}
}

func (b *Builder) buildInfAdic(parent *etree.Element, info string) {
	infAdic := parent.CreateElement("infAdic")
	infAdic.CreateElement("infCpl").SetText(info)
}

func formatDecimal(d decimal.Decimal, places int32) string {
	return d.Truncate(places).StringFixed(places)
}
