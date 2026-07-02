package service

import (
	"bytes"
	"fmt"
	"image/png"
	"path/filepath"
	"runtime"
	"strings"

	"armazenda/pkg/nfe/entity"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/shopspring/decimal"
	"github.com/signintech/gopdf"
)

func getFontPath(bold bool) string {
	_, filename, _, _ := runtime.Caller(0)
	// filename is .../pkg/nfe/service/danfe.go
	// Go up 3 directories to reach project root
	dir := filepath.Dir(filename)
	for i := 0; i < 3; i++ {
		dir = filepath.Dir(dir)
	}
	name := "LiberationSans-Regular.ttf"
	if bold {
		name = "LiberationSans-Bold.ttf"
	}
	return filepath.Join(dir, "assets", "fonts", name)
}

// DANFEGenerator generates DANFE PDFs.
type DANFEGenerator struct{}

// NewDANFEGenerator creates a new DANFE generator.
func NewDANFEGenerator() *DANFEGenerator {
	return &DANFEGenerator{}
}

// Generate creates a DANFE PDF and returns it as bytes.
func (g *DANFEGenerator) Generate(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, false)
}

// GeneratePreview creates a preview DANFE PDF with a watermark banner.
func (g *DANFEGenerator) GeneratePreview(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, true)
}

func (g *DANFEGenerator) generatePDF(data entity.DANFEData, isPreview bool) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	const margin = 10.0
	const pageW = 595.0
	const pageH = 842.0
	const right = pageW - margin
	const usableW = pageW - margin*2

	fontPath := getFontPath(false)
	if err := pdf.AddTTFFont("liberation", fontPath); err != nil {
		return nil, fmt.Errorf("failed to add regular font: %w", err)
	}
	boldFontPath := getFontPath(true)
	if err := pdf.AddTTFFont("liberation-bold", boldFontPath); err != nil {
		return nil, fmt.Errorf("failed to add bold font: %w", err)
	}

	y := margin

	if isPreview {
		pdf.SetFillColor(255, 200, 200)
		pdf.Rectangle(margin, y, right, y+24, "FD", 0.5, 0)
		pdf.SetTextColor(180, 0, 0)
		pdf.SetFont("liberation-bold", "", 12)
		pdf.SetXY(margin+4, y+8)
		pdf.Cell(nil, "DOCUMENTO DE PRÉ-VISUALIZAÇÃO — SEM VALOR FISCAL")
		pdf.SetTextColor(0, 0, 0)
		y += 28
	}

	// Section 1: Canhoto / Receipt + DANFE header
	y = g.drawHeaderBlock(&pdf, data, margin, y, usableW)

	// Section 2: Access Key + Barcode
	y = g.drawAccessKeyBlock(&pdf, data, margin, y, usableW)

	// Section 3: Natureza da Operação
	y = g.drawNaturezaOp(&pdf, data, margin, y, usableW)

	// Section 4: Emitente
	y = g.drawEmitente(&pdf, data, margin, y, usableW)

	// Section 5: Destinatário
	y = g.drawDestinatario(&pdf, data, margin, y, usableW)

	// Section 6: Cálculo do Imposto
	y = g.drawTaxCalc(&pdf, data, margin, y, usableW)

	// Section 7: Transportador / Volumes
	y = g.drawTransport(&pdf, data, margin, y, usableW)

	// Section 8: Dados do Produto
	y = g.drawProductTable(&pdf, data, margin, y, usableW)

	// Section 9: Total da Nota
	y = g.drawTotalNota(&pdf, data, margin, y, usableW)

	// Section 10: ISSQN (conditional)
	if data.VISSQN.GreaterThan(decimal.Zero) {
		y = g.drawISSQN(&pdf, data, margin, y, usableW)
	}

	// Section 11: Dados Adicionais + Reservado ao Fisco
	y = g.drawAdditionalInfo(&pdf, data, margin, y, usableW)

	// Section 12: Protocolo
	if data.Protocol != "" {
		y = g.drawProtocol(&pdf, data, margin, y, usableW)
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// --- Layout helpers ---

func (g *DANFEGenerator) box(pdf *gopdf.GoPdf, x, y, w, h float64) {
	pdf.Rectangle(x, y, x+w, y+h, "D", 0.5, 0)
}

func (g *DANFEGenerator) labelValue(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string, labelSize, valueSize int) {
	pdf.SetFont("liberation", "", labelSize)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, label)
	if value != "" {
		pdf.SetFont("liberation", "", valueSize)
		pdf.SetXY(x+4, y+h-8)
		pdf.Cell(nil, value)
	}
}

