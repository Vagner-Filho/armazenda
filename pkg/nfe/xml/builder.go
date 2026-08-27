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
func (b *Builder) documentForAccessKey(emit entity.EmitterData) string {
	if emit.Type == 2 {
		return padLeftZeros(emit.Document, 14)
	}
	return padLeftZeros(emit.Document, 14)
}

func padLeftZeros(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}
	if len(s) > length {
		return s[:length]
	}
	return s
}

func (b *Builder) Build(input entity.InvoiceInput) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true

	nfe := etree.NewElement("NFe")
	nfe.CreateAttr("xmlns", nfeNamespace)

	infNFe := nfe.CreateElement("infNFe")
	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:      defaults.UFCode(input.Emitter.UF),
		AAMM:     time.Now().Format("0601"),
		Document: b.documentForAccessKey(input.Emitter),
		Mod:      defaults.ModeloNFe,
		Serie:    input.Serie,
		NNF:      input.Numero,
		TpEmis:   input.TpEmis.String(),
		CNF:      input.CNF,
	})
	infNFe.CreateAttr("Id", "NFe"+accessKey)
	infNFe.CreateAttr("versao", defaults.VersaoLayout)

	// Build sections
	b.buildIDE(infNFe, input, accessKey)
	b.buildEmit(infNFe, input.Emitter)
	b.buildDest(infNFe, input.Recipient, input.Environment)
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

func (b *Builder) buildIDE(parent *etree.Element, input entity.InvoiceInput, accessKey string) {
	ide := parent.CreateElement("ide")
	ide.CreateElement("cUF").SetText(defaults.UFCode(input.Emitter.UF))
	ide.CreateElement("cNF").SetText(input.CNF)
	setSchemaText(ide, "natOp", input.NaturezaOp)
	ide.CreateElement("mod").SetText(defaults.ModeloNFe)
	ide.CreateElement("serie").SetText(strconv.Itoa(input.Serie))
	ide.CreateElement("nNF").SetText(strconv.Itoa(input.Numero))
	ide.CreateElement("dhEmi").SetText(time.Now().Format(time.RFC3339))
	ide.CreateElement("tpNF").SetText("1") // 1=Saida

	// idDest: 1=Operacao interna, 2=Inter estadual
	idDest := "1"
	if input.Emitter.UF != input.Recipient.UF && input.Recipient.UF != "" {
		idDest = "2"
	}
	ide.CreateElement("idDest").SetText(idDest)

	ide.CreateElement("cMunFG").SetText(input.Emitter.CodigoMun)
	ide.CreateElement("tpImp").SetText("1") // 1=Retrato
	ide.CreateElement("tpEmis").SetText(input.TpEmis.String())

	// Contingency fields: required when tpEmis != 1
	if input.TpEmis != defaults.EmissaoNormal {
		if input.DhCont != nil {
			ide.CreateElement("dhCont").SetText(input.DhCont.Format(time.RFC3339))
		}
		if input.XJust != "" {
			setSchemaText(ide, "xJust", input.XJust)
		}
	}

	// DV is the last digit of the 44-digit access key
	cDV := "0"
	if len(accessKey) == 44 {
		cDV = accessKey[43:]
	}
	ide.CreateElement("cDV").SetText(cDV)

	// tpAmb from input environment
	tpAmb := "2"
	if input.Environment == 1 {
		tpAmb = "1"
	}
	ide.CreateElement("tpAmb").SetText(tpAmb)

	ide.CreateElement("finNFe").SetText("1") // 1=NF-e normal

	// indFinal: 0=Normal (B2B), 1=Consumidor final
	indFinal := "0"
	if input.Recipient.IndIEDest == "9" {
		indFinal = "1" // Non-contributor is typically a final consumer
	}
	ide.CreateElement("indFinal").SetText(indFinal)

	// indPres: 9=Nao presencial (remote/online) for grain sales
	ide.CreateElement("indPres").SetText("9")

	// indIntermed: 0=Sem operacao com intermediador, 1=Com intermediador
	ide.CreateElement("indIntermed").SetText("0")

	ide.CreateElement("procEmi").SetText("0") // 0=Emissao de NF-e com aplicativo do contribuinte
	ide.CreateElement("verProc").SetText("Armazenda-1.0")
}

