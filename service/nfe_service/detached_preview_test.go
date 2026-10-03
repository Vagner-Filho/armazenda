package nfe_service

import (
	"fmt"
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// testDetachedItem builds an entity.ItemData with coherent per-item tax
// values so the DANFE-mapping tests can assert the summed totals.
func testDetachedItem(num int, total decimal.Decimal) entity.ItemData {
	icmsRate := decimal.RequireFromString("0.17")
	pisRate := decimal.RequireFromString("0.0165")
	cofinsRate := decimal.RequireFromString("0.076")

	return entity.ItemData{
		Numero: num,
		Produto: entity.ProdutoData{
			Codigo:  fmt.Sprintf("%d", num),
			XProd:   "Soja",
			NCM:     defaults.NCMSoja,
			CFOP:    "5101",
			UCom:    "KG",
			QCom:    total,
			VUnCom:  decimal.NewFromInt(1),
			VProd:   total,
			UTrib:   "KG",
			QTrib:   total,
			VUnTrib: decimal.NewFromInt(1),
			IndTot:  1,
		},
		Imposto: entity.ImpostoData{
			ICMS: entity.ICMSData{
				Origem: defaults.ICMSOrigemNacional,
				CST:    defaults.CSTTributadaIntegral,
				ModBC:  "3",
				VBC:    total,
				PICMS:  decimal.NewFromInt(17),
				VICMS:  total.Mul(icmsRate),
			},
			PIS: entity.PISData{
				CST:  "01",
				VBC:  total,
				PPIS: decimal.RequireFromString("1.65"),
				VPIS: total.Mul(pisRate),
			},
			COFINS: entity.COFINSData{
				CST:     "01",
				VBC:     total,
				PCOFINS: decimal.RequireFromString("7.60"),
				VCOFINS: total.Mul(cofinsRate),
			},
			IBSCBS: entity.IBSCBSData{
				CST:        defaults.IBSCBSCSTTributadaIntegral,
				CClassTrib: defaults.CClassTribDefault,
				VBC:        total,
				VIBSUF:     total.Mul(defaults.IBSRate2026),
				VIBSMun:    decimal.Zero,
				VIBS:       total.Mul(defaults.IBSRate2026),
				PCBS:       decimal.RequireFromString("0.90"),
				VCBS:       total.Mul(defaults.CBSRate2026),
			},
		},
	}
}

func TestBuildDetachedDANFEData_MultiItemTotals(t *testing.T) {
	svc := NewNFeService()

	itemA := testDetachedItem(1, decimal.RequireFromString("1000"))
	itemB := testDetachedItem(2, decimal.RequireFromString("500"))
	itemC := testDetachedItem(3, decimal.RequireFromString("250"))
	total := decimal.RequireFromString("1750")

	input := entity.InvoiceInput{
		Serie:                 1,
		Numero:                0,
		Environment:           1,
		TpEmis:                defaults.EmissaoNormal,
		NaturezaOp:            "VENDA",
		Emitter:               entity.EmitterData{Document: "12345678000195", UF: "MT", XNome: "Fazenda"},
		Recipient:             entity.RecipientData{XNome: "Cliente", CNPJ: "98765432000195"},
		Items:                 []entity.ItemData{itemA, itemB, itemC},
		Transport:             entity.TransportData{ModFrete: 9},
		TotalValue:            total,
		InformacoesAdicionais: "teste",
	}

	data := svc.buildDetachedDANFEData(input)

	if data.AccessKey != "" {
		t.Errorf("preview must not carry an access key, got %q", data.AccessKey)
	}
	if data.Numero != 0 {
		t.Errorf("preview must not carry a number, got %d", data.Numero)
	}
	if data.Serie != 1 {
		t.Errorf("expected serie 1, got %d", data.Serie)
	}
	if data.TpEmis != "1" {
		t.Errorf("expected TpEmis '1', got %q", data.TpEmis)
	}
	if data.TpAmb != "1" {
		t.Errorf("expected TpAmb '1', got %q", data.TpAmb)
	}
	if len(data.Products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(data.Products))
	}

	// Per-row totals must stay with their own item.
	if !data.Products[1].Total.Equal(decimal.RequireFromString("500")) {
		t.Errorf("expected product 2 total 500, got %s", data.Products[1].Total)
	}
	if !data.Products[2].Quantity.Equal(decimal.RequireFromString("250")) {
		t.Errorf("expected product 3 quantity 250, got %s", data.Products[2].Quantity)
	}

	// Invoice-level totals must be summed across every item.
	if !data.TotalValue.Equal(total) {
		t.Errorf("expected TotalValue %s, got %s", total, data.TotalValue)
	}
	if !data.VBC.Equal(total) {
		t.Errorf("expected VBC %s, got %s", total, data.VBC)
	}

	expectedVICMS := itemA.Imposto.ICMS.VICMS.Add(itemB.Imposto.ICMS.VICMS).Add(itemC.Imposto.ICMS.VICMS)
	if !data.VICMS.Equal(expectedVICMS) {
		t.Errorf("expected VICMS %s, got %s", expectedVICMS, data.VICMS)
	}

	expectedVPIS := itemA.Imposto.PIS.VPIS.Add(itemB.Imposto.PIS.VPIS).Add(itemC.Imposto.PIS.VPIS)
	if !data.VPIS.Equal(expectedVPIS) {
		t.Errorf("expected VPIS %s, got %s", expectedVPIS, data.VPIS)
	}

	expectedVCOFINS := itemA.Imposto.COFINS.VCOFINS.Add(itemB.Imposto.COFINS.VCOFINS).Add(itemC.Imposto.COFINS.VCOFINS)
	if !data.VCOFINS.Equal(expectedVCOFINS) {
		t.Errorf("expected VCOFINS %s, got %s", expectedVCOFINS, data.VCOFINS)
	}

	expectedVIBS := itemA.Imposto.IBSCBS.VIBS.Add(itemB.Imposto.IBSCBS.VIBS).Add(itemC.Imposto.IBSCBS.VIBS)
	if !data.VIBS.Equal(expectedVIBS) {
		t.Errorf("expected VIBS %s, got %s", expectedVIBS, data.VIBS)
	}
	if !data.VBCIBSCBS.Equal(total) {
		t.Errorf("expected VBCIBSCBS %s, got %s", total, data.VBCIBSCBS)
	}

	expectedVCBS := itemA.Imposto.IBSCBS.VCBS.Add(itemB.Imposto.IBSCBS.VCBS).Add(itemC.Imposto.IBSCBS.VCBS)
	if !data.VCBS.Equal(expectedVCBS) {
		t.Errorf("expected VCBS %s, got %s", expectedVCBS, data.VCBS)
	}

	// PIBS is derived from VIBSUF / VBC: the 2026 symbolic 0.1 % rate.
	if !data.Products[0].PIBS.Equal(decimal.RequireFromString("0.1")) {
		t.Errorf("expected derived PIBS 0.1, got %s", data.Products[0].PIBS)
	}

	if data.ModFrete != "9" {
		t.Errorf("expected ModFrete '9', got %q", data.ModFrete)
	}
	if data.InfCpl != "teste" {
		t.Errorf("expected InfCpl 'teste', got %q", data.InfCpl)
	}
}