func (g *DANFEGenerator) cell(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string) {
	g.box(pdf, x, y, w, h)
	g.labelValue(pdf, x, y, w, h, label, value, 6, 8)
}

func (g *DANFEGenerator) cellBoldValue(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string) {
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, label)
	if value != "" {
		pdf.SetFont("liberation-bold", "", 8)
		pdf.SetXY(x+4, y+h-9)
		pdf.Cell(nil, value)
	}
}

func (g *DANFEGenerator) drawBarcode(pdf *gopdf.GoPdf, key string, x, y, w, h float64) error {
	if key == "" {
		return fmt.Errorf("empty barcode key")
	}
	bcRaw, err := code128.EncodeWithColor(key, barcode.ColorScheme8)
	if err != nil {
		return err
	}
	bcScaled, err := barcode.Scale(bcRaw, int(w*3), int(h*3))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, bcScaled); err != nil {
		return err
	}
	img, err := gopdf.ImageHolderByBytes(buf.Bytes())
	if err != nil {
		return err
	}
	return pdf.ImageByHolder(img, x, y, &gopdf.Rect{W: w, H: h})
}

func formatAccessKey(key string) string {
	var parts []string
	for i := 0; i < len(key); i += 4 {
		end := i + 4
		if end > len(key) {
			end = len(key)
		}
		parts = append(parts, key[i:end])
	}
	return strings.Join(parts, " ")
}

func fmtDec(d decimal.Decimal, places int32) string {
	if d.IsZero() {
		return "0,00"
	}
	return d.StringFixed(places)
}

func fmtDec0(d decimal.Decimal) string {
	return fmtDec(d, 0)
}

func fmtDec2(d decimal.Decimal) string {
	return fmtDec(d, 2)
}

func fmtDec3(d decimal.Decimal) string {
	return fmtDec(d, 3)
}

func fmtDec4(d decimal.Decimal) string {
	return fmtDec(d, 4)
}

// --- Sections ---

func (g *DANFEGenerator) drawHeaderBlock(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 48.0
	g.box(pdf, x, y, w, h)

	// Left: receipt text
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+4)
	pdf.Cell(nil, "RECEBEMOS DE")
	pdf.SetFont("liberation-bold", "", 7)
	pdf.SetXY(x+4, y+10)
	pdf.Cell(nil, truncate(data.EmitterName, 45))
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+18)
	pdf.Cell(nil, "OS PRODUTOS/SERVIÇOS CONSTANTES DA NOTA FISCAL INDICADA AO LADO")

	pdf.SetXY(x+4, y+18)
	pdf.Cell(nil, "DATA DE EMISSÃO")
	pdf.Cell(nil, data.EmissionDate)

	pdf.SetXY(x+4, y+26)
	pdf.Cell(nil, "DATA DO RECEBIMENTO: ______________________")
	pdf.SetXY(x+4, y+34)
	pdf.Cell(nil, "IDENTIFICAÇÃO E ASSINATURA DO RECEBEDOR: _________________________________")

	// Right: DANFE title block
	leftW := w * 0.55
	rightX := x + leftW
	pdf.Line(rightX, y, rightX, y+h)

	pdf.SetFont("liberation-bold", "", 14)
	pdf.SetXY(rightX+4, y+4)
	pdf.Cell(nil, "DANFE")
	pdf.SetFont("liberation", "", 7)
	pdf.SetXY(rightX+4, y+16)
	pdf.Cell(nil, "Documento Auxiliar da Nota Fiscal Eletrônica")
	pdf.SetXY(rightX+4, y+26)
	pdf.Cell(nil, "0 - ENTRADA")
	pdf.SetXY(rightX+4, y+34)
	pdf.Cell(nil, "1 - SAÍDA")

	pdf.SetLineWidth(.25)
	pdf.Rectangle(rightX+66, y+26, rightX+80, y+44, "D", 2, 10)
	pdf.SetLineWidth(.5)
	pdf.SetFont("liberation-bold", "", 10)
	pdf.SetXY(rightX+70, y+31)
	pdf.Cell(nil, "1")

	pdf.SetFont("liberation-bold", "", 10)
	pdf.SetXY(rightX+180, y+10)
	pdf.Cell(nil, fmt.Sprintf("Nº %d", data.Numero))
	pdf.SetXY(rightX+180, y+20)
	pdf.Cell(nil, fmt.Sprintf("SÉRIE %d", data.Serie))
	pdf.SetXY(rightX+180, y+30)
	pdf.Cell(nil, "Página 1/1")

	return y + h
}

