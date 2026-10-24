package service

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	// Internal - Config
	"armazenda/pkg/nfe/config"

	// Internal - Entities
	"armazenda/pkg/nfe/entity"

	// Third-party
	"github.com/shopspring/decimal"
	"github.com/signintech/gopdf"
)

// getFont returns the embedded TTF for the DANFE. Fonts are embedded via
// config.FontsFS so no runtime filesystem lookup is needed.
func getFont(bold bool) ([]byte, error) {
	name := "fonts/LiberationSerif-Regular.ttf"
	if bold {
		name = "fonts/LiberationSerif-Bold.ttf"
	}
	fontBytes, readErr := config.FontsFS.ReadFile(name)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read embedded font %s: %w", name, readErr)
	}
	return fontBytes, nil
}

// DANFEGenerator generates DANFE PDFs.
type DANFEGenerator struct{}

// NewDANFEGenerator creates a new DANFE generator.
func NewDANFEGenerator() *DANFEGenerator {
	return &DANFEGenerator{}
}

// Generate creates a single-page DANFE PDF and returns it as bytes.
func (g *DANFEGenerator) Generate(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, "", true)
}

// GeneratePreview creates a single-page preview DANFE PDF with a watermark banner.
func (g *DANFEGenerator) GeneratePreview(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, "DOCUMENTO DE PRÉ-VISUALIZAÇÃO — SEM VALOR FISCAL", true)
}

// GenerateCancelled creates a single-page DANFE PDF for a cancelled NF-e with a
// prominent "NF-e CANCELADA" banner, as required by MOC Anexo II.
func (g *DANFEGenerator) GenerateCancelled(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, "NF-e CANCELADA", true)
}

// GenerateMultiPage creates a DANFE PDF that may span multiple pages when the
// content exceeds a single A4 sheet. This is the entry point for long invoices;
// Generate remains strictly single-page.
func (g *DANFEGenerator) GenerateMultiPage(data entity.DANFEData) ([]byte, error) {
	return g.generatePDF(data, "", false)
}