func TestBuildDetachedDANFEData_SimplesNacionalUsesCSOSN(t *testing.T) {
	svc := NewNFeService()

	item := testDetachedItem(1, decimal.RequireFromString("100"))
	item.Imposto.ICMS.CST = ""
	item.Imposto.ICMS.CSOSN = defaults.CSOSNSemPermissaoCredito

	data := svc.buildDetachedDANFEData(entity.InvoiceInput{Items: []entity.ItemData{item}})

	if len(data.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(data.Products))
	}
	if data.Products[0].CST != defaults.CSOSNSemPermissaoCredito {
		t.Errorf("expected CSOSN fallback %q, got %q", defaults.CSOSNSemPermissaoCredito, data.Products[0].CST)
	}
}

func TestBuildDetachedItems_DefaultsAndTaxTotals(t *testing.T) {
	svc := NewNFeService()

	cfg := &entity_public.FarmConfig{
		TaxRegime:       3, // normal (non-Simples) regime
		DefaultUnit:     "",
		DefaultModFrete: 9,
		ICMSRate:        decimal.RequireFromString("0.12"),
		PISRate:         decimal.RequireFromString("0.0165"),
		COFINSRate:      decimal.RequireFromString("0.076"),
		IBSRate:         decimal.RequireFromString("0.001"),
		CBSRate:         decimal.RequireFromString("0.009"),
	}

	inputs := []DetachedItemInput{
		{
			ProductName: "",
			NCM:         "",
			Quantity:    decimal.RequireFromString("10"),
			GrossWeight: decimal.RequireFromString("11"),
			UnitPrice:   decimal.RequireFromString("100"),
			Unit:        "",
		},
		{
			ProductName: "Milho",
			NCM:         "10059000",
			Quantity:    decimal.RequireFromString("2"),
			GrossWeight: decimal.RequireFromString("2.2"),
			UnitPrice:   decimal.RequireFromString("120"),
			Unit:        "KG",
		},
	}

	items, itemsForDB, totalValue, totalIBS, totalCBS := svc.buildDetachedItems(inputs, "5101", cfg, entity.TaxRates{}, nil)

	if len(items) != 2 || len(itemsForDB) != 2 {
		t.Fatalf("expected 2 items, got %d/%d", len(items), len(itemsForDB))
	}

	// Fallbacks for missing product data.
	if items[0].Produto.XProd != "Produto Agricola" {
		t.Errorf("expected fallback description, got %q", items[0].Produto.XProd)
	}
	if items[0].Produto.NCM != defaults.NCMSoja {
		t.Errorf("expected fallback NCM %q, got %q", defaults.NCMSoja, items[0].Produto.NCM)
	}
	if items[0].Produto.UCom != "KG" {
		t.Errorf("expected fallback unit KG, got %q", items[0].Produto.UCom)
	}

	// Per-item totals and summed invoice totals.
	if !itemsForDB[0].TotalValue.Equal(decimal.RequireFromString("1000")) {
		t.Errorf("expected item 1 total 1000, got %s", itemsForDB[0].TotalValue)
	}
	if !itemsForDB[1].TotalValue.Equal(decimal.RequireFromString("240")) {
		t.Errorf("expected item 2 total 240, got %s", itemsForDB[1].TotalValue)
	}
	expectedTotal := decimal.RequireFromString("1240")
	if !totalValue.Equal(expectedTotal) {
		t.Errorf("expected total value %s, got %s", expectedTotal, totalValue)
	}

	// Rates come from the farm config when the user provides none.
	if !items[0].Imposto.ICMS.PICMS.Equal(decimal.RequireFromString("12")) {
		t.Errorf("expected PICMS 12, got %s", items[0].Imposto.ICMS.PICMS)
	}
	if !items[0].Imposto.ICMS.VICMS.Equal(decimal.RequireFromString("120")) {
		t.Errorf("expected VICMS 120, got %s", items[0].Imposto.ICMS.VICMS)
	}

	// Tax reform totals: per-item VIBS/VCBS are summed into the totals.
	expectedIBS := items[0].Imposto.IBSCBS.VIBS.Add(items[1].Imposto.IBSCBS.VIBS)
	if !totalIBS.Equal(expectedIBS) {
		t.Errorf("expected total IBS %s, got %s", expectedIBS, totalIBS)
	}
	expectedCBS := items[0].Imposto.IBSCBS.VCBS.Add(items[1].Imposto.IBSCBS.VCBS)
	if !totalCBS.Equal(expectedCBS) {
		t.Errorf("expected total CBS %s, got %s", expectedCBS, totalCBS)
	}
	if !items[0].Imposto.IBSCBS.VBC.Equal(decimal.RequireFromString("1000")) {
		t.Errorf("expected per-item IBS VBC 1000, got %s", items[0].Imposto.IBSCBS.VBC)
	}
}