func (g *DANFEGenerator) drawAccessKeyBlock(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 48.0
	g.box(pdf, x, y, w, h)

	// Barcode left side (~55%)
	barW := w * 0.55
	key := data.AccessKey
	if key == "" {
		key = "00000000000000000000000000000000000000000000"
	}
	if err := g.drawBarcode(pdf, key, x+4, y+4, barW-8, h-18); err != nil {
		fmt.Printf("\nbar code error: %s", err.Error())
		pdf.SetFont("liberation", "", 8)
		pdf.SetXY(x+4, y+18)
		pdf.Cell(nil, "[ERRO AO GERAR CÓDIGO DE BARRAS]")
	}
	// Key text below barcode
	keyText := formatAccessKey(key)
	pdf.SetFont("liberation", "", 7)
	textWidth, _ := pdf.MeasureTextWidth(keyText)
	centerX := x + 4 + (barW-8-textWidth)/2
	pdf.SetXY(centerX, y+h-12)
	pdf.Cell(nil, keyText)

	// Right side: key again + consulta text
	midX := x + barW
	pdf.Line(midX, y, midX, y+h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(midX+4, y+4)
	pdf.Cell(nil, "CHAVE DE ACESSO")
	pdf.SetFont("liberation-bold", "", 9)
	pdf.SetXY(midX+4, y+14)
	pdf.Cell(nil, formatAccessKey(key))
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(midX+4, y+28)
	pdf.Cell(nil, "Consulta de autenticidade no portal nacional da NF-e")
	pdf.SetXY(midX+4, y+36)
	pdf.Cell(nil, "www.nfe.fazenda.gov.br/portal ou no site da Sefaz Autorizadora")

	return y + h
}

func (g *DANFEGenerator) drawNaturezaOp(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 28.0
	g.cell(pdf, x, y, w, h, "NATUREZA DA OPERAÇÃO", data.NaturezaOp)
	return y + h
}

func (g *DANFEGenerator) drawEmitente(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 60.0
	g.box(pdf, x, y, w, h)

	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "EMITENTE")

	// Razão Social (full width, bold)
	pdf.SetFont("liberation-bold", "", 9)
	pdf.SetXY(x+4, y+11)
	pdf.Cell(nil, truncate(data.EmitterName, 80))

	// CNPJ | IE | CRT
	row1H := 20.0
	col1 := w * 0.35
	col2 := w * 0.35
	col3 := w - col1 - col2
	g.cell(pdf, x, y+22, col1, row1H, "CNPJ", formatCNPJ(data.EmitterCNPJ))
	g.cell(pdf, x+col1, y+22, col2, row1H, "INSCRIÇÃO ESTADUAL", data.EmitterIE)
	g.cell(pdf, x+col1+col2, y+22, col3, row1H, "REGIME TRIBUTÁRIO", crtLabel(data.EmitterCRT))

	// Endereço completo
	row2H := h - 22 - row1H
	addr := joinNonEmpty(
		data.EmitterAddress,
		data.EmitterNumber,
		data.EmitterComplement,
		data.EmitterNeighborhood,
		data.EmitterCEP,
		fmt.Sprintf("%s/%s", data.EmitterCity, data.EmitterUF),
		data.EmitterPhone,
	)
	g.cell(pdf, x, y+22+row1H, w, row2H, "ENDEREÇO", addr)

	return y + h
}