// generatePDF renders the DANFE. When banner is non-empty, a highlighted
// banner with the given text is drawn at the top of the page. When singlePage
// is false the renderer inserts page breaks between sections as needed.
func (g *DANFEGenerator) generatePDF(data entity.DANFEData, banner string, singlePage bool) ([]byte, error) {
	// Per MOC Anexo II: FS-DA (tpEmis=5) and EPEC (tpEmis=4) are reserved
	// and not actively wired (see AGENTS.md). Return a clear error so the
	// caller can surface it instead of silently producing a non-compliant DANFE.
	switch data.TpEmis {
	case "4":
		return nil, fmt.Errorf("DANFE para contingência EPEC (tpEmis=4) não implementado")
	case "5":
		return nil, fmt.Errorf("DANFE para contingência FS-DA (tpEmis=5) não implementado")
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	const margin = 10.0
	const pageW = 595.0
	const pageH = 842.0
	const right = pageW - margin
	const usableW = pageW - margin*2

	regularFont, fontErr := getFont(false)
	if fontErr != nil {
		return nil, fmt.Errorf("failed to add regular font: %w", fontErr)
	}
	if err := pdf.AddTTFFontByReader("serif", bytes.NewReader(regularFont)); err != nil {
		return nil, fmt.Errorf("failed to add regular font: %w", err)
	}
	boldFont, boldErr := getFont(true)
	if boldErr != nil {
		return nil, fmt.Errorf("failed to add bold font: %w", boldErr)
	}
	if err := pdf.AddTTFFontByReader("serif-bold", bytes.NewReader(boldFont)); err != nil {
		return nil, fmt.Errorf("failed to add bold font: %w", err)
	}

	// Per MOC Anexo II §3.10.5: copy vICMSDeson into infCpl so it appears
	// on the DANFE (no dedicated field in the leiaute).
	if data.VICMSDeson.GreaterThan(decimal.Zero) {
		desonLine := fmt.Sprintf(" Valor ICMS Desonerado: %s.", fmtDec2(data.VICMSDeson))
		if data.InfCpl != "" {
			data.InfCpl += "\n" + desonLine
		} else {
			data.InfCpl = strings.TrimSpace(desonLine)
		}
	}

	// Per MOC Anexo II §3: homologação DANFE must contain "SEM VALOR FISCAL"
	// in Informações Complementares or as a watermark.
	isHomologation := data.TpAmb == "2"
	if isHomologation {
		svf := "SEM VALOR FISCAL"
		if data.InfCpl != "" {
			data.InfCpl = svf + " " + data.InfCpl
		} else {
			data.InfCpl = svf
		}
	}

	y := margin

	if banner != "" {
		pdf.SetFillColor(255, 200, 200)
		pdf.SetLineWidth(0)
		pdf.Rectangle(margin, y, right, y+24, "FD", 0.5, 0)
		pdf.SetTextColor(180, 0, 0)
		pdf.SetFont("serif-bold", "", 12)
		pdf.SetXY(margin+4, y+8)
		pdf.Cell(nil, banner)
		pdf.SetTextColor(0, 0, 0)
		y += 28
	}

	// Section 1: Canhoto / Receipt + DANFE header
	y = g.drawHeaderBlock(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 1b: Contingency banner (SVC only)
	if data.TpEmis == "6" || data.TpEmis == "7" {
		y = g.drawContingencyBanner(&pdf, data, margin, y, usableW)
		y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)
	}

	// Section 2: Access Key + Barcode + Campo 1/2
	y = g.drawAccessKeyBlock(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 3: Natureza da Operação
	y = g.drawNaturezaOp(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 4: Emitente
	y = g.drawEmitente(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 5: Destinatário
	y = g.drawDestinatario(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 6: Cálculo do Imposto
	y = g.drawTaxCalc(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 7: Transportador / Volumes
	y = g.drawTransport(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 8: Dados do Produto
	y = g.drawProductTable(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 9: Total da Nota
	y = g.drawTotalNota(&pdf, data, margin, y, usableW)
	y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)

	// Section 10: ISSQN (conditional)
	if data.VISSQN.GreaterThan(decimal.Zero) {
		y = g.drawISSQN(&pdf, data, margin, y, usableW)
		y = g.checkPageBreak(&pdf, y, pageH, margin, singlePage)
	}

	// Section 11: Dados Adicionais + Reservado ao Fisco
	y = g.drawAdditionalInfo(&pdf, data, margin, y, usableW)

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// --- Layout helpers ---

func (g *DANFEGenerator) box(pdf *gopdf.GoPdf, x, y, w, h float64) {
	pdf.Rectangle(x, y, x+w, y+h, "D", 3, 10)
}

// checkPageBreak adds a new page and resets y when the current y position is
// near the bottom of the page. It is a no-op when singlePage is true.
func (g *DANFEGenerator) checkPageBreak(pdf *gopdf.GoPdf, y, pageH, margin float64, singlePage bool) float64 {
	if singlePage {
		return y
	}
	if y > pageH-margin-20 {
		pdf.AddPage()
		return margin
	}
	return y
}

// labelValue renders a label (6pt) and value at the given size inside a boxed cell.
// Per MOC Anexo II §3.7.3, field headers are 6pt minimum, uppercase.
// Per §3.7.9, field content is 10pt minimum for "demais campos".
func (g *DANFEGenerator) labelValue(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string, labelSize, valueSize int) {
	pdf.SetFont("serif", "", labelSize)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, label)
	if value != "" {
		pdf.SetFont("serif", "", valueSize)
		pdf.SetXY(x+4, y+h-10)
		pdf.Cell(nil, value)
	}
}

// cell renders a boxed cell with label (6pt) and value (10pt) — for "demais campos".
func (g *DANFEGenerator) cell(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string) {
	g.box(pdf, x, y, w, h)
	g.labelValue(pdf, x, y, w, h, label, value, 6, 10)
}

// cellEmit renders a boxed cell with label (6pt) and value (8pt) — for the
// emitente block per §3.7.6 (address, CNPJ, IE etc. minimum 8pt).
func (g *DANFEGenerator) cellEmit(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string) {
	g.box(pdf, x, y, w, h)
	g.labelValue(pdf, x, y, w, h, label, value, 6, 8)
}

func (g *DANFEGenerator) cellBoldValue(pdf *gopdf.GoPdf, x, y, w, h float64, label, value string) {
	g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, label)
	if value != "" {
		pdf.SetFont("serif-bold", "", 10)
		pdf.SetXY(x+4, y+h-11)
		pdf.Cell(nil, value)
	}
}

// drawImageFit renders image bytes scaled to fit within (maxW, maxH) while
// preserving aspect ratio, centered in the available area.
func (g *DANFEGenerator) drawImageFit(pdf *gopdf.GoPdf, imgBytes []byte, x, y, maxW, maxH float64) error {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return err
	}
	if cfg.Width == 0 || cfg.Height == 0 {
		return fmt.Errorf("invalid image dimensions")
	}
	imgW := float64(cfg.Width)
	imgH := float64(cfg.Height)
	scale := maxW / imgW
	if maxH/imgH < scale {
		scale = maxH / imgH
	}
	w := imgW * scale
	h := imgH * scale
	ix := x + (maxW-w)/2
	iy := y + (maxH-h)/2

	holder, err := gopdf.ImageHolderByBytes(imgBytes)
	if err != nil {
		return err
	}
	return pdf.ImageByHolder(holder, ix, iy, &gopdf.Rect{W: w, H: h})
}

// drawBarcode renders a strict CODE-128C barcode for the given numeric key
// using the custom encoder (per MOC Anexo II §2). Dimensions must satisfy
// the minimums: width ≥ 6cm (≈170pt for laser), height ≥ 0.8cm (≈23pt).
func (g *DANFEGenerator) drawBarcode(pdf *gopdf.GoPdf, key string, x, y, w, h float64) error {
	if key == "" {
		return fmt.Errorf("empty barcode key")
	}
	// Convert pt dimensions to pixels (1pt ≈ 1.333px at 96dpi).
	pxW := int(w * 3)
	pxH := int(h * 3)
	pngBytes, err := Code128CBarcode(key, pxW, pxH)
	if err != nil {
		return err
	}
	img, err := gopdf.ImageHolderByBytes(pngBytes)
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
		if places <= 0 {
			return "0"
		}
		return "0," + strings.Repeat("0", int(places))
	}
	s := d.StringFixed(places)
	parts := strings.SplitN(s, ".", 2)
	intPart := formatThousands(parts[0])
	if len(parts) == 1 {
		return intPart
	}
	return intPart + "," + parts[1]
}

// formatThousands inserts a dot every three digits in the integer part of a
// number (Brazilian grouping separator). It preserves a leading minus sign.
func formatThousands(intPart string) string {
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}
	n := len(intPart)
	if n <= 3 {
		if negative {
			return "-" + intPart
		}
		return intPart
	}
	var groups []string
	for n > 3 {
		groups = append([]string{intPart[n-3 : n]}, groups...)
		intPart = intPart[:n-3]
		n = len(intPart)
	}
	groups = append([]string{intPart}, groups...)
	result := strings.Join(groups, ".")
	if negative {
		return "-" + result
	}
	return result
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
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+4)
	pdf.Cell(nil, "RECEBEMOS DE")
	pdf.SetFont("serif-bold", "", 8)
	//pdf.SetXY(x+44, y+4)
	// Wrap emitter name within the canhoto width (§3.1: represent full content)
	canhotoW := w*0.55 - 8
	pdf.SetFont("serif-bold", "", 6)
	canhotoLines := wrapText(pdf, strings.ToUpper(data.EmitterName), canhotoW)
	canhotoY := y + 4
	for _, line := range canhotoLines {
		pdf.SetXY(x+54, canhotoY)
		pdf.Cell(nil, line)
		canhotoY += 7
	}
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+12)
	pdf.Cell(nil, "OS PRODUTOS/SERVIÇOS CONSTANTES DA NOTA FISCAL INDICADA AO LADO")

	pdf.SetXY(x+4, y+20)
	pdf.Cell(nil, "DATA DE EMISSÃO:")
	pdf.SetXY(x+64, y+20)
	pdf.Cell(nil, data.EmissionDate)

	pdf.SetXY(x+4, y+28)
	pdf.Cell(nil, "DATA DO RECEBIMENTO: ______________________")
	pdf.SetXY(x+4, y+36)
	pdf.Cell(nil, "IDENTIFICAÇÃO E ASSINATURA DO RECEBEDOR: _________________________________")

	// Right: keep only NF-e identification in the first row (cropped on sign)
	leftW := w * 0.55
	rightX := x + leftW
	pdf.Line(rightX, y, rightX, y+h)

	pdf.SetFont("serif-bold", "", 12)
	pdf.SetXY(rightX+4, y+4)
	pdf.Cell(nil, "NF-e")

	pdf.SetFont("serif-bold", "", 10)
	pdf.SetXY(rightX+4, y+18)
	pdf.Cell(nil, fmt.Sprintf("Nº %d", data.Numero))
	pdf.SetXY(rightX+4, y+30)
	pdf.Cell(nil, fmt.Sprintf("SÉRIE %d", data.Serie))

	return y + h + 4
}

// drawContingencyBanner renders a highlighted contingency header per MOC
// Anexo IV (SVC-AN / SVC-RS only — FS-DA and EPEC are stubbed).
func (g *DANFEGenerator) drawContingencyBanner(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	const minH = 20.0
	const infoStartY = 12.0
	const lineHeight = 7.0
	const padding = 4.0

	svcName := "SVC"
	switch data.TpEmis {
	case "6":
		svcName = "SVC-AN"
	case "7":
		svcName = "SVC-RS"
	}

	contInfo := ""
	if data.DhCont != "" {
		contInfo = "Entrada em contingência: " + data.DhCont
	}
	if data.XJust != "" {
		if contInfo != "" {
			contInfo += " — "
		}
		contInfo += "Justificativa: " + data.XJust
	}

	var contLines []string
	if contInfo != "" {
		pdf.SetFont("serif", "", 8)
		contLines = wrapText(pdf, contInfo, w-8)
	}

	infoH := float64(len(contLines))*lineHeight + padding
	h := infoStartY + infoH + padding
	if h < minH {
		h = minH
	}

	pdf.SetFillColor(255, 240, 200)
	pdf.Rectangle(x, y, x+w, y+h, "FD", 0.5, 0)

	pdf.SetFont("serif-bold", "", 10)
	pdf.SetTextColor(180, 80, 0)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, fmt.Sprintf("EMITIDA EM CONTINGÊNCIA %s", svcName))
	pdf.SetTextColor(0, 0, 0)

	if len(contLines) > 0 {
		pdf.SetFont("serif", "", 8)
		contY := y + infoStartY
		for _, line := range contLines {
			pdf.SetXY(x+4, contY)
			pdf.Cell(nil, line)
			contY += lineHeight
		}
	}

	return y + h + 4
}