func (b *Builder) buildEmit(parent *etree.Element, emit entity.EmitterData) {
	e := parent.CreateElement("emit")
	if emit.Type == 2 {
		e.CreateElement("CPF").SetText(emit.Document)
	} else {
		e.CreateElement("CNPJ").SetText(emit.Document)
	}
	setSchemaText(e, "xNome", emit.XNome)
	if emit.XFant != "" {
		setSchemaText(e, "xFant", emit.XFant)
	}
	enderEmit := e.CreateElement("enderEmit")
	setSchemaText(enderEmit, "xLgr", emit.Logradouro)
	setSchemaText(enderEmit, "nro", emit.Numero)
	// xBairro is 1-1 mandatory (C09); the caller must ensure it's populated.
	setSchemaText(enderEmit, "xBairro", emit.Bairro)
	enderEmit.CreateElement("cMun").SetText(emit.CodigoMun)
	setSchemaText(enderEmit, "xMun", emit.Municipio)
	enderEmit.CreateElement("UF").SetText(emit.UF)
	// CEP is 1-1 mandatory (C13); "Informar os zeros não significativos."
	cep := strings.ReplaceAll(emit.CEP, "-", "")
	enderEmit.CreateElement("CEP").SetText(cep)
	if emit.Fone != "" {
		enderEmit.CreateElement("fone").SetText(emit.Fone)
	}
	e.CreateElement("IE").SetText(strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, emit.IE))
	e.CreateElement("CRT").SetText(emit.CRT)
}

func (b *Builder) buildDest(parent *etree.Element, dest entity.RecipientData, environment int) {
	d := parent.CreateElement("dest")
	if dest.Type == 2 {
		d.CreateElement("CPF").SetText(dest.CPF)
	} else {
		d.CreateElement("CNPJ").SetText(dest.CNPJ)
	}
	// Homologation environment requires fixed recipient name per SEFAZ rules
	if environment == 2 {
		dest.XNome = "NF-E EMITIDA EM AMBIENTE DE HOMOLOGACAO - SEM VALOR FISCAL"
	}
	setSchemaText(d, "xNome", dest.XNome)
	enderDest := d.CreateElement("enderDest")
	setSchemaText(enderDest, "xLgr", dest.Logradouro)
	setSchemaText(enderDest, "nro", dest.Numero)
	// xBairro is 1-1 mandatory (E09); the caller must ensure it's populated.
	setSchemaText(enderDest, "xBairro", dest.Bairro)
	enderDest.CreateElement("cMun").SetText(dest.CodigoMun)
	setSchemaText(enderDest, "xMun", dest.Municipio)
	enderDest.CreateElement("UF").SetText(dest.UF)
	if dest.CEP != "" {
		enderDest.CreateElement("CEP").SetText(strings.ReplaceAll(dest.CEP, "-", ""))
	}
	if dest.Fone != "" {
		enderDest.CreateElement("fone").SetText(dest.Fone)
	}
	d.CreateElement("indIEDest").SetText(dest.IndIEDest)
	if dest.IE != "" {
		d.CreateElement("IE").SetText(dest.IE)
	}
}

func (b *Builder) buildDet(parent *etree.Element, item entity.ItemData, nItem int) {
	det := parent.CreateElement("det")
	det.CreateAttr("nItem", strconv.Itoa(nItem))

	// prod
	prod := det.CreateElement("prod")
	prod.CreateElement("cProd").SetText(item.Produto.Codigo)
	prod.CreateElement("cEAN").SetText(item.Produto.CEAN)
	setSchemaText(prod, "xProd", item.Produto.XProd)
	prod.CreateElement("NCM").SetText(item.Produto.NCM)
	if item.Produto.CEST != "" {
		prod.CreateElement("CEST").SetText(item.Produto.CEST)
	}
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
	// IBSCBS is emitted for every item: when IsTaxReformActive is false the
	// caller passes zero-valued IBSCBSData so the group renders but stays
	// zero-valued (stable schema across years).
	b.buildIBSCBS(imp, item.Imposto.IBSCBS)

	if item.InfAdProd != "" {
		setSchemaText(det, "infAdProd", item.InfAdProd)
	}
}

