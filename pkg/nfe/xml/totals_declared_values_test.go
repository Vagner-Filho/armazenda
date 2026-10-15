package xml_test

import (
	"testing"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/xml"

	"github.com/shopspring/decimal"
)

// TestBuilder_ICMSTot_OnlyDeclaredValues pins the totals invariant behind
// SEFAZ Rej. 531 ("Total da BC ICMS difere do somatório dos itens", G4
// homolog replay): ICMSTot must equal the sum of the values the items
// actually declare. For groups that declare nothing (ICMS40/41/50 and
// PISNT/COFINSNT, routed for CST 04–09) the entity values stay zero, so
// the totals must be zero while vProd/vNF remain intact.
func TestBuilder_ICMSTot_OnlyDeclaredValues(t *testing.T) {
	builder := xml.NewBuilder()

	input := minimalInvoiceInput()
	input.TotalValue = decimal.NewFromInt(600)
	input.Items = []entity.ItemData{
		{
			Numero: 1,
			Produto: entity.ProdutoData{
				Codigo:   "1",
				CEAN:     "SEM GTIN",
				XProd:    "Veneziana",
				NCM:      "84339090",
				CFOP:     "5913",
				UCom:     "UN",
				QCom:     decimal.NewFromInt(1),
				VUnCom:   decimal.NewFromInt(600),
				VProd:    decimal.NewFromInt(600),
				CEANTrib: "SEM GTIN",
				UTrib:    "UN",
				QTrib:    decimal.NewFromInt(1),
				VUnTrib:  decimal.NewFromInt(600),
				IndTot:   1,
			},
			Imposto: entity.ImpostoData{
				// CST 50 → ICMS40 declared with only orig+CST; entity values
				// (mirroring the mapping layer) stay empty.
				ICMS: entity.ICMSData{
					Origem: defaults.ICMSOrigemNacional,
					CST:    "50",
				},
				PIS:    entity.PISData{CST: "08"},
				COFINS: entity.COFINSData{CST: "08"},
			},
		},
	}

	doc, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	tot := findElement(t, doc.Root(), "./infNFe/total/ICMSTot")
	for _, want := range []struct{ tag, value string }{
		{"vBC", "0.00"},
		{"vICMS", "0.00"},
		{"vPIS", "0.00"},
		{"vCOFINS", "0.00"},
		{"vProd", "600.00"},
		{"vNF", "600.00"},
	} {
		if got := rwElText(tot, want.tag); got != want.value {
			t.Errorf("ICMSTot/%s = %q, want %q", want.tag, got, want.value)
		}
	}

	// The item's ICMS group itself declares nothing beyond orig+CST.
	icms := findElement(t, doc.Root(), "./infNFe/det/imposto/ICMS/ICMS40")
	if got := childTags(icms); len(got) != 2 {
		t.Errorf("ICMS40 children = %v, want only orig+CST", got)
	}
}