func (g *DANFEGenerator) drawAccessKeyBlock(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 70.0
	g.box(pdf, x, y, w, h)

	col1W := w * 0.20
	col2W := w * 0.40
	col3W := w - col1W - col2W
	col1X := x
	col2X := x + col1W
	col3X := x + col1W + col2W

	pdf.Line(col2X, y, col2X, y+h)
	pdf.Line(col3X, y, col3X, y+h)

	// Column 1: emitter logo (user upload or Armazenda default)
	logoBytes := data.LogoData
	if len(logoBytes) == 0 {
		if b, err := config.DefaultLogoFS.ReadFile("default_logo.png"); err == nil {
			logoBytes = b
		}
	}
	if len(logoBytes) > 0 {
		_ = g.drawImageFit(pdf, logoBytes, col1X+4, y+4, col1W-8, h-8)
	}

	// Column 2: information moved from the first row
	pdf.SetFont("serif-bold", "", 12)
	pdf.SetXY(col2X+4, y+4)
	pdf.Cell(nil, "DANFE")
	pdf.SetFont("serif", "", 8)
	pdf.SetXY(col2X+4, y+16)
	pdf.Cell(nil, "Documento Auxiliar da Nota Fiscal Eletrônica")

	pdf.SetXY(col2X+4, y+26)
	pdf.Cell(nil, "0 - ENTRADA")
	pdf.SetXY(col2X+4, y+34)
	pdf.Cell(nil, "1 - SAÍDA")

	tpNF := data.TpNF
	if tpNF == "" {
		tpNF = "1"
	}
	pdf.SetLineWidth(0)
	pdf.Rectangle(col2X+66, y+26, col2X+80, y+44, "D", 2, 10)
	pdf.SetLineWidth(0)
	pdf.SetFont("serif-bold", "", 10)
	pdf.SetXY(col2X+70, y+31)
	pdf.Cell(nil, tpNF)

	pdf.SetFont("serif-bold", "", 10)
	pdf.SetXY(col2X+4, y+46)
	pdf.Cell(nil, fmt.Sprintf("Nº %d", data.Numero))
	pdf.SetXY(col2X+4, y+56)
	pdf.Cell(nil, fmt.Sprintf("SÉRIE %d", data.Serie))
	pdf.SetXY(col2X+4, y+64)
	// TODO(P1): multi-page support — currently single page only
	pdf.Cell(nil, "FOLHA 1/1")

	// Column 3: barcode + existing access-key/protocol info
	barW := col3W * 0.55
	key := data.AccessKey
	if key == "" {
		key = "00000000000000000000000000000000000000000000"
	}
	if err := g.drawBarcode(pdf, key, col3X+4, y+4, barW-8, h-46); err != nil {
		pdf.SetFont("serif", "", 8)
		pdf.SetXY(col3X+4, y+18)
		pdf.Cell(nil, "[ERRO AO GERAR CÓDIGO DE BARRAS]")
	}
	keyText := formatAccessKey(key)
	pdf.SetFont("serif-bold", "", 8)
	textWidth, _ := pdf.MeasureTextWidth(keyText)
	centerX := col3X + 4 + (barW-8-textWidth)/2
	pdf.SetXY(centerX, y+h-36)
	pdf.Cell(nil, keyText)

	midX := col3X + barW
	pdf.Line(midX, y, midX, y+h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(midX+4, y+4)
	pdf.Cell(nil, "CHAVE DE ACESSO")
	pdf.SetFont("serif-bold", "", 9)
	pdf.SetXY(midX+4, y+12)
	pdf.Cell(nil, formatAccessKey(key))

	pdf.SetFont("serif", "", 7)
	pdf.SetXY(midX+4, y+24)
	pdf.Cell(nil, "Consulta de autenticidade no portal nacional da NF-e")
	pdf.SetXY(midX+4, y+32)
	pdf.Cell(nil, "www.nfe.fazenda.gov.br/portal")

	pdf.SetFont("serif", "", 6)
	pdf.SetXY(midX+4, y+44)
	pdf.Cell(nil, "PROTOCOLO DE AUTORIZAÇÃO DE USO")
	if data.Protocol != "" {
		pdf.SetFont("serif-bold", "", 10)
		pdf.SetXY(midX+4, y+52)
		protText := data.Protocol
		if data.ProtocolDate != "" {
			protText += " " + data.ProtocolDate
		}
		pdf.Cell(nil, protText)
	} else {
		pdf.SetFont("serif", "", 8)
		pdf.SetXY(midX+4, y+52)
		pdf.Cell(nil, "(sem autorização)")
	}

	pdf.Line(x, y+h-18, x+w, y+h-18)

	return y + h + 4
}

func (g *DANFEGenerator) drawNaturezaOp(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 24.0
	g.cell(pdf, x, y, w, h, "NATUREZA DA OPERAÇÃO", data.NaturezaOp)
	return y + h + 4
}

func (g *DANFEGenerator) drawEmitente(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	const row1H = 24.0
	const row2H = 23.0

	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "EMITENTE")

	// Razão Social (12pt bold per §3.7.6). Measure and wrap using the same
	// font that will be used for rendering so the box height grows correctly.
	nameBoxTop := y + 10
	nameLabelY := nameBoxTop + 2
	nameValueStartY := nameBoxTop + 12
	nameLineHeight := 10.0

	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, nameLabelY)
	pdf.Cell(nil, "NOME / RAZÃO SOCIAL")

	pdf.SetFont("serif-bold", "", 12)
	nameLines := wrapText(pdf, data.EmitterName, w-8)
	nameY := nameValueStartY
	for _, line := range nameLines {
		pdf.SetXY(x+4, nameY)
		pdf.Cell(nil, line)
		nameY += nameLineHeight
	}

	nameBoxH := nameValueStartY - nameBoxTop + float64(len(nameLines))*nameLineHeight + 4
	g.box(pdf, x, nameBoxTop, w, nameBoxH)

	// CNPJ | IE | CRT (8pt per §3.7.6)
	col1 := w * 0.35
	col2 := w * 0.35
	col3 := w - col1 - col2
	row1Y := nameBoxTop + nameBoxH
	g.cellEmit(pdf, x, row1Y, col1, row1H, "CNPJ", formatCNPJ(data.EmitterCNPJ))
	g.cellEmit(pdf, x+col1, row1Y, col2, row1H, "INSCRIÇÃO ESTADUAL", data.EmitterIE)
	g.cellEmit(pdf, x+col1+col2, row1Y, col3, row1H, "REGIME TRIBUTÁRIO", crtLabel(data.EmitterCRT))

	// Endereço completo (8pt per §3.7.6)
	addr := joinNonEmpty(
		data.EmitterAddress,
		data.EmitterNumber,
		data.EmitterComplement,
		data.EmitterNeighborhood,
		data.EmitterCEP,
		fmt.Sprintf("%s/%s", data.EmitterCity, data.EmitterUF),
		data.EmitterPhone,
	)
	g.cellEmit(pdf, x, row1Y+row1H, w, row2H, "ENDEREÇO", addr)

	return y + 10 + nameBoxH + row1H + row2H + 4
}

