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
      <dhSaiEnt>2025-03-15T11:00:00-03:00</dhSaiEnt>
      <tpNF>1</tpNF>
      <tpEmis>1</tpEmis>
      <tpAmb>2</tpAmb>
      <verProc>Armazenda-1.0</verProc>
    </ide>
    <emit>
      <CNPJ>12345678000190</CNPJ>
      <xNome>Fazenda Exemplo LTDA</xNome>
      <IE>123456789</IE>
      <CRT>3</CRT>
      <enderEmit>
        <xLgr>Rodovia BR-163</xLgr>
        <nro>1000</nro>
        <xCpl>Km 500</xCpl>
        <xBairro>Zona Rural</xBairro>
        <cMun>5107875</cMun>
        <xMun>Sorriso</xMun>
        <UF>MT</UF>
        <CEP>78890000</CEP>
        <fone>6635441234</fone>
      </enderEmit>
    </emit>
    <dest>
      <CNPJ>98765432000190</CNPJ>
      <xNome>Cliente Exemplo S/A</xNome>
      <IE>987654321</IE>
      <indIEDest>1</indIEDest>
      <enderDest>
        <xLgr>Av. Principal</xLgr>
        <nro>500</nro>
        <xCpl>Sala 101</xCpl>
        <xBairro>Centro</xBairro>
        <cMun>5103402</cMun>
        <xMun>Cuiabá</xMun>
        <UF>MT</UF>
        <CEP>78000000</CEP>
        <fone>6533224455</fone>
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
        <vFrete>0.00</vFrete>
        <vSeg>0.00</vSeg>
        <vDesc>0.00</vDesc>
        <vOutro>0.00</vOutro>
        <uTrib>KG</uTrib>
        <qTrib>50000.0000</qTrib>
        <vUnTrib>150.0000</vUnTrib>
      </prod>
      <infAdProd>Produto conferido na balança rodoviária.</infAdProd>
      <imposto>
        <ICMS>
          <ICMS00>
            <CST>00</CST>
            <vBC>7500000.00</vBC>
            <pICMS>17.00</pICMS>
            <vICMS>1275000.00</vICMS>
          </ICMS00>
        </ICMS>
        <IPI>
          <IPITrib>
            <pIPI>0.00</pIPI>
            <vIPI>0.00</vIPI>
          </IPITrib>
        </IPI>
        <PIS>
          <PISAliq>
            <pPIS>0.65</pPIS>
            <vPIS>48750.00</vPIS>
          </PISAliq>
        </PIS>
        <COFINS>
          <COFINSAliq>
            <pCOFINS>3.00</pCOFINS>
            <vCOFINS>225000.00</vCOFINS>
          </COFINSAliq>
        </COFINS>
      </imposto>
    </det>
    <total>
      <ICMSTot>
        <vBC>7500000.00</vBC>
        <vICMS>1275000.00</vICMS>
        <vICMSDeson>0.00</vICMSDeson>
        <vBCST>0.00</vBCST>
        <vST>0.00</vST>
        <vII>0.00</vII>
        <vIPI>0.00</vIPI>
        <vPIS>48750.00</vPIS>
        <vCOFINS>225000.00</vCOFINS>
        <vFrete>0.00</vFrete>
        <vSeg>0.00</vSeg>
        <vDesc>0.00</vDesc>
        <vOutro>0.00</vOutro>
        <vNF>7500000.00</vNF>
        <vTotTrib>0.00</vTotTrib>
      </ICMSTot>
    </total>
    <transp>
      <modFrete>9</modFrete>
      <transporta>
        <CNPJ>11223344000155</CNPJ>
        <xNome>Transportadora Exemplo LTDA</xNome>
        <IE>112233445</IE>
        <xEnder>Rua dos Transportes, 100</xEnder>
        <xMun>Sorriso</xMun>
        <UF>MT</UF>
      </transporta>
      <vol>
        <qVol>1</qVol>
        <esp>Granel</esp>
        <pesoL>50000.000</pesoL>
        <pesoB>50020.000</pesoB>
      </vol>
    </transp>
    <infAdic>
      <infCpl>Entrega conforme contrato nº 12345. Peso aferido na balança.</infCpl>
    </infAdic>
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
        <IE>123456789</IE>
        <CRT>3</CRT>
        <enderEmit>
          <xLgr>Rodovia BR-163</xLgr>
          <nro>1000</nro>
          <xCpl>Km 500</xCpl>
          <xBairro>Zona Rural</xBairro>
          <cMun>5107875</cMun>
          <xMun>Sorriso</xMun>
          <UF>MT</UF>
          <CEP>78890000</CEP>
          <fone>6635441234</fone>
        </enderEmit>
      </emit>
      <dest>
        <CNPJ>98765432000190</CNPJ>
        <xNome>Cliente Exemplo S/A</xNome>
        <IE>987654321</IE>
        <indIEDest>1</indIEDest>
        <enderDest>
          <xLgr>Av. Principal</xLgr>
          <nro>500</nro>
          <xCpl>Sala 101</xCpl>
          <xBairro>Centro</xBairro>
          <cMun>5103402</cMun>
          <xMun>Cuiabá</xMun>
          <UF>MT</UF>
          <CEP>78000000</CEP>
          <fone>6533224455</fone>
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
      <cStat>100</cStat>
      <xMotivo>Autorizado o uso da NF-e</xMotivo>
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

	// Emission context (Grupo B)
	if data.TpEmis != "1" {
		t.Errorf("TpEmis = %s, want '1'", data.TpEmis)
	}
	if data.TpAmb != "2" {
		t.Errorf("TpAmb = %s, want '2' (homologação)", data.TpAmb)
	}
	if data.TpNF != "1" {
		t.Errorf("TpNF = %s, want '1' (saída)", data.TpNF)
	}
	if !strings.Contains(data.DhSaiEnt, "15/03/2025") {
		t.Errorf("DhSaiEnt = %s, want to contain '15/03/2025'", data.DhSaiEnt)
	}
	if data.DhCont != "" {
		t.Errorf("DhCont = %s, want empty for normal emission", data.DhCont)
	}
	if data.XJust != "" {
		t.Errorf("XJust = %s, want empty for normal emission", data.XJust)
	}
	if data.VerProc != "Armazenda-1.0" {
		t.Errorf("VerProc = %s, want 'Armazenda-1.0'", data.VerProc)
	}

	if data.EmitterName != "Fazenda Exemplo LTDA" {
		t.Errorf("EmitterName = %s, want 'Fazenda Exemplo LTDA'", data.EmitterName)
	}
	if data.EmitterCNPJ != "12345678000190" {
		t.Errorf("EmitterCNPJ = %s, want '12345678000190'", data.EmitterCNPJ)
	}
	if data.EmitterIE != "123456789" {
		t.Errorf("EmitterIE = %s, want '123456789'", data.EmitterIE)
	}
	if data.EmitterCRT != "3" {
		t.Errorf("EmitterCRT = %s, want '3'", data.EmitterCRT)
	}
	if data.EmitterAddress != "Rodovia BR-163" {
		t.Errorf("EmitterAddress = %s, want 'Rodovia BR-163'", data.EmitterAddress)
	}
	if data.EmitterNumber != "1000" {
		t.Errorf("EmitterNumber = %s, want '1000'", data.EmitterNumber)
	}
	if data.EmitterComplement != "Km 500" {
		t.Errorf("EmitterComplement = %s, want 'Km 500'", data.EmitterComplement)
	}
	if data.EmitterNeighborhood != "Zona Rural" {
		t.Errorf("EmitterNeighborhood = %s, want 'Zona Rural'", data.EmitterNeighborhood)
	}
	if data.EmitterCEP != "78890000" {
		t.Errorf("EmitterCEP = %s, want '78890000'", data.EmitterCEP)
	}
	if data.EmitterCity != "Sorriso" {
		t.Errorf("EmitterCity = %s, want 'Sorriso'", data.EmitterCity)
	}
	if data.EmitterUF != "MT" {
		t.Errorf("EmitterUF = %s, want 'MT'", data.EmitterUF)
	}
	if data.EmitterPhone != "6635441234" {
		t.Errorf("EmitterPhone = %s, want '6635441234'", data.EmitterPhone)
	}

	if data.DestName != "Cliente Exemplo S/A" {
		t.Errorf("DestName = %s, want 'Cliente Exemplo S/A'", data.DestName)
	}
	if data.DestCNPJ != "98765432000190" {
		t.Errorf("DestCNPJ = %s, want '98765432000190'", data.DestCNPJ)
	}
	if data.DestIE != "987654321" {
		t.Errorf("DestIE = %s, want '987654321'", data.DestIE)
	}
	if data.DestIndIEDest != "1" {
		t.Errorf("DestIndIEDest = %s, want '1'", data.DestIndIEDest)
	}
	if data.DestAddress != "Av. Principal" {
		t.Errorf("DestAddress = %s, want 'Av. Principal'", data.DestAddress)
	}
	if data.DestNumber != "500" {
		t.Errorf("DestNumber = %s, want '500'", data.DestNumber)
	}
	if data.DestComplement != "Sala 101" {
		t.Errorf("DestComplement = %s, want 'Sala 101'", data.DestComplement)
	}
	if data.DestNeighborhood != "Centro" {
		t.Errorf("DestNeighborhood = %s, want 'Centro'", data.DestNeighborhood)
	}
	if data.DestCEP != "78000000" {
		t.Errorf("DestCEP = %s, want '78000000'", data.DestCEP)
	}
	if data.DestCity != "Cuiabá" {
		t.Errorf("DestCity = %s, want 'Cuiabá'", data.DestCity)
	}
	if data.DestUF != "MT" {
		t.Errorf("DestUF = %s, want 'MT'", data.DestUF)
	}
	if data.DestPhone != "6533224455" {
		t.Errorf("DestPhone = %s, want '6533224455'", data.DestPhone)
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
	if p.CST != "00" {
		t.Errorf("Product.CST = %s, want '00'", p.CST)
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
	if !p.VBC.Equal(decimal.NewFromFloat(7500000.0)) {
		t.Errorf("Product.VBC = %s, want 7500000.00", p.VBC.String())
	}
	if !p.PICMS.Equal(decimal.NewFromFloat(17.0)) {
		t.Errorf("Product.PICMS = %s, want 17.00", p.PICMS.String())
	}
	if !p.VICMS.Equal(decimal.NewFromFloat(1275000.0)) {
		t.Errorf("Product.VICMS = %s, want 1275000.00", p.VICMS.String())
	}
	if !p.VPIS.Equal(decimal.NewFromFloat(48750.0)) {
		t.Errorf("Product.VPIS = %s, want 48750.00", p.VPIS.String())
	}
	if !p.VCOFINS.Equal(decimal.NewFromFloat(225000.0)) {
		t.Errorf("Product.VCOFINS = %s, want 225000.00", p.VCOFINS.String())
	}

	// Tributable unit (Grupo I)
	if p.UTrib != "KG" {
		t.Errorf("Product.UTrib = %s, want 'KG'", p.UTrib)
	}
	if !p.QTrib.Equal(decimal.NewFromFloat(50000.0)) {
		t.Errorf("Product.QTrib = %s, want 50000.0000", p.QTrib.String())
	}
	if !p.VUnTrib.Equal(decimal.NewFromFloat(150.0)) {
		t.Errorf("Product.VUnTrib = %s, want 150.0000", p.VUnTrib.String())
	}
	// infAdProd (sibling of prod inside det)
	if p.InfAdProd != "Produto conferido na balança rodoviária." {
		t.Errorf("Product.InfAdProd = %s, want 'Produto conferido na balança rodoviária.'", p.InfAdProd)
	}

	if !data.TotalValue.Equal(decimal.NewFromFloat(7500000.0)) {
		t.Errorf("TotalValue = %s, want 7500000.00", data.TotalValue.String())
	}
	if !data.VICMS.Equal(decimal.NewFromFloat(1275000.0)) {
		t.Errorf("VICMS = %s, want 1275000.00", data.VICMS.String())
	}
	if !data.VPIS.Equal(decimal.NewFromFloat(48750.0)) {
		t.Errorf("VPIS = %s, want 48750.00", data.VPIS.String())
	}
	if !data.VCOFINS.Equal(decimal.NewFromFloat(225000.0)) {
		t.Errorf("VCOFINS = %s, want 225000.00", data.VCOFINS.String())
	}

	if data.ModFrete != "9" {
		t.Errorf("ModFrete = %s, want '9'", data.ModFrete)
	}
	if data.TranspName != "Transportadora Exemplo LTDA" {
		t.Errorf("TranspName = %s, want 'Transportadora Exemplo LTDA'", data.TranspName)
	}
	if data.TranspCNPJ != "11223344000155" {
		t.Errorf("TranspCNPJ = %s, want '11223344000155'", data.TranspCNPJ)
	}
	if data.QVol != "1" {
		t.Errorf("QVol = %s, want '1'", data.QVol)
	}
	if data.Esp != "Granel" {
		t.Errorf("Esp = %s, want 'Granel'", data.Esp)
	}
	if !data.PesoL.Equal(decimal.NewFromFloat(50000.0)) {
		t.Errorf("PesoL = %s, want 50000.000", data.PesoL.String())
	}
	if !data.PesoB.Equal(decimal.NewFromFloat(50020.0)) {
		t.Errorf("PesoB = %s, want 50020.000", data.PesoB.String())
	}

	if data.InfCpl != "Entrega conforme contrato nº 12345. Peso aferido na balança." {
		t.Errorf("InfCpl = %s, want 'Entrega conforme contrato nº 12345. Peso aferido na balança.'", data.InfCpl)
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
	if data.CStat != "100" {
		t.Errorf("CStat = %s, want '100'", data.CStat)
	}
	if data.XMotivo != "Autorizado o uso da NF-e" {
		t.Errorf("XMotivo = %s, want 'Autorizado o uso da NF-e'", data.XMotivo)
	}
}

func TestParseDANFEData_Empty(t *testing.T) {
	_, err := xml.ParseDANFEData("")
	if err == nil {
		t.Fatal("expected error for empty XML, got nil")
	}
}

// TestParseDANFEData_ContingencyFields verifies that contingency-only fields
// (dhCont, xJust) and a non-default tpEmis are extracted when present.
func TestParseDANFEData_ContingencyFields(t *testing.T) {
	const svcXML = `
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe51250312345678000190550010000001231734567890" versao="4.00">
    <ide>
      <cUF>51</cUF>
      <cNF>73456789</cNF>
      <natOp>Venda de Mercadoria</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>123</nNF>
      <dhEmi>2025-03-15T10:30:00-03:00</dhEmi>
      <tpEmis>7</tpEmis>
      <dhCont>2025-03-15T10:25:00-03:00</dhCont>
      <xJust>SEFAZ originária indisponível</xJust>
      <tpAmb>1</tpAmb>
    </ide>
    <emit>
      <CNPJ>12345678000190</CNPJ>
      <xNome>Fazenda Exemplo LTDA</xNome>
      <IE>123456789</IE>
      <CRT>3</CRT>
    </emit>
    <total>
      <ICMSTot>
        <vNF>0.00</vNF>
      </ICMSTot>
    </total>
  </infNFe>
</NFe>
`
	data, err := xml.ParseDANFEData(svcXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.TpEmis != "7" {
		t.Errorf("TpEmis = %s, want '7' (SVC-RS)", data.TpEmis)
	}
	if data.TpAmb != "1" {
		t.Errorf("TpAmb = %s, want '1' (produção)", data.TpAmb)
	}
	if !strings.Contains(data.DhCont, "15/03/2025") {
		t.Errorf("DhCont = %s, want to contain '15/03/2025'", data.DhCont)
	}
	if data.XJust != "SEFAZ originária indisponível" {
		t.Errorf("XJust = %s, want contingency justification", data.XJust)
	}
}

// TestParseDANFEData_DifferingTributableUnit verifies that vUnTrib differing
// from vUnCom is captured separately (MOC Anexo II §3.1.7 requires both on
// the DANFE when they differ).
func TestParseDANFEData_DifferingTributableUnit(t *testing.T) {
	const xmlStr = `
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe51250312345678000190550010000001231234567890" versao="4.00">
    <ide>
      <cUF>51</cUF>
      <cNF>12345678</cNF>
      <natOp>Venda</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>123</nNF>
      <dhEmi>2025-03-15T10:30:00-03:00</dhEmi>
      <tpEmis>1</tpEmis>
      <tpAmb>2</tpAmb>
    </ide>
    <emit>
      <CNPJ>12345678000190</CNPJ>
      <xNome>Fazenda Exemplo LTDA</xNome>
      <IE>123456789</IE>
      <CRT>3</CRT>
    </emit>
    <det nItem="1">
      <prod>
        <cProd>SOJA</cProd>
        <xProd>Soja em Graos</xProd>
        <NCM>12010010</NCM>
        <CFOP>5102</CFOP>
        <uCom>SC</uCom>
        <qCom>500.0000</qCom>
        <vUnCom>15000.0000</vUnCom>
        <vProd>7500000.00</vProd>
        <uTrib>KG</uTrib>
        <qTrib>50000.0000</qTrib>
        <vUnTrib>150.0000</vUnTrib>
      </prod>
    </det>
    <total>
      <ICMSTot>
        <vNF>7500000.00</vNF>
      </ICMSTot>
    </total>
  </infNFe>
</NFe>
`
	data, err := xml.ParseDANFEData(xmlStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Products) != 1 {
		t.Fatalf("Products len = %d, want 1", len(data.Products))
	}
	p := data.Products[0]
	if p.UTrib != "KG" || p.Unit != "SC" {
		t.Errorf("Unit=%s UTrib=%s, want SC/KG (differing units)", p.Unit, p.UTrib)
	}
	if !p.UnitPrice.Equal(decimal.NewFromFloat(15000.0)) {
		t.Errorf("UnitPrice = %s, want 15000", p.UnitPrice.String())
	}
	if !p.VUnTrib.Equal(decimal.NewFromFloat(150.0)) {
		t.Errorf("VUnTrib = %s, want 150", p.VUnTrib.String())
	}
	if p.UnitPrice.Equal(p.VUnTrib) {
		t.Error("expected vUnCom != vUnTrib, got equal values")
	}
}
