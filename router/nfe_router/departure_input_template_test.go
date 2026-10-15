package nfe_router

import (
	"strings"
	"testing"

	"armazenda/pkg/nfe/defaults"
	"armazenda/service/nfe_service"

	"github.com/shopspring/decimal"
)

func departureModalTestData(regime defaults.TaxRegime) map[string]interface{} {
	return map[string]interface{}{
		"HasNFEFarmConfig":        true,
		"DepartureID":             24,
		"Product":                 "Milho",
		"NetWeight":               decimal.NewFromInt(25000),
		"DefaultICMSRate":         "17.00",
		"DefaultPISRate":          "1.00",
		"DefaultCOFINSRate":       "7.00",
		"DefaultIBSRate":          "0.10",
		"DefaultCBSRate":          "0.90",
		"DefaultNaturezaOp":       "Venda de produção do estabelecimento",
		"DefaultProductDesc":      "Milho",
		"DefaultNCM":              "10059010",
		"DefaultCEST":             "",
		"DefaultUnit":             "KG",
		"DefaultModFrete":         1,
		"DefaultTPag":             "90",
		"DefaultIndPag":           "0",
		"DefaultIncludeCobranca":  true,
		"DefaultUseVehicle":       true,
		"DefaultTransportPlate":   "ABC1234",
		"DefaultTransportPlateUF": "MT",
		"DefaultUseVolumes":       true,
		"DefaultVolQtd":           "1",
		"DefaultVolEsp":           "Granel",
		"DefaultVolPesoL":         "25000.000",
		"DefaultVolPesoB":         "50000.000",
		"DefaultICMSCST":          "51",
		"DefaultPISCST":           "08",
		"DefaultCOFINSCST":        "08",
		"DefaultCBSCST":           "000",
		"DefaultIBSCST":           "000",
		"DefaultCClassTrib":       "000001",
		"DefaultInfCpl":           "",
		"TaxRegime":               regime,
		"IsSimpleNational":        regime == defaults.TaxRegimeSimplesNacional,
		"ICMSCodeOptions":         nfe_service.DepartureICMSOptions(regime),
		"PISCOFINSOptions":        nfe_service.DeparturePISCOFINSOptions(),
		"CSPNonce":                "test-nonce",
	}
}

func TestDepartureNFeModalRendersInputsAndTaxSelects(t *testing.T) {
	parsed := parseAllTemplates(t)
	view := departureModalTestData(defaults.TaxRegimeLucroReal)
	out := executeBlock(t, parsed, "nfe-emit-modal", view)

	for _, want := range []string{
		`name="tpag"`,
		`name="indPag"`,
		`name="includeCobranca"`,
		`name="transportNome"`,
		`name="transportDocumento"`,
		`name="transportPlaca"`,
		`value="ABC1234"`,
		`name="includeVehicleSpecified"`,
		`name="includeVolumesSpecified"`,
		`name="volPesoL"`,
		`value="25000.000"`,
		`<select name="icmsCST" id="icmsCST"`,
		`<option value="51" selected>51 - Diferimento</option>`,
		`<option value="08" selected>08 - Sem incidência</option>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("departure modal must render %s", want)
		}
	}
	if strings.Contains(out, `<input type="text" name="icmsCST"`) {
		t.Error("ICMS CST must not remain a free-text input")
	}
}

func TestDepartureNFeModalUsesCSOSNForSimplesNacional(t *testing.T) {
	parsed := parseAllTemplates(t)
	view := departureModalTestData(defaults.TaxRegimeSimplesNacional)
	view["DefaultICMSCST"] = "102"
	out := executeBlock(t, parsed, "nfe-emit-modal", view)
	if !strings.Contains(out, `CSOSN ICMS`) || !strings.Contains(out, `<option value="102" selected>102 - Tributada sem permissão de crédito</option>`) {
		t.Errorf("Simples Nacional modal must render the configured CSOSN: %s", out)
	}
}

func TestDepartureNFePreviewCarriesInputsToConfirm(t *testing.T) {
	parsed := parseAllTemplates(t)
	view := map[string]interface{}{
		"DepartureID": 24, "Product": "Milho", "NetWeight": decimal.NewFromInt(100),
		"UnitPrice": decimal.NewFromInt(2), "TotalValue": decimal.NewFromInt(200), "PDFBase64": "",
		"CFOP": "5101", "ICMSRate": "17.00", "PISRate": "1.00", "COFINSRate": "7.00",
		"IBSRate": "0.10", "CBSRate": "0.90", "NaturezaOp": "Venda", "ProductDesc": "Milho",
		"NCM": "10059010", "CEST": "", "Unit": "KG", "ModFrete": "1", "ICMSCST": "51",
		"PISCST": "08", "COFINSCST": "08", "IBSCST": "000", "CBSCST": "000",
		"CClassTrib": "000001", "InfCpl": "", "TPag": "01", "IndPag": "1",
		"IncludeCobranca": false, "TransportNome": "Transportador", "TransportDocumento": "52998224725",
		"TransportIE": "", "TransportLogradouro": "Rodovia MT-100", "TransportNumero": "Km 10",
		"TransportMunicipio": "Cuiaba", "TransportUF": "MT", "IncludeVehicle": true,
		"TransportPlaca": "ABC1D23", "TransportPlacaUF": "MT", "TransportRNTC": "12345",
		"IncludeVolumes": true, "VolQtd": "2", "VolEsp": "Sacas", "VolMarca": "Teste",
		"VolPesoL": "95.500", "VolPesoB": "100.250", "CSPNonce": "test-nonce",
	}
	out := executeBlock(t, parsed, "nfe-preview", view)
	for _, want := range []string{
		`name="tpag" value="01"`,
		`name="indPag" value="1"`,
		`name="includeCobranca" value="0"`,
		`name="transportNome" value="Transportador"`,
		`name="includeVehicle" value="1"`,
		`name="transportPlaca" value="ABC1D23"`,
		`name="includeVolumes" value="1"`,
		`name="volPesoB" value="100.250"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("departure preview must preserve %s", want)
		}
	}
}