func (g *DANFEGenerator) drawDestinatario(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	const row1H = 24.0
	const row2H = 23.0

	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DESTINATÁRIO / REMETENTE")

	nameBoxTop := y + 10
	nameLabelY := nameBoxTop + 2
	nameValueStartY := nameBoxTop + 12
	nameLineHeight := 10.0

	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, nameLabelY)
	pdf.Cell(nil, "NOME / RAZÃO SOCIAL")

	pdf.SetFont("serif-bold", "", 12)
	nameLines := wrapText(pdf, data.DestName, w-8)
	nameY := nameValueStartY
	for _, line := range nameLines {
		pdf.SetXY(x+4, nameY)
		pdf.Cell(nil, line)
		nameY += nameLineHeight
	}

	nameBoxH := nameValueStartY - nameBoxTop + float64(len(nameLines))*nameLineHeight + 4
	g.box(pdf, x, nameBoxTop, w, nameBoxH)

	col1 := w * 0.35
	col2 := w * 0.35
	col3 := w - col1 - col2
	row1Y := nameBoxTop + nameBoxH
	g.cell(pdf, x, row1Y, col1, row1H, "CNPJ/CPF", formatCNPJ(data.DestCNPJ))
	g.cell(pdf, x+col1, row1Y, col2, row1H, "INSCRIÇÃO ESTADUAL", data.DestIE)
	g.cell(pdf, x+col1+col2, row1Y, col3, row1H, "INDICADOR IE DEST.", indIEDestLabel(data.DestIndIEDest))

	addr := joinNonEmpty(
		data.DestAddress,
		data.DestNumber,
		data.DestComplement,
		data.DestNeighborhood,
		data.DestCEP,
		fmt.Sprintf("%s/%s", data.DestCity, data.DestUF),
		data.DestPhone,
	)
	g.cell(pdf, x, row1Y+row1H, w, row2H, "ENDEREÇO", addr)

	return y + 10 + nameBoxH + row1H + row2H + 4
}