func (b *Builder) buildICMS(parent *etree.Element, icms entity.ICMSData) {
	icmsElem := parent.CreateElement("ICMS")

	// Simples Nacional: use CSOSN-based tags
	if icms.CSOSN != "" {
		switch icms.CSOSN {
		case "101":
			b.buildICMSSN101(icmsElem, icms)
		case "201":
			b.buildICMSSN201(icmsElem, icms)
		case "202", "203":
			b.buildICMSSN202(icmsElem, icms)
		case "500":
			b.buildICMSSN500(icmsElem, icms)
		case "900":
			b.buildICMSSN900(icmsElem, icms)
		default:
			// 102, 103, 300, 400 → ICMSSN102
			b.buildICMSSN102(icmsElem, icms)
		}
		return
	}

	// Regime normal: use CST-based tags
	icms00 := icmsElem.CreateElement("ICMS00")
	icms00.CreateElement("orig").SetText(icms.Origem)
	icms00.CreateElement("CST").SetText(icms.CST)
	icms00.CreateElement("modBC").SetText(icms.ModBC)
	icms00.CreateElement("vBC").SetText(formatDecimal(icms.VBC, 2))
	icms00.CreateElement("pICMS").SetText(formatDecimal(icms.PICMS, 4))
	icms00.CreateElement("vICMS").SetText(formatDecimal(icms.VICMS, 2))
}

func (b *Builder) buildICMSSN101(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN101")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
}

func (b *Builder) buildICMSSN102(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN102")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
}

func (b *Builder) buildICMSSN201(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN201")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
}

func (b *Builder) buildICMSSN202(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN202")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
}

func (b *Builder) buildICMSSN500(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN500")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
}

