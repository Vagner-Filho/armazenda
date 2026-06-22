package xml_test

import (
	"strings"
	"testing"

	"armazenda/pkg/nfe/xml"

	"github.com/shopspring/decimal"
)

const sampleSignedNFe = `
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00">
    <ide>
      <cUF>51</cUF>
      <cNF>12345678</cNF>
      <natOp>Venda de Mercadoria</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>123</nNF>
      <dhEmi>2025-03-15T10:30:00-03:00</dhEmi>
    </ide>
    <emit>
      <CNPJ>12345678000190</CNPJ>
      <xNome>Fazenda Exemplo LTDA</xNome>
      <enderEmit>
        <xLgr>Rodovia BR-163</xLgr>
        <nro>1000</nro>
        <xBairro>Zona Rural</xBairro>
        <cMun>5107875</cMun>
        <xMun>Sorriso</xMun>
        <UF>MT</UF>
      </enderEmit>
    </emit>
    <dest>
      <CNPJ>98765432000190</CNPJ>
      <xNome>Cliente Exemplo S/A</xNome>
      <enderDest>
        <xLgr>Av. Principal</xLgr>
        <nro>500</nro>
        <xBairro>Centro</xBairro>
        <cMun>5103402</cMun>
        <xMun>Cuiabá</xMun>
        <UF>MT</UF>
      </enderDest>
    </dest>
    <det nItem="1">
      <prod>
        <cProd>SOJA</cProd>
        <xProd>Soja em Graos</xProd>
        <NCM>12010010</NCM>
        <CFOP>5102</CFOP>
        <uCom>KG</uCom>
        <qCom>50000.0000</qCom>
        <vUnCom>150.0000</vUnCom>
        <vProd>7500000.00</vProd>
      </prod>
    </det>
    <total>
      <ICMSTot>
        <vBC>7500000.00</vBC>
        <vICMS>1275000.00</vICMS>
        <vProd>7500000.00</vProd>
        <vNF>7500000.00</vNF>
      </ICMSTot>
    </total>
  </infNFe>
</NFe>
`

const sampleAuthorizedNFe = `
<nfeProc xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00">
  <NFe>
    <infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00">
      <ide>
        <cUF>51</cUF>
        <cNF>12345678</cNF>
        <natOp>Venda de Mercadoria</natOp>
        <mod>55</mod>
        <serie>1</serie>
        <nNF>123</nNF>
        <dhEmi>2025-03-15T10:30:00-03:00</dhEmi>
      </ide>
      <emit>
        <CNPJ>12345678000190</CNPJ>
        <xNome>Fazenda Exemplo LTDA</xNome>
        <enderEmit>
          <xLgr>Rodovia BR-163</xLgr>
          <nro>1000</nro>
          <xBairro>Zona Rural</xBairro>
          <cMun>5107875</cMun>
          <xMun>Sorriso</xMun>
          <UF>MT</UF>
        </enderEmit>
      </emit>
      <dest>
        <CNPJ>98765432000190</CNPJ>
        <xNome>Cliente Exemplo S/A</xNome>
        <enderDest>
          <xLgr>Av. Principal</xLgr>
          <nro>500</nro>
          <xBairro>Centro</xBairro>
          <cMun>5103402</cMun>
          <xMun>Cuiabá</xMun>
          <UF>MT</UF>
        </enderDest>
      </dest>
      <det nItem="1">
        <prod>
          <cProd>SOJA</cProd>
          <xProd>Soja em Graos</xProd>
          <NCM>12010010</NCM>
          <CFOP>5102</CFOP>
          <uCom>KG</uCom>
          <qCom>50000.0000</qCom>
          <vUnCom>150.0000</vUnCom>
          <vProd>7500000.00</vProd>
        </prod>
      </det>
      <total>
        <ICMSTot>
          <vBC>7500000.00</vBC>
          <vICMS>1275000.00</vICMS>
          <vProd>7500000.00</vProd>
          <vNF>7500000.00</vNF>
        </ICMSTot>
      </total>
    </infNFe>
  </NFe>
  <protNFe versao="4.00">
    <infProt>
      <nProt>351250123456789</nProt>
      <dhRecbto>2025-03-15T10:31:00-03:00</dhRecbto>
    </infProt>
  </protNFe>
</nfeProc>
`