func TestBuildDetachedItems_UserRatesAndOverrides(t *testing.T) {
	svc := NewNFeService()

	cfg := &entity_public.FarmConfig{
		TaxRegime: 3,
		ICMSRate:  decimal.RequireFromString("0.12"),
		CBSRate:   decimal.RequireFromString("0.009"),
		IBSRate:   decimal.RequireFromString("0.001"),
	}

	icmsRate := decimal.RequireFromString("0.05")
	icmsCST := "20"
	inputs := []DetachedItemInput{
		{
			ProductName: "Soja",
			NCM:         defaults.NCMSoja,
			Quantity:    decimal.RequireFromString("10"),
			UnitPrice:   decimal.RequireFromString("100"),
			Unit:        "KG",
		},
	}

	items, _, _, _, _ := svc.buildDetachedItems(
		inputs, "5101", cfg,
		entity.TaxRates{ICMSRate: &icmsRate},
		&entity.InvoiceOverrides{ICMSCST: &icmsCST},
	)

	if items[0].Imposto.ICMS.CST != "20" {
		t.Errorf("expected ICMS CST override '20', got %q", items[0].Imposto.ICMS.CST)
	}
	if !items[0].Imposto.ICMS.PICMS.Equal(decimal.RequireFromString("5")) {
		t.Errorf("expected PICMS 5 from user rate, got %s", items[0].Imposto.ICMS.PICMS)
	}
	if !items[0].Imposto.ICMS.VICMS.Equal(decimal.RequireFromString("50")) {
		t.Errorf("expected VICMS 50, got %s", items[0].Imposto.ICMS.VICMS)
	}
}

func TestDetachedDocumentType(t *testing.T) {
	cases := []struct {
		document string
		want     int
	}{
		{"12345678901", 1},    // CPF (11 digits) → natural
		{"", 1},               // missing document stays natural
		{"12345678000195", 2}, // CNPJ (14 digits) → legal
	}

	for _, tc := range cases {
		if got := detachedDocumentType(tc.document); got != tc.want {
			t.Errorf("detachedDocumentType(%q) = %d, want %d", tc.document, got, tc.want)
		}
	}
}