func (b *Builder) buildICMSSN900(parent *etree.Element, icms entity.ICMSData) {
	elem := parent.CreateElement("ICMSSN900")
	elem.CreateElement("orig").SetText(icms.Origem)
	elem.CreateElement("CSOSN").SetText(icms.CSOSN)
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

// buildIBSCBS emits the per-item <IBSCBS> group required by the indirect tax
// reform (NT 2025.002-RTC v1.51 §6.7.4, Grupo UB). For the 2026 symbolic
// phase the IBS rate is allocated entirely to the state share (pIBSUF) and
// the municipal share stays zero — state-level split rates are not yet
// published. vIBSUF + vIBSMun = vIBS (per-item total, UB54a).
//
// Emitted always (even with zero values) so the schema stays stable when the
// caller passes zero-valued IBSCBSData for pre-reform emissions.
//
// Per-item <gIBSCBS> (UB15) is a FLAT sequence — gIBSUF, gIBSMun, vIBS,
// gCBS are direct siblings of each other inside gIBSCBS. There is NO
// <gIBS> wrapper at this level (the totals block uses one, but the per-item
// block does not). SEFAZ rejects with status 215 cvc-complex-type.2.4.a
// ("Invalid content starting with element 'gIBS'. One of 'gIBSUF' is
// expected") when a <gIBS> wrapper is mistakenly added.
func (b *Builder) buildIBSCBS(parent *etree.Element, ibscbs entity.IBSCBSData) {
	ibscbsElem := parent.CreateElement("IBSCBS")
	ibscbsElem.CreateElement("CST").SetText(ibscbs.CST)
	ibscbsElem.CreateElement("cClassTrib").SetText(ibscbs.CClassTrib)

	gIBSCBS := ibscbsElem.CreateElement("gIBSCBS")
	// gIBSCBS children per UB15/NT v1.51: vBC, gIBSUF, gIBSMun, vIBS, gCBS.
	// vBC is mandatory (UB16, 1-1). gIBSUF (UB17), gIBSMun (UB36), gCBS
	// (UB55) are mandatory inner groups. vIBS (UB54a) is the per-item IBS
	// total = vIBSUF + vIBSMun, mandatory 1-1 directly under gIBSCBS.
	gIBSCBS.CreateElement("vBC").SetText(formatDecimal(ibscbs.VBC, 2))

	gIBSUF := gIBSCBS.CreateElement("gIBSUF")
	gIBSUF.CreateElement("pIBSUF").SetText(formatDecimal(perItemIBSRate(ibscbs), 4))
	gIBSUF.CreateElement("vIBSUF").SetText(formatDecimal(ibscbs.VIBSUF, 2))

	gIBSMun := gIBSCBS.CreateElement("gIBSMun")
	gIBSMun.CreateElement("pIBSMun").SetText("0.0000")
	gIBSMun.CreateElement("vIBSMun").SetText(formatDecimal(ibscbs.VIBSMun, 2))

	// <vIBS> (UB54a): per-item IBS total = vIBSUF + vIBSMun.
	gIBSCBS.CreateElement("vIBS").SetText(formatDecimal(ibscbs.VIBS, 2))

	gCBS := gIBSCBS.CreateElement("gCBS")
	gCBS.CreateElement("pCBS").SetText(formatDecimal(ibscbs.PCBS, 4))
	gCBS.CreateElement("vCBS").SetText(formatDecimal(ibscbs.VCBS, 2))
	// Per-item <vCredPres> / <vCredPresCondSus> are NOT direct children of
	// <gCBS> — they are totals-only fields (W48/W49 in totals; per-item
	// credit-presumption fields belong inside the optional <gIBSCredPres>
	// and <gCBSCredPres> subgroups, which we don't emit in the 2026 phase).
}

// perItemIBSRate recovers the per-item IBS rate (as a percentage value, e.g.
// 0.10 for 0.1%) from VIBSUF/VBC. Used by buildIBSCBS to keep <pIBSUF>
// consistent with <vIBSUF> in the per-item <gIBSUF> group. The 2026 phase
// allocates the full IBS rate to the UF share (vIBSMun = 0).
func perItemIBSRate(ibscbs entity.IBSCBSData) decimal.Decimal {
	if ibscbs.VBC.IsZero() {
		return decimal.Zero
	}
	rate := ibscbs.VIBSUF.Div(ibscbs.VBC)
	return rate.Mul(decimal.NewFromInt(100))
}

func (b *Builder) buildTotal(parent *etree.Element, input entity.InvoiceInput) {
	total := parent.CreateElement("total")
	icmsTot := total.CreateElement("ICMSTot")

	var vBC, vICMS, vProd, vPIS, vCOFINS decimal.Decimal
	var vBCIBSCBS, vIBS, vCBS decimal.Decimal
	for _, item := range input.Items {
		vBC = vBC.Add(item.Imposto.ICMS.VBC)
		vICMS = vICMS.Add(item.Imposto.ICMS.VICMS)
		vProd = vProd.Add(item.Produto.VProd)
		vPIS = vPIS.Add(item.Imposto.PIS.VPIS)
		vCOFINS = vCOFINS.Add(item.Imposto.COFINS.VCOFINS)
		vBCIBSCBS = vBCIBSCBS.Add(item.Imposto.IBSCBS.VBC)
		vIBS = vIBS.Add(item.Imposto.IBSCBS.VIBS)
		vCBS = vCBS.Add(item.Imposto.IBSCBS.VCBS)
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
	icmsTot.CreateElement("vPIS").SetText(formatDecimal(vPIS, 2))
	icmsTot.CreateElement("vCOFINS").SetText(formatDecimal(vCOFINS, 2))
	icmsTot.CreateElement("vOutro").SetText("0.00")
	icmsTot.CreateElement("vNF").SetText(formatDecimal(vNF, 2))

	// IBSCBSTot — sibling of ICMSTot inside <total>. Mandatory in the NF-e
	// layout from August 2026 (Reforma Tributária). The IBS total equals the
	// UF share for now (municipal share is zero in the 2026 symbolic phase).
	//
	// Element order and occurrence are dictated by NT 2025.002-RTC v1.36 §6.7.4
	// and verified against the NFe_Util 2Gv5.02b reference. SEFAZ rejects the
	// document (status 215, cvc-complex-type.2.4.a) when children are missing
	// or out of order — this is the canonical sequence:
	//
	//   IBSCBSTot: vBCIBSCBS, gIBS, gCBS            (gMono optional 0-1, skipped)
	//   gIBS:      gIBSUF, gIBSMun, vIBS, vCredPres, vCredPresCondSus
	//   gIBSUF:    vDif, vDevTrib, vIBSUF           (all 1-1)
	//   gIBSMun:   vDif, vDevTrib, vIBSMun          (all 1-1)
	//   gCBS:      vDif, vDevTrib, vCBS, vCredPres, vCredPresCondSus
	//
	// Deferral / credit fields (pDif, vDif, vDevTrib, pRedAliq, pAliqEfet,
	// gIBSCredPres, gCBSCredPres per item) are 0-1 and skipped — grain sales
	// do not exercise them in the 2026 symbolic phase.
	ibscbsTot := total.CreateElement("IBSCBSTot")
	ibscbsTot.CreateElement("vBCIBSCBS").SetText(formatDecimal(vBCIBSCBS, 2))

	gIBS := ibscbsTot.CreateElement("gIBS")
	gIBSUF := gIBS.CreateElement("gIBSUF")
	gIBSUF.CreateElement("vDif").SetText("0.00")                          // W38
	gIBSUF.CreateElement("vDevTrib").SetText("0.00")                      // W39
	gIBSUF.CreateElement("vIBSUF").SetText(formatDecimal(vIBS, 2))        // W41

	gIBSMun := gIBS.CreateElement("gIBSMun")
	gIBSMun.CreateElement("vDif").SetText("0.00")                         // W43
	gIBSMun.CreateElement("vDevTrib").SetText("0.00")                     // W44
	gIBSMun.CreateElement("vIBSMun").SetText("0.00")                      // W46

	gIBS.CreateElement("vIBS").SetText(formatDecimal(vIBS, 2))            // W47
	gIBS.CreateElement("vCredPres").SetText("0.00")                       // W48
	gIBS.CreateElement("vCredPresCondSus").SetText("0.00")                // W49

	gCBS := ibscbsTot.CreateElement("gCBS")
	gCBS.CreateElement("vDif").SetText("0.00")                            // W53
	gCBS.CreateElement("vDevTrib").SetText("0.00")                        // W54
	gCBS.CreateElement("vCBS").SetText(formatDecimal(vCBS, 2))            // W56
	gCBS.CreateElement("vCredPres").SetText("0.00")                       // W51
	gCBS.CreateElement("vCredPresCondSus").SetText("0.00")                // W52
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
		setSchemaText(transporta, "xNome", transp.Transportadora.XNome)
		if transp.Transportadora.IE != "" {
			transporta.CreateElement("IE").SetText(transp.Transportadora.IE)
		}
		setSchemaText(transporta, "xEnder", transp.Transportadora.Endereco)
		setSchemaText(transporta, "xMun", transp.Transportadora.Municipio)
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
		for _, v := range transp.Volumes {
			vol := transpElem.CreateElement("vol")
			vol.CreateElement("qVol").SetText(strconv.Itoa(v.QVol))
			if v.Esp != "" {
				setSchemaText(vol, "esp", v.Esp)
			}
			if v.Marca != "" {
				setSchemaText(vol, "marca", v.Marca)
			}
			if v.NVol != "" {
				setSchemaText(vol, "nVol", v.NVol)
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
	for _, det := range pag.Detalhes {
		detPag := pagElem.CreateElement("detPag")
		detPag.CreateElement("indPag").SetText(strconv.Itoa(pag.IndPag))
		detPag.CreateElement("tPag").SetText(det.TPag)
		// SEFAZ requires vPag=0.00 when tPag=90 (Sem pagamento), actual value otherwise
		if det.TPag == "90" {
			detPag.CreateElement("vPag").SetText("0.00")
		} else {
			detPag.CreateElement("vPag").SetText(formatDecimal(det.VPag, 2))
		}
	}
}

func (b *Builder) buildInfAdic(parent *etree.Element, info string) {
	infAdic := parent.CreateElement("infAdic")
	setSchemaText(infAdic, "infCpl", info)
}

func formatDecimal(d decimal.Decimal, places int32) string {
	return d.Truncate(places).StringFixed(places)
}