func (g *DANFEGenerator) drawTaxCalc(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	const rowH = 24.0
	h := 10 + rowH*2
	//g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
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
		{"V. PRODUTOS", fmtDec2(data.VProd)},
		{"V. FRETE", fmtDec2(data.VFrete)},
		{"V. SEGURO", fmtDec2(data.VSeg)},
		{"DESCONTO", fmtDec2(data.VDesc)},
		{"V. IPI", fmtDec2(data.VIPI)},
		{"OUT.DESP.", fmtDec2(data.VOutro)},
		{"V. NF", fmtDec2(data.TotalValue)},
		{"V.EST.TRI.", fmtDec2(data.VTotTrib)},
	}
	colW := w / 6.0
	for i := 0; i < 6; i++ {
		f := fields[i]
		fx := x + float64(i)*colW
		g.cell(pdf, fx, y+10, colW, rowH, f.label, f.value)
	}
	for i := 6; i < 12; i++ {
		f := fields[i]
		fx := x + float64(i-6)*colW
		g.cell(pdf, fx, y+10+rowH, colW, rowH, f.label, f.value)
	}
	return y + h + 4
}

func (g *DANFEGenerator) drawTransport(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	// Vehicle row (PLACA / UF / RNTRC) is only drawn when the NF-e carries
	// vehicle data, so DANFEs without it keep the classic layout height.
	row4H := 24.0
	hasVehicle := data.VeicPlate != "" || data.VeicUF != "" || data.VeicRNTC != ""
	h := 72.0
	if hasVehicle {
		h += row4H
	}
	//g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "TRANSPORTADOR / VOLUMES TRANSPORTADOS")

	row1H := 22.0
	row2H := 18.0
	row3H := 22.0

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

	// Row 4 (vehicle), only when the NF-e carries vehicle data (per
	// DANFE_RULES §4.10: placa X19, UF do veículo, RNTRC/ANTT).
	if hasVehicle {
		p1 := w * 0.25
		p2 := w * 0.15
		p3 := w - p1 - p2
		row4Y := y + 10 + row1H + row2H + row3H
		g.cell(pdf, x, row4Y, p1, row4H, "PLACA", data.VeicPlate)
		g.cell(pdf, x+p1, row4Y, p2, row4H, "UF", data.VeicUF)
		g.cell(pdf, x+p1+p2, row4Y, p3, row4H, "RNTRC/ANTT", data.VeicRNTC)
	}

	return y + h + 4
}

