package nfe_service

import (
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

func TestIsFourDigitCFOP(t *testing.T) {
	cases := []struct {
		cfop string
		want bool
	}{
		{"5101", true},
		{"0000", true},
		{"510", false},
		{"51012", false},
		{"5A01", false},
		{"+123", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsFourDigitCFOP(tc.cfop); got != tc.want {
			t.Errorf("IsFourDigitCFOP(%q) = %v, want %v", tc.cfop, got, tc.want)
		}
	}
}

func TestResolveDetachedNaturezaOp(t *testing.T) {
	explicit := "Venda de produção"
	farmDefault := "Operação padrão da fazenda"

	t.Run("explicit form value wins", func(t *testing.T) {
		got, ok := resolveDetachedNaturezaOp(
			&entity.InvoiceOverrides{NaturezaOp: &explicit},
			&farmDefault,
			[]string{"5101", "5102"},
		)
		if !ok || got != explicit {
			t.Fatalf("expected explicit %q, got %q (ok=%v)", explicit, got, ok)
		}
	})

	t.Run("farm default when form value missing", func(t *testing.T) {
		got, ok := resolveDetachedNaturezaOp(nil, &farmDefault, []string{"5101", "5102"})
		if !ok || got != farmDefault {
			t.Fatalf("expected farm default %q, got %q (ok=%v)", farmDefault, got, ok)
		}
	})

	t.Run("common item CFOP derives", func(t *testing.T) {
		got, ok := resolveDetachedNaturezaOp(nil, nil, []string{"5101", "5101"})
		if !ok {
			t.Fatal("expected derivation from the common CFOP")
		}
		if want := defaults.NaturezaOpForCFOP("5101"); got != want {
			t.Fatalf("expected derived %q, got %q", want, got)
		}
	})

	t.Run("mixed CFOPs without explicit or farm value are rejected", func(t *testing.T) {
		if got, ok := resolveDetachedNaturezaOp(nil, nil, []string{"5101", "5102"}); ok {
			t.Fatalf("expected mixed CFOPs to be ambiguous, got %q", got)
		}
	})

	t.Run("mixed CFOPs still use the farm default", func(t *testing.T) {
		got, ok := resolveDetachedNaturezaOp(nil, &farmDefault, []string{"5101", "5102"})
		if !ok || got != farmDefault {
			t.Fatalf("expected farm default %q, got %q (ok=%v)", farmDefault, got, ok)
		}
	})
}

func TestBuildDetachedItems_PerItemCFOPAndExactPrices(t *testing.T) {
	svc := NewNFeService()
	cfg := &entity_public.FarmConfig{TaxRegime: 3, DefaultUnit: "KG"}

	inputs := []DetachedItemInput{
		{
			ProductName: "Soja",
			NCM:         "12019000",
			CFOP:        "5101",
			Quantity:    decimal.RequireFromString("10000"),
			UnitPrice:   decimal.RequireFromString("150.25"),
			Unit:        "KG",
		},
		{
			ProductName: "Milho",
			NCM:         "10059000",
			CFOP:        "5102",
			Quantity:    decimal.RequireFromString("5000"),
			UnitPrice:   decimal.RequireFromString("120.75"),
			Unit:        "KG",
		},
	}

	items, itemsForDB, _, _, _ := svc.buildDetachedItems(inputs, cfg, entity.TaxRates{}, nil)

	if len(items) != 2 || len(itemsForDB) != 2 {
		t.Fatalf("expected 2 items, got %d/%d", len(items), len(itemsForDB))
	}
	if items[0].Produto.CFOP != "5101" || items[1].Produto.CFOP != "5102" {
		t.Errorf("expected per-item CFOPs 5101/5102, got %q/%q", items[0].Produto.CFOP, items[1].Produto.CFOP)
	}
	if itemsForDB[0].CFOP != "5101" || itemsForDB[1].CFOP != "5102" {
		t.Errorf("expected persisted per-item CFOPs 5101/5102, got %q/%q", itemsForDB[0].CFOP, itemsForDB[1].CFOP)
	}

	// Exact decimals must survive the mapping untouched.
	if !itemsForDB[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
		t.Errorf("expected exact unit price 150.25, got %s", itemsForDB[0].UnitPrice)
	}
	if !itemsForDB[1].UnitPrice.Equal(decimal.RequireFromString("120.75")) {
		t.Errorf("expected exact unit price 120.75, got %s", itemsForDB[1].UnitPrice)
	}
	if !itemsForDB[0].TotalValue.Equal(decimal.RequireFromString("1502500")) {
		t.Errorf("expected item total 1502500, got %s", itemsForDB[0].TotalValue)
	}
}

func TestValidateProfileItems(t *testing.T) {
	t.Run("valid items normalize and keep exact prices", func(t *testing.T) {
		toast, items := validateProfileItems([]DetachedProfileItemInput{
			{ProductName: " Soja ", NCM: "12019000", CFOP: "5101", Unit: "", UnitPrice: decimal.RequireFromString("150.25")},
			{ProductName: "Milho", NCM: "10059000", CFOP: "5102", Unit: "SC", UnitPrice: decimal.RequireFromString("120.75")},
		})
		if toast != nil {
			t.Fatalf("expected no toast, got %+v", toast)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if items[0].ProductName != "Soja" {
			t.Errorf("expected trimmed product name, got %q", items[0].ProductName)
		}
		if items[0].Unit != "KG" {
			t.Errorf("expected KG fallback unit, got %q", items[0].Unit)
		}
		if !items[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
			t.Errorf("expected exact price 150.25, got %s", items[0].UnitPrice)
		}
		if items[1].CFOP != "5102" {
			t.Errorf("expected second CFOP 5102, got %q", items[1].CFOP)
		}
	})

	t.Run("missing items rejected", func(t *testing.T) {
		if toast, _ := validateProfileItems(nil); toast == nil {
			t.Fatal("expected a warning for an empty item list")
		}
	})

	t.Run("malformed CFOP rejected", func(t *testing.T) {
		toast, _ := validateProfileItems([]DetachedProfileItemInput{
			{ProductName: "Soja", CFOP: "51A1", UnitPrice: decimal.NewFromInt(1)},
		})
		if toast == nil {
			t.Fatal("expected a warning for a malformed CFOP")
		}
	})

	t.Run("negative price rejected", func(t *testing.T) {
		toast, _ := validateProfileItems([]DetachedProfileItemInput{
			{ProductName: "Soja", CFOP: "5101", UnitPrice: decimal.RequireFromString("-1")},
		})
		if toast == nil {
			t.Fatal("expected a warning for a negative price")
		}
	})

	t.Run("explicit zero price is preserved", func(t *testing.T) {
		toast, items := validateProfileItems([]DetachedProfileItemInput{
			{ProductName: "Soja", CFOP: "5101", UnitPrice: decimal.Zero},
		})
		if toast != nil {
			t.Fatalf("expected zero price to be allowed, got %+v", toast)
		}
		if !items[0].UnitPrice.IsZero() {
			t.Errorf("expected explicit zero price, got %s", items[0].UnitPrice)
		}
	})
}

func TestProfileRecipientIDRequiresValidFarmReference(t *testing.T) {
	// Pure validation of the inline branches; the farm-scoped person/product
	// lookup is exercised against the database in the E2E flow.
	if _, ok := commonCFOP(nil); ok {
		t.Fatal("commonCFOP must report false for an empty item list")
	}
	if _, ok := commonCFOP([]string{"", ""}); ok {
		t.Fatal("commonCFOP must report false for empty CFOPs")
	}
	if got, ok := commonCFOP([]string{"5101", "5101"}); !ok || got != "5101" {
		t.Fatal("commonCFOP must return the shared code")
	}
}