func (g *DANFEGenerator) drawDestinatario(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 60.0
	g.box(pdf, x, y, w, h)

	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DESTINATÁRIO / REMETENTE")

	pdf.SetFont("liberation-bold", "", 9)
	pdf.SetXY(x+4, y+11)
	pdf.Cell(nil, truncate(data.DestName, 80))

	row1H := 20.0
	col1 := w * 0.35
	col2 := w * 0.35
	col3 := w - col1 - col2
	g.cell(pdf, x, y+22, col1, row1H, "CNPJ/CPF", formatCNPJ(data.DestCNPJ))
	g.cell(pdf, x+col1, y+22, col2, row1H, "INSCRIÇÃO ESTADUAL", data.DestIE)
	g.cell(pdf, x+col1+col2, y+22, col3, row1H, "INDICADOR IE DEST.", indIEDestLabel(data.DestIndIEDest))

	row2H := h - 22 - row1H
	addr := joinNonEmpty(
		data.DestAddress,
		data.DestNumber,
		data.DestComplement,
		data.DestNeighborhood,
		data.DestCEP,
		fmt.Sprintf("%s/%s", data.DestCity, data.DestUF),
		data.DestPhone,
	)
	g.cell(pdf, x, y+22+row1H, w, row2H, "ENDEREÇO", addr)

	return y + h
}

func (g *DANFEGenerator) drawTaxCalc(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 34.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "CÁLCULO DO IMPOSTO")

	fields := []struct {
		label string
		value string
	}{
		{"B.CÁLC. ICMS", fmtDec2(data.VBC)},
		{"V. ICMS", fmtDec2(data.VICMS)},
		{"B.CÁLC. ST", fmtDec2(data.VBCST)},
		{"V. ICMS ST", fmtDec2(data.VST)},
		{"V. PRODUTOS", fmtDec2(data.TotalValue)},
		{"V. FRETE", fmtDec2(data.VFrete)},
		{"V. SEGURO", fmtDec2(data.VSeg)},
		{"DESCONTO", fmtDec2(data.VDesc)},
		{"V. IPI", fmtDec2(data.VIPI)},
		{"OUT.DESP.", fmtDec2(data.VOutro)},
		{"V. NF", fmtDec2(data.TotalValue)},
		{"V.EST.TRI.", fmtDec2(data.VTotTrib)},
	}
	colW := w / float64(len(fields))
	for i, f := range fields {
		fx := x + float64(i)*colW
		g.cell(pdf, fx, y+10, colW, h-10, f.label, f.value)
	}
	return y + h
}

func (g *DANFEGenerator) drawTransport(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 60.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "TRANSPORTADOR / VOLUMES TRANSPORTADOS")

	row1H := 18.0
	row2H := 14.0
	row3H := 18.0

	// Row 1
	c1 := w * 0.12
	c2 := w * 0.30
	c3 := w * 0.20
	c4 := w * 0.18
	c5 := w - c1 - c2 - c3 - c4
	g.cell(pdf, x, y+10, c1, row1H, "FRETE POR CONTA", modFreteLabel(data.ModFrete))
	g.cell(pdf, x+c1, y+10, c2, row1H, "RAZÃO SOCIAL / NOME", data.TranspName)
	g.cell(pdf, x+c1+c2, y+10, c3, row1H, "CNPJ/CPF", formatCNPJ(data.TranspCNPJ))
	g.cell(pdf, x+c1+c2+c3, y+10, c4, row1H, "INSCRIÇÃO ESTADUAL", data.TranspIE)
	g.cell(pdf, x+c1+c2+c3+c4, y+10, c5, row1H, "UF", data.TranspUF)

	// Row 2
	d1 := w * 0.40
	d2 := w * 0.35
	d3 := w - d1 - d2
	g.cell(pdf, x, y+10+row1H, d1, row2H, "ENDEREÇO", data.TranspAddress)
	g.cell(pdf, x+d1, y+10+row1H, d2, row2H, "MUNICÍPIO", data.TranspCity)
	g.cell(pdf, x+d1+d2, y+10+row1H, d3, row2H, "UF", data.TranspUF)

	// Row 3 (volumes)
	v1 := w * 0.12
	v2 := w * 0.15
	v3 := w * 0.15
	v4 := w * 0.15
	v5 := w * 0.18
	v6 := w - v1 - v2 - v3 - v4 - v5
	g.cell(pdf, x, y+10+row1H+row2H, v1, row3H, "QUANTIDADE", data.QVol)
	g.cell(pdf, x+v1, y+10+row1H+row2H, v2, row3H, "ESPÉCIE", data.Esp)
	g.cell(pdf, x+v1+v2, y+10+row1H+row2H, v3, row3H, "MARCA", data.Marca)
	g.cell(pdf, x+v1+v2+v3, y+10+row1H+row2H, v4, row3H, "NUMERAÇÃO", data.NVol)
	g.cell(pdf, x+v1+v2+v3+v4, y+10+row1H+row2H, v5, row3H, "PESO LÍQ. (KG)", fmtDec3(data.PesoL))
	g.cell(pdf, x+v1+v2+v3+v4+v5, y+10+row1H+row2H, v6, row3H, "PESO BRUTO (KG)", fmtDec3(data.PesoB))

	return y + h
}