func (g *DANFEGenerator) drawProductTable(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	headerH := 22.0
	baseRowH := 20.0
	lineH := 8.0 // line height for 6pt font

	descColW := w * 0.22
	descColInner := descColW - 8 // padding

	// Pre-wrap description and infAdProd for each item to calculate heights.
	type itemLayout struct {
		rowH       float64
		descLines  []string
		infALines  []string
		hasVUnTrib bool
	}
	layout := make([]itemLayout, len(data.Products))
	totalRowsH := 0.0
	pdf.SetFont("serif", "", 6)
	for i, p := range data.Products {
		lh := baseRowH
		descLines := wrapText(pdf, p.Desc, descColInner)
		if len(descLines) > 1 {
			lh = float64(len(descLines))*lineH + 6
			if lh < baseRowH {
				lh = baseRowH
			}
		}

		hasVUnTrib := !p.UnitPrice.Equal(p.VUnTrib) && !p.VUnTrib.IsZero()
		if hasVUnTrib {
			lh += lineH
		}

		var infALines []string
		if p.InfAdProd != "" {
			infALines = wrapText(pdf, p.InfAdProd, w-8)
			lh += float64(len(infALines))*lineH + 2
		}

		layout[i] = itemLayout{rowH: lh, descLines: descLines, infALines: infALines, hasVUnTrib: hasVUnTrib}
		totalRowsH += lh
	}
	h := headerH + totalRowsH
	if h < headerH+baseRowH {
		h = headerH + baseRowH
	}

	//g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DADOS DO PRODUTO / SERVIÇO")

	// content box
	g.box(pdf, x, y+10, w, h-10)

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
	cyAcc := y + headerH
	for i, p := range data.Products {
		l := layout[i]
		lh := l.rowH
		cx := x

		// Render non-description columns (single line)
		singleVals := []string{
			p.Code,
			"", // description handled separately
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
		pdf.SetFont("serif", "", 6)
		for j, c := range cols {
			if j == 1 {
				// Description column: render wrapped lines
				descX := cx + 4
				descY := cyAcc + 5
				for _, line := range l.descLines {
					pdf.SetXY(descX, descY)
					pdf.Cell(nil, line)
					descY += lineH
				}
			} else {
				pdf.SetXY(cx+4, cyAcc+5)
				pdf.Cell(nil, singleVals[j])
			}
			cx += c.w
		}

		// Second line for vUnTrib when different from vUnCom (§3.1.7)
		if l.hasVUnTrib {
			cx := x
			for j, c := range cols {
				if j == 7 { // V. UNIT column
					pdf.SetFont("serif", "", 6)
					pdf.SetXY(cx+4, cyAcc+5+lineH*float64(max(len(l.descLines), 1)))
					pdf.Cell(nil, "Trib: "+fmtDec4(p.VUnTrib))
				}
				cx += c.w
			}
		}

		// infAdProd below the item (§3.1.7) — render wrapped lines
		if len(l.infALines) > 0 {
			infY := cyAcc + lh - float64(len(l.infALines))*lineH - 2
			pdf.SetFont("serif", "", 6)
			drawWrapped(pdf, l.infALines, x+4, infY, lineH)
		}

		// Item divider between items per §5.3
		if i < len(data.Products)-1 {
			pdf.SetLineWidth(0)
			pdf.Line(x+2, cyAcc+lh, x+w-2, cyAcc+lh)
			pdf.SetLineWidth(0)
		}

		cyAcc += lh
	}

	return y + h + 4
}

func (g *DANFEGenerator) drawTotalNota(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	const rowH = 24.0
	h := 10 + rowH*2
	//g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
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
		{"V. PRODUTOS", fmtDec2(data.VProd)},
		{"V. FRETE", fmtDec2(data.VFrete)},
		{"V. SEGURO", fmtDec2(data.VSeg)},
		{"DESCONTO", fmtDec2(data.VDesc)},
		{"V. IPI", fmtDec2(data.VIPI)},
		{"OUT.DESP.", fmtDec2(data.VOutro)},
		{"V. NF", fmtDec2(data.TotalValue)},
		{"V.EST.TRI.", fmtDec2(data.VTotTrib)},
	}
	colW := w / 6.0
	for i := 0; i < 6; i++ {
		f := fields[i]
		fx := x + float64(i)*colW
		g.cell(pdf, fx, y+10, colW, rowH, f.label, f.value)
	}
	for i := 6; i < 12; i++ {
		f := fields[i]
		fx := x + float64(i-6)*colW
		g.cell(pdf, fx, y+10+rowH, colW, rowH, f.label, f.value)
	}
	return y + h + 4
}