func TestParseDANFEData_Signed(t *testing.T) {
	data, err := xml.ParseDANFEData(sampleSignedNFe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.AccessKey != "51250312345678000190550010000001231234567890" {
		t.Errorf("AccessKey = %s, want %s", data.AccessKey, "51250312345678000190550010000001231234567890")
	}
	if data.Serie != 1 {
		t.Errorf("Serie = %d, want 1", data.Serie)
	}
	if data.Numero != 123 {
		t.Errorf("Numero = %d, want 123", data.Numero)
	}
	if data.NaturezaOp != "Venda de Mercadoria" {
		t.Errorf("NaturezaOp = %s, want 'Venda de Mercadoria'", data.NaturezaOp)
	}
	if !strings.Contains(data.EmissionDate, "15/03/2025") {
		t.Errorf("EmissionDate = %s, want to contain '15/03/2025'", data.EmissionDate)
	}

	if data.EmitterName != "Fazenda Exemplo LTDA" {
		t.Errorf("EmitterName = %s, want 'Fazenda Exemplo LTDA'", data.EmitterName)
	}
	if data.EmitterCNPJ != "12345678000190" {
		t.Errorf("EmitterCNPJ = %s, want '12345678000190'", data.EmitterCNPJ)
	}
	if data.EmitterCity != "Sorriso" {
		t.Errorf("EmitterCity = %s, want 'Sorriso'", data.EmitterCity)
	}
	if data.EmitterUF != "MT" {
		t.Errorf("EmitterUF = %s, want 'MT'", data.EmitterUF)
	}
	if !strings.Contains(data.EmitterAddress, "Rodovia BR-163") {
		t.Errorf("EmitterAddress = %s, want to contain 'Rodovia BR-163'", data.EmitterAddress)
	}

	if data.DestName != "Cliente Exemplo S/A" {
		t.Errorf("DestName = %s, want 'Cliente Exemplo S/A'", data.DestName)
	}
	if data.DestCNPJ != "98765432000190" {
		t.Errorf("DestCNPJ = %s, want '98765432000190'", data.DestCNPJ)
	}
	if data.DestCity != "Cuiabá" {
		t.Errorf("DestCity = %s, want 'Cuiabá'", data.DestCity)
	}
	if data.DestUF != "MT" {
		t.Errorf("DestUF = %s, want 'MT'", data.DestUF)
	}

	if len(data.Products) != 1 {
		t.Fatalf("Products len = %d, want 1", len(data.Products))
	}
	p := data.Products[0]
	if p.Code != "SOJA" {
		t.Errorf("Product.Code = %s, want 'SOJA'", p.Code)
	}
	if p.Desc != "Soja em Graos" {
		t.Errorf("Product.Desc = %s, want 'Soja em Graos'", p.Desc)
	}
	if p.NCM != "12010010" {
		t.Errorf("Product.NCM = %s, want '12010010'", p.NCM)
	}
	if p.CFOP != "5102" {
		t.Errorf("Product.CFOP = %s, want '5102'", p.CFOP)
	}
	if p.Unit != "KG" {
		t.Errorf("Product.Unit = %s, want 'KG'", p.Unit)
	}
	if !p.Quantity.Equal(decimal.NewFromFloat(50000.0)) {
		t.Errorf("Product.Quantity = %s, want 50000.0000", p.Quantity.String())
	}
	if !p.UnitPrice.Equal(decimal.NewFromFloat(150.0)) {
		t.Errorf("Product.UnitPrice = %s, want 150.0000", p.UnitPrice.String())
	}
	if !p.Total.Equal(decimal.NewFromFloat(7500000.0)) {
		t.Errorf("Product.Total = %s, want 7500000.00", p.Total.String())
	}

	if !data.TotalValue.Equal(decimal.NewFromFloat(7500000.0)) {
		t.Errorf("TotalValue = %s, want 7500000.00", data.TotalValue.String())
	}
	if !data.ICMSValue.Equal(decimal.NewFromFloat(1275000.0)) {
		t.Errorf("ICMSValue = %s, want 1275000.00", data.ICMSValue.String())
	}

	if data.Protocol != "" {
		t.Errorf("Protocol should be empty for signed NFe, got %s", data.Protocol)
	}
}

func TestParseDANFEData_Authorized(t *testing.T) {
	data, err := xml.ParseDANFEData(sampleAuthorizedNFe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Protocol != "351250123456789" {
		t.Errorf("Protocol = %s, want '351250123456789'", data.Protocol)
	}
	if !strings.Contains(data.ProtocolDate, "15/03/2025") {
		t.Errorf("ProtocolDate = %s, want to contain '15/03/2025'", data.ProtocolDate)
	}
}

func TestParseDANFEData_Empty(t *testing.T) {
	_, err := xml.ParseDANFEData("")
	if err == nil {
		t.Fatal("expected error for empty XML, got nil")
	}
}