func (g *DANFEGenerator) drawProductTable(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	headerH := 22.0
	rowH := 20.0
	h := headerH + rowH*float64(len(data.Products))
	if h < headerH+rowH {
		h = headerH + rowH
	}

	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DADOS DO PRODUTO / SERVIÇO")

	// Header
	cols := []struct {
		w     float64
		label string
	}{
		{w * 0.08, "CÓDIGO"},
		{w * 0.22, "DESCRIÇÃO DO PRODUTO / SERVIÇO"},
		{w * 0.07, "NCM"},
		{w * 0.05, "CST"},
		{w * 0.05, "CFOP"},
		{w * 0.05, "UN"},
		{w * 0.07, "QTD."},
		{w * 0.08, "V. UNIT."},
		{w * 0.09, "V. TOTAL"},
		{w * 0.07, "BC ICMS"},
		{w * 0.07, "V. ICMS"},
		{w * 0.05, "AL. ICMS"},
		{w * 0.05, "V. IPI"},
	}
	cy := y + headerH - 8
	cx := x
	for _, c := range cols {
		pdf.SetXY(cx+4, cy)
		pdf.Cell(nil, c.label)
		cx += c.w
	}
	pdf.Line(x, y+headerH-3, x+w, y+headerH-3)

	// Rows
	for i, p := range data.Products {
		cy := y + headerH + float64(i)*rowH
		cx := x
		vals := []string{
			p.Code,
			truncate(p.Desc, 30),
			p.NCM,
			p.CST,
			p.CFOP,
			p.Unit,
			fmtDec3(p.Quantity),
			fmtDec4(p.UnitPrice),
			fmtDec2(p.Total),
			fmtDec2(p.VBC),
			fmtDec2(p.VICMS),
			fmtDec2(p.PICMS),
			fmtDec2(p.VIPI),
		}
		for j, c := range cols {
			pdf.SetXY(cx+4, cy+5)
			pdf.Cell(nil, vals[j])
			cx += c.w
		}
	}

	return y + h
}

func (g *DANFEGenerator) drawTotalNota(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 34.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DADOS DOS TOTAIS DA NOTA")

	fields := []struct {
		label string
		value string
	}{
		{"B.CÁLC. ICMS", fmtDec2(data.VBC)},
		{"V. ICMS", fmtDec2(data.VICMS)},
		{"B.CÁLC. ST", fmtDec2(data.VBCST)},
		{"V. ICMS ST", fmtDec2(data.VST)},
		{"V. PRODUTOS", fmtDec2(data.TotalValue)},
		{"V. FRETE", fmtDec2(data.VFrete)},
		{"V. SEGURO", fmtDec2(data.VSeg)},
		{"DESCONTO", fmtDec2(data.VDesc)},
		{"V. IPI", fmtDec2(data.VIPI)},
		{"OUT.DESP.", fmtDec2(data.VOutro)},
		{"V. NF", fmtDec2(data.TotalValue)},
		{"V.EST.TRI.", fmtDec2(data.VTotTrib)},
	}
	colW := w / float64(len(fields))
	for i, f := range fields {
		fx := x + float64(i)*colW
		g.cell(pdf, fx, y+10, colW, h-10, f.label, f.value)
	}
	return y + h
}