func (g *DANFEGenerator) drawISSQN(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	h := 34.0
	g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "CÁLCULO DO ISSQN")

	cols := w / 5.0
	g.cell(pdf, x, y+10, cols, h-10, "IM", "")
	g.cell(pdf, x+cols, y+10, cols, h-10, "VALOR TOTAL DOS SERVIÇOS", fmtDec2(data.VServ))
	g.cell(pdf, x+cols*2, y+10, cols, h-10, "BASE DE CÁLCULO DO ISSQN", fmtDec2(data.VBCISSQN))
	g.cell(pdf, x+cols*3, y+10, cols, h-10, "VALOR DO ISSQN", fmtDec2(data.VISSQN))
	g.cell(pdf, x+cols*4, y+10, w-cols*4, h-10, "VALOR TOTAL DA NOTA", fmtDec2(data.TotalValue))
	return y + h + 4
}

func (g *DANFEGenerator) drawAdditionalInfo(pdf *gopdf.GoPdf, data entity.DANFEData, x, y, w float64) float64 {
	// Dynamic height based on wrapped content (§3.1: fields shall represent
	// full XML content; §3.1.8: "Deverá conter todas as Informações Adicionais").
	leftW := w * 0.60
	rightW := w - leftW
	lineH := 8.0 // 6pt font line height
	minH := 56.0

	pdf.SetFont("serif", "", 6)
	infCplLines := wrapText(pdf, data.InfCpl, leftW-8)
	infFiscoLines := wrapText(pdf, data.InfAdFisco, rightW-8)

	contentH := 20.0 // header area
	if len(infCplLines) > 0 {
		contentH += float64(len(infCplLines))*lineH + 4
	} else {
		contentH += 10
	}
	// Use the taller of the two panels
	fiscoH := 20.0
	if len(infFiscoLines) > 0 {
		fiscoH += float64(len(infFiscoLines))*lineH + 4
	} else {
		fiscoH += 10
	}
	if fiscoH > contentH {
		contentH = fiscoH
	}
	h := contentH
	if h < minH {
		h = minH
	}

	g.box(pdf, x, y, w, h)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+4, y+3)
	pdf.Cell(nil, "DADOS ADICIONAIS")

	pdf.Line(x+leftW, y, x+leftW, y+h)

	// Informações Complementares (left panel)
	pdf.SetXY(x+4, y+10)
	pdf.Cell(nil, "INFORMAÇÕES COMPLEMENTARES")
	pdf.SetFont("serif", "", 6)
	if len(infCplLines) > 0 {
		drawWrapped(pdf, infCplLines, x+4, y+20, lineH)
	}

	// Reservado ao Fisco (right panel)
	pdf.SetFont("serif", "", 6)
	pdf.SetXY(x+leftW+4, y+10)
	pdf.Cell(nil, "RESERVADO AO FISCO")
	pdf.SetFont("serif", "", 6)
	if len(infFiscoLines) > 0 {
		drawWrapped(pdf, infFiscoLines, x+leftW+4, y+20, lineH)
	}

	return y + h + 4
}

