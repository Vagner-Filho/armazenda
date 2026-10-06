package nfe_router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"
	"armazenda/service/nfe_service"
	cfop_view "armazenda/view/cfop"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// newFormContext builds a gin test context carrying an application/x-www-form-urlencoded body.
func newFormContext(t *testing.T, values url.Values) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req
	return c
}

func TestParseDetachedItems_PerItemCFOP(t *testing.T) {
	values := url.Values{
		"itemCount":              {"2"},
		"items[0].cfop":          {"5101"},
		"items[0].productName":   {"Soja"},
		"items[0].quantity":      {"10"},
		"items[0].unitPrice":     {"150.25"},
		"items[0].unit":          {"KG"},
		"items[1].cfop":          {"5102"},
		"items[1].productName":   {"Milho"},
		"items[1].quantity":      {"5"},
		"items[1].unitPrice":     {"120.75"},
		"items[1].unit":          {"KG"},
		"items[1].farmProductId": {"7"},
	}

	items, toast := parseDetachedItems(newFormContext(t, values))
	if toast.Type != 0 {
		t.Fatalf("expected no toast, got %+v", toast)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].CFOP != "5101" || items[1].CFOP != "5102" {
		t.Errorf("expected distinct CFOPs 5101/5102, got %q/%q", items[0].CFOP, items[1].CFOP)
	}
	if items[1].FarmProductID == nil || *items[1].FarmProductID != 7 {
		t.Errorf("expected farm product 7, got %v", items[1].FarmProductID)
	}
	if !items[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
		t.Errorf("expected exact price 150.25, got %s", items[0].UnitPrice)
	}
}

func TestParseDetachedItems_RejectsMalformedCFOP(t *testing.T) {
	values := url.Values{
		"itemCount":            {"1"},
		"items[0].cfop":        {"51A1"},
		"items[0].productName": {"Soja"},
		"items[0].quantity":    {"10"},
		"items[0].unitPrice":   {"150"},
		"items[0].unit":        {"KG"},
	}

	if _, toast := parseDetachedItems(newFormContext(t, values)); toast.Type == 0 {
		t.Fatal("expected a warning for a malformed per-item CFOP")
	}
}

func TestParseDetachedItems_RejectsMissingCFOP(t *testing.T) {
	values := url.Values{
		"itemCount":            {"1"},
		"items[0].productName": {"Soja"},
		"items[0].quantity":    {"10"},
		"items[0].unitPrice":   {"150"},
		"items[0].unit":        {"KG"},
	}

	if _, toast := parseDetachedItems(newFormContext(t, values)); toast.Type == 0 {
		t.Fatal("expected a warning for a missing per-item CFOP")
	}
}

func TestParseDetachedProfileItems_DoesNotRequireQuantity(t *testing.T) {
	values := url.Values{
		"itemCount":            {"2"},
		"items[0].cfop":        {"5101"},
		"items[0].productName": {"Soja"},
		"items[0].unitPrice":   {"150.25"},
		"items[0].unit":        {"KG"},
		"items[1].cfop":        {"5102"},
		"items[1].productName": {"Milho"},
		"items[1].unitPrice":   {"120.75"},
		"items[1].cest":        {"0100101"},
	}

	items, toast := parseDetachedProfileItems(newFormContext(t, values))
	if toast.Type != 0 {
		t.Fatalf("expected no toast, got %+v", toast)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].CFOP != "5101" || items[1].CFOP != "5102" {
		t.Errorf("expected distinct CFOPs, got %q/%q", items[0].CFOP, items[1].CFOP)
	}
	if !items[0].UnitPrice.Equal(decimal.RequireFromString("150.25")) {
		t.Errorf("expected exact price 150.25, got %s", items[0].UnitPrice)
	}
	if !items[1].UnitPrice.Equal(decimal.RequireFromString("120.75")) {
		t.Errorf("expected exact price 120.75, got %s", items[1].UnitPrice)
	}
	if items[1].CEST == nil || *items[1].CEST != "0100101" {
		t.Errorf("expected CEST to be parsed, got %v", items[1].CEST)
	}
}

func TestParseDetachedProfileItems_RejectsInvalidPrice(t *testing.T) {
	values := url.Values{
		"itemCount":            {"1"},
		"items[0].cfop":        {"5101"},
		"items[0].productName": {"Soja"},
		"items[0].unitPrice":   {"-10"},
	}

	if _, toast := parseDetachedProfileItems(newFormContext(t, values)); toast.Type == 0 {
		t.Fatal("expected a warning for a negative unit price")
	}
}