func (g *DANFEGenerator) drawISSQN(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 34.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "CÁLCULO DO ISSQN")

	cols := w / 5.0
	g.cell(pdf, x, y+10, cols, h-10, "IM", "")
	g.cell(pdf, x+cols, y+10, cols, h-10, "VALOR TOTAL DOS SERVIÇOS", fmtDec2(data.TotalValue))
	g.cell(pdf, x+cols*2, y+10, cols, h-10, "BASE DE CÁLCULO DO ISSQN", fmtDec2(data.VBCISSQN))
	g.cell(pdf, x+cols*3, y+10, cols, h-10, "VALOR DO ISSQN", fmtDec2(data.VISSQN))
	g.cell(pdf, x+cols*4, y+10, w-cols*4, h-10, "VALOR TOTAL DOS SERVIÇOS", fmtDec2(data.TotalValue))
	return y + h
}

func (g *DANFEGenerator) drawAdditionalInfo(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 56.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DADOS ADICIONAIS")

	leftW := w * 0.60
	pdf.Line(x+leftW, y, x+leftW, y+h)

	pdf.SetXY(x+4, y+10)
	pdf.Cell(nil, "INFORMAÇÕES COMPLEMENTARES")
	pdf.SetFont("liberation", "", 7)
	pdf.SetXY(x+4, y+20)
	pdf.Cell(nil, truncate(data.InfCpl, 200))

	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+leftW+4, y+10)
	pdf.Cell(nil, "RESERVADO AO FISCO")
	pdf.SetFont("liberation", "", 7)
	pdf.SetXY(x+leftW+4, y+20)
	pdf.Cell(nil, truncate(data.InfAdFisco, 120))

	return y + h
}

func (g *DANFEGenerator) drawProtocol(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 28.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("liberation", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "PROTOCOLO DE AUTORIZAÇÃO DE USO")

	c1 := w * 0.40
	c2 := w * 0.30
	c3 := w - c1 - c2
	g.cellBoldValue(pdf, x, y+10, c1, h-10, "NÚMERO DO PROTOCOLO", data.Protocol)
	g.cellBoldValue(pdf, x+c1, y+10, c2, h-10, "DATA DE AUTORIZAÇÃO", data.ProtocolDate)
	if data.XMotivo != "" {
		g.cellBoldValue(pdf, x+c1+c2, y+10, c3, h-10, "STATUS", fmt.Sprintf("%s - %s", data.CStat, data.XMotivo))
	} else {
		g.cell(pdf, x+c1+c2, y+10, c3, h-10, "STATUS", "")
	}

	return y + h
}

// --- Formatting utilities ---

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func joinNonEmpty(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ", ")
}

func formatCNPJ(cnpj string) string {
	if len(cnpj) == 14 {
		return fmt.Sprintf("%s.%s.%s/%s-%s", cnpj[:2], cnpj[2:5], cnpj[5:8], cnpj[8:12], cnpj[12:])
	}
	return cnpj
}

func crtLabel(crt string) string {
	switch crt {
	case "1":
		return "Simples Nacional"
	case "2":
		return "Simples Nacional - excesso de sublimite"
	case "3":
		return "Regime Normal"
	default:
		return crt
	}
}

func indIEDestLabel(ind string) string {
	switch ind {
	case "1":
		return "Contribuinte ICMS"
	case "2":
		return "Contribuinte isento"
	case "9":
		return "Não Contribuinte"
	default:
		return ind
	}
}

func modFreteLabel(mod string) string {
	switch mod {
	case "0":
		return "0 - Emitente"
	case "1":
		return "1 - Destinatário"
	case "2":
		return "2 - Terceiros"
	case "3":
		return "3 - Próprio emitente"
	case "4":
		return "4 - Próprio destinatário"
	case "9":
		return "9 - Sem frete"
	default:
		return mod
	}
}
