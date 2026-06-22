package service

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"armazenda/pkg/nfe/entity"

	"github.com/signintech/gopdf"
)

func getFontPath() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename is .../pkg/nfe/service/danfe.go
	// Go up 3 directories to reach project root
	dir := filepath.Dir(filename)
	for i := 0; i < 3; i++ {
		dir = filepath.Dir(dir)
	}
	return filepath.Join(dir, "assets", "fonts", "LiberationSans-Regular.ttf")
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

	const margin = 15.0
	const pageWidth = 595.0
	const rightMargin = pageWidth - margin

	fontPath := getFontPath()
	if err := pdf.AddTTFFont("liberation", fontPath); err != nil {
		return nil, fmt.Errorf("failed to add font: %w", err)
	}

	drawLine := func(x1, y1, x2, y2 float64) {
		pdf.SetLineWidth(0.5)
		pdf.Line(x1, y1, x2, y2)
	}

	offset := 0.0
	if isPreview {
		// Draw preview banner
		pdf.SetFillColor(255, 200, 200)
		pdf.Rectangle(margin, 5, rightMargin, 28, "FD", 0.5, 0)
		pdf.SetTextColor(180, 0, 0)
		pdf.SetFont("liberation", "", 12)
		pdf.SetXY(margin, 12)
		pdf.Cell(nil, "DOCUMENTO DE PRÉ-VISUALIZAÇÃO — SEM VALOR FISCAL")
		pdf.SetTextColor(0, 0, 0)
		offset = 22
	}

	// Title
	pdf.SetFont("liberation", "", 14)
	pdf.SetXY(margin, 15+offset)
	pdf.Cell(nil, "DANFE - Documento Auxiliar da Nota Fiscal Eletrônica")

	// Access key formatted in groups of 4
	pdf.SetFont("liberation", "", 10)
	pdf.SetXY(margin, 30+offset)
	pdf.Cell(nil, "Chave de Acesso: "+formatAccessKey(data.AccessKey))

	// Basic info
	pdf.SetXY(margin, 42+offset)
	pdf.Cell(nil, fmt.Sprintf("Série: %d    Número: %d    Natureza da Operação: %s", data.Serie, data.Numero, data.NaturezaOp))
	pdf.SetXY(margin, 54+offset)
	pdf.Cell(nil, fmt.Sprintf("Data de Emissão: %s", data.EmissionDate))

	// Line
	y := 66.0 + offset
	drawLine(margin, y, rightMargin, y)

	// Emitter
	pdf.SetFont("liberation", "", 9)
	pdf.SetXY(margin, y+4)
	pdf.Cell(nil, "EMITENTE")
	pdf.SetFont("liberation", "", 10)
	pdf.SetXY(margin, y+14)
	pdf.Cell(nil, data.EmitterName)
	pdf.SetXY(margin, y+26)
	pdf.Cell(nil, fmt.Sprintf("CNPJ/CPF: %s", data.EmitterCNPJ))
	pdf.SetXY(margin+250, y+26)
	pdf.Cell(nil, fmt.Sprintf("Município: %s / %s", data.EmitterCity, data.EmitterUF))
	pdf.SetXY(margin, y+38)
	pdf.Cell(nil, fmt.Sprintf("Endereço: %s", data.EmitterAddress))

	// Line
	y += 50
	drawLine(margin, y, rightMargin, y)

	// Recipient
	pdf.SetFont("liberation", "", 9)
	pdf.SetXY(margin, y+4)
	pdf.Cell(nil, "DESTINATÁRIO")
	pdf.SetFont("liberation", "", 10)
	pdf.SetXY(margin, y+14)
	pdf.Cell(nil, data.DestName)
	pdf.SetXY(margin, y+26)
	pdf.Cell(nil, fmt.Sprintf("CNPJ/CPF: %s", data.DestCNPJ))
	pdf.SetXY(margin+250, y+26)
	pdf.Cell(nil, fmt.Sprintf("Município: %s / %s", data.DestCity, data.DestUF))
	pdf.SetXY(margin, y+38)
	pdf.Cell(nil, fmt.Sprintf("Endereço: %s", data.DestAddress))

	// Line
	y += 50
	drawLine(margin, y, rightMargin, y)

	// Products
	pdf.SetFont("liberation", "", 9)
	pdf.SetXY(margin, y+4)
	pdf.Cell(nil, "PRODUTOS")

	y += 14
	drawProductHeader(&pdf, margin, y)
	y += 12
	for _, p := range data.Products {
		drawProductRow(&pdf, margin, y, p)
		y += 12
	}

	// Line
	y += 6
	drawLine(margin, y, rightMargin, y)

	// Totals
	pdf.SetFont("liberation", "", 10)
	pdf.SetXY(margin+350, y+10)
	pdf.Cell(nil, fmt.Sprintf("Total da Nota: R$ %s", data.TotalValue.StringFixed(2)))
	pdf.SetXY(margin+350, y+22)
	pdf.Cell(nil, fmt.Sprintf("ICMS: R$ %s", data.ICMSValue.StringFixed(2)))

	// Protocol
	if data.Protocol != "" {
		y += 40
		drawLine(margin, y, rightMargin, y)
		pdf.SetFont("liberation", "", 9)
		pdf.SetXY(margin, y+4)
		pdf.Cell(nil, "PROTOCOLO DE AUTORIZAÇÃO")
		pdf.SetFont("liberation", "", 10)
		pdf.SetXY(margin, y+14)
		pdf.Cell(nil, fmt.Sprintf("Protocolo: %s", data.Protocol))
		pdf.SetXY(margin, y+26)
		pdf.Cell(nil, fmt.Sprintf("Data de Autorização: %s", data.ProtocolDate))
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}
	return buf.Bytes(), nil
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

func drawProductHeader(pdf *gopdf.GoPdf, x, y float64) {
	pdf.SetFont("liberation", "", 9)
	pdf.SetXY(x, y)
	pdf.Cell(nil, "Código")
	pdf.SetXY(x+60, y)
	pdf.Cell(nil, "Descrição")
	pdf.SetXY(x+220, y)
	pdf.Cell(nil, "NCM")
	pdf.SetXY(x+280, y)
	pdf.Cell(nil, "CFOP")
	pdf.SetXY(x+330, y)
	pdf.Cell(nil, "Un")
	pdf.SetXY(x+370, y)
	pdf.Cell(nil, "Qtd")
	pdf.SetXY(x+420, y)
	pdf.Cell(nil, "V.Unit")
	pdf.SetXY(x+480, y)
	pdf.Cell(nil, "V.Total")
}

func drawProductRow(pdf *gopdf.GoPdf, x, y float64, p entity.DANFEProduct) {
	pdf.SetFont("liberation", "", 9)
	pdf.SetXY(x, y)
	pdf.Cell(nil, p.Code)
	pdf.SetXY(x+60, y)
	pdf.Cell(nil, truncate(p.Desc, 35))
	pdf.SetXY(x+220, y)
	pdf.Cell(nil, p.NCM)
	pdf.SetXY(x+280, y)
	pdf.Cell(nil, p.CFOP)
	pdf.SetXY(x+330, y)
	pdf.Cell(nil, p.Unit)
	pdf.SetXY(x+370, y)
	pdf.Cell(nil, p.Quantity.StringFixed(3))
	pdf.SetXY(x+420, y)
	pdf.Cell(nil, p.UnitPrice.StringFixed(4))
	pdf.SetXY(x+480, y)
	pdf.Cell(nil, p.Total.StringFixed(2))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