func TestParseDetachedProfileRecipient(t *testing.T) {
	t.Run("existing person", func(t *testing.T) {
		recipient, toast := parseDetachedProfileRecipient(newFormContext(t, url.Values{
			"recipientPersonId": {"42"},
		}))
		if toast.Type != 0 {
			t.Fatalf("expected no toast, got %+v", toast)
		}
		if recipient.PersonID == nil || *recipient.PersonID != 42 {
			t.Fatalf("expected person 42, got %v", recipient.PersonID)
		}
	})

	t.Run("no recipient is allowed", func(t *testing.T) {
		recipient, toast := parseDetachedProfileRecipient(newFormContext(t, url.Values{}))
		if toast.Type != 0 {
			t.Fatalf("expected no toast, got %+v", toast)
		}
		if recipient.PersonID != nil || recipient.Name != "" {
			t.Fatalf("expected an empty recipient, got %+v", recipient)
		}
	})

	t.Run("inline recipient", func(t *testing.T) {
		recipient, toast := parseDetachedProfileRecipient(newFormContext(t, url.Values{
			"recipientName":     {"Cliente Teste"},
			"recipientDocument": {"12345678000195"},
			"recipientCity":     {"Cuiabá"},
			"recipientState":    {"MT"},
		}))
		if toast.Type != 0 {
			t.Fatalf("expected no toast, got %+v", toast)
		}
		if recipient.Name != "Cliente Teste" || recipient.Document != "12345678000195" {
			t.Fatalf("expected inline recipient to parse, got %+v", recipient)
		}
		if recipient.City == nil || *recipient.City != "Cuiabá" {
			t.Fatalf("expected city to parse, got %v", recipient.City)
		}
	})

	t.Run("incomplete inline recipient rejected", func(t *testing.T) {
		if _, toast := parseDetachedProfileRecipient(newFormContext(t, url.Values{
			"recipientName": {"Cliente Teste"},
		})); toast.Type == 0 {
			t.Fatal("expected a warning for an incomplete inline recipient")
		}
	})
}

func TestDetachedProfileFormView_Rates(t *testing.T) {
	zero := decimal.Zero
	icms := decimal.RequireFromString("0.17")

	t.Run("explicit zero preserved and unset stays blank", func(t *testing.T) {
		view := detachedProfileFormView(&nfe_model.DetachedProfile{
			ID:   1,
			Name: "Rascunho",
			TaxRates: &entity.TaxRates{
				ICMSRate: &icms,
				PISRate:  &zero,
			},
		}, nil, nil, cfop_view.CfopOptions{Default: "5101"}, "nonce")

		if view["ICMSRate"].(string) != "17.00" {
			t.Errorf("expected ICMS rate 17.00, got %q", view["ICMSRate"])
		}
		if view["PISRate"].(string) != "0.00" {
			t.Errorf("expected explicit zero PIS displayed as 0.00, got %q", view["PISRate"])
		}
		if view["COFINSRate"].(string) != "" {
			t.Errorf("expected unset COFINS to stay blank, got %q", view["COFINSRate"])
		}
		if view["UseDefaultTaxRates"].(bool) {
			t.Error("expected use-default to be false when any rate is explicit")
		}
	})

	t.Run("no explicit rates uses farm defaults", func(t *testing.T) {
		view := detachedProfileFormView(&nfe_model.DetachedProfile{ID: 2, Name: "Rascunho"}, nil, nil, cfop_view.CfopOptions{Default: "5101"}, "nonce")
		if !view["UseDefaultTaxRates"].(bool) {
			t.Error("expected use-default to be true when no rate is explicit")
		}
	})

	t.Run("new rascunho starts with one empty row", func(t *testing.T) {
		view := detachedProfileFormView(nil, nil, nil, cfop_view.CfopOptions{Default: "5101"}, "nonce")
		rows := view["Rows"].([]detachedProfileFormRow)
		if len(rows) != 1 || rows[0].CFOP != "" || rows[0].UnitPrice != "" {
			t.Fatalf("expected one blank row, got %+v", rows)
		}
	})

	t.Run("rows expose profile item values", func(t *testing.T) {
		farmProductID := uint16(3)
		view := detachedProfileFormView(&nfe_model.DetachedProfile{
			Items: []nfe_model.DetachedProfileItem{
				{
					FarmProductID: &farmProductID,
					ProductName:   "Soja",
					NCM:           "12019000",
					CFOP:          "5101",
					Unit:          "KG",
					UnitPrice:     decimal.RequireFromString("150.25"),
				},
			},
		}, nil, nil, cfop_view.CfopOptions{Default: "5101"}, "nonce")

		rows := view["Rows"].([]detachedProfileFormRow)
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if rows[0].FarmProductID != "3" || rows[0].UnitPrice != "150.25" || rows[0].CFOP != "5101" {
			t.Fatalf("unexpected row %+v", rows[0])
		}
	})
}