// --- Formatting utilities ---

// wrapText splits text into lines that fit within maxWidth using the
// currently set font. Returns the slice of lines. Per MOC Anexo II §3.1:
// "O conteúdo dos campos poderá ser impresso em mais de uma linha desde
// que a leitura possa ser feita de forma clara." This replaces truncation
// to ensure fields "deverão representar o conteúdo das respectivas TAG XML."
func wrapText(pdf *gopdf.GoPdf, text string, maxWidth float64) []string {
	if text == "" {
		return nil
	}
	paragraphs := strings.Split(text, "\n")
	var lines []string
	for i, para := range paragraphs {
		words := strings.Fields(para)
		if len(words) == 0 {
			// Preserve intentional blank lines (e.g. double \n) unless leading or trailing.
			if i > 0 && i < len(paragraphs)-1 {
				lines = append(lines, "")
			}
			continue
		}
		current := words[0]
		for _, word := range words[1:] {
			candidate := current + " " + word
			tw, _ := pdf.MeasureTextWidth(candidate)
			if tw <= maxWidth {
				current = candidate
			} else {
				lines = append(lines, current)
				current = word
			}
		}
		lines = append(lines, current)
	}
	return lines
}

// drawWrapped renders wrapped text starting at (x, y) with the given line
// height. Returns the y position after the last line.
func drawWrapped(pdf *gopdf.GoPdf, lines []string, x, y, lineH float64) float64 {
	for _, line := range lines {
		pdf.SetXY(x, y)
		pdf.Cell(nil, line)
		y += lineH
	}
	return y
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
	if len(cnpj) == 11 {
		return formatCPF(cnpj)
	}
	return cnpj
}

func formatCPF(cpf string) string {
	if len(cpf) == 11 {
		return fmt.Sprintf("%s.%s.%s-%s", cpf[:3], cpf[3:6], cpf[6:9], cpf[9:])
	}
	return cpf
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