func TestDetachedPreviewItems_RoundTrip(t *testing.T) {
	cest := "0100101"
	farmProductID := uint16(9)
	parsed := []nfe_service.DetachedItemInput{
		{
			FarmProductID: &farmProductID,
			ProductName:   "Soja",
			NCM:           "12019000",
			CEST:          &cest,
			CFOP:          "5101",
			Quantity:      decimal.RequireFromString("10000"),
			GrossWeight:   decimal.RequireFromString("10200"),
			UnitPrice:     decimal.RequireFromString("150.25"),
			Unit:          "KG",
		},
		{
			ProductName: "Milho",
			NCM:         "10059000",
			CFOP:        "5102",
			Quantity:    decimal.RequireFromString("5000"),
			GrossWeight: decimal.RequireFromString("5100"),
			UnitPrice:   decimal.RequireFromString("120.75"),
			Unit:        "KG",
		},
	}

	items := detachedPreviewItems(parsed)
	if len(items) != 2 {
		t.Fatalf("expected 2 preview items, got %d", len(items))
	}
	if items[0].CFOP != "5101" || items[1].CFOP != "5102" {
		t.Errorf("expected per-item CFOPs to round-trip, got %q/%q", items[0].CFOP, items[1].CFOP)
	}
	if items[0].UnitPrice != "150.25" || items[1].UnitPrice != "120.75" {
		t.Errorf("expected exact price strings, got %q/%q", items[0].UnitPrice, items[1].UnitPrice)
	}
	if items[0].Quantity != "10000" || items[1].GrossWeight != "5100" {
		t.Errorf("expected quantity/weight round-trip, got %q/%q", items[0].Quantity, items[1].GrossWeight)
	}
	if items[0].FarmProductID != "9" || items[0].CEST != "0100101" {
		t.Errorf("expected optional references to round-trip, got %q/%q", items[0].FarmProductID, items[0].CEST)
	}
}

func TestCommonProfileItemCFOP(t *testing.T) {
	if got, ok := commonProfileItemCFOP([]nfe_model.DetachedProfileItem{{CFOP: "5101"}, {CFOP: "5101"}}); !ok || got != "5101" {
		t.Fatalf("expected shared CFOP 5101, got %q (ok=%v)", got, ok)
	}
	if _, ok := commonProfileItemCFOP([]nfe_model.DetachedProfileItem{{CFOP: "5101"}, {CFOP: "5102"}}); ok {
		t.Fatal("expected mixed CFOPs to be ambiguous")
	}
	if _, ok := commonProfileItemCFOP(nil); ok {
		t.Fatal("expected an empty item list to be ambiguous")
	}
}

func TestPercentDisplayExact(t *testing.T) {
	if got := percentDisplayExact(decimal.Zero); got != "0.00" {
		t.Errorf("expected explicit zero to display as 0.00, got %q", got)
	}
	if got := percentDisplayExact(decimal.RequireFromString("0.17")); got != "17.00" {
		t.Errorf("expected 0.17 to display as 17.00, got %q", got)
	}
}

func TestDanfeStatusAllowed(t *testing.T) {
	for _, status := range []string{"authorized", "cancelled"} {
		if !danfeStatusAllowed(status) {
			t.Errorf("expected status %q to allow a DANFE", status)
		}
	}
	for _, status := range []string{"draft", "pending", "denied", "superseded", ""} {
		if danfeStatusAllowed(status) {
			t.Errorf("expected status %q to deny a DANFE", status)
		}
	}
}

func TestDetachedDANFEXML(t *testing.T) {
	authorizedXML := "<nfeProc>authorized</nfeProc>"
	signedXML := "<NFe>signed</NFe>"
	empty := ""

	t.Run("prefers authorized XML", func(t *testing.T) {
		invoice := &nfe_model.DetachedInvoice{XMLAuthorized: &authorizedXML, XMLSigned: &signedXML}
		if xml, ok := detachedDANFEXML(invoice); !ok || xml != authorizedXML {
			t.Errorf("expected authorized XML to win, got %q (ok=%v)", xml, ok)
		}
	})

	t.Run("falls back to signed XML", func(t *testing.T) {
		invoice := &nfe_model.DetachedInvoice{XMLSigned: &signedXML}
		if xml, ok := detachedDANFEXML(invoice); !ok || xml != signedXML {
			t.Errorf("expected signed XML fallback, got %q (ok=%v)", xml, ok)
		}
	})

	t.Run("authorized only", func(t *testing.T) {
		invoice := &nfe_model.DetachedInvoice{XMLAuthorized: &authorizedXML}
		if xml, ok := detachedDANFEXML(invoice); !ok || xml != authorizedXML {
			t.Errorf("expected authorized XML, got %q (ok=%v)", xml, ok)
		}
	})

	t.Run("no XML stored", func(t *testing.T) {
		invoice := &nfe_model.DetachedInvoice{}
		if _, ok := detachedDANFEXML(invoice); ok {
			t.Error("expected ok=false when neither XML is stored")
		}
	})

	t.Run("empty strings treated as absent", func(t *testing.T) {
		invoice := &nfe_model.DetachedInvoice{XMLAuthorized: &empty}
		if _, ok := detachedDANFEXML(invoice); ok {
			t.Error("expected an empty authorized XML to fall through to the signed one")
		}
		invoice = &nfe_model.DetachedInvoice{XMLAuthorized: &empty, XMLSigned: &empty}
		if _, ok := detachedDANFEXML(invoice); ok {
			t.Error("expected empty XML pointers to yield ok=false")
		}
	})
}
