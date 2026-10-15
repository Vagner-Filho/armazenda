package nfe_router

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	entity_public "armazenda/entity/public"
	"armazenda/model/nfe_model"
	cfop_view "armazenda/view/cfop"
)

// nfeTemplateFuncs mirrors the application template helpers used by the NF-e
// templates so they can be executed in isolation.
func nfeTemplateFuncs() template.FuncMap {
	dict := func(values ...interface{}) (map[string]interface{}, error) {
		out := make(map[string]interface{}, len(values)/2)
		for i := 0; i+1 < len(values); i += 2 {
			out[values[i].(string)] = values[i+1]
		}
		return out, nil
	}
	return template.FuncMap{
		"deref": func(p *uint32) uint32 {
			if p == nil {
				return 0
			}
			return *p
		},
		"ptrString": func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		},
		"dict": dict,
		"decIsZero": func(v interface{}) bool {
			if d, ok := v.(decimal.Decimal); ok {
				return d.IsZero()
			}
			return false
		},
		"decIsNotZero": func(v interface{}) bool {
			if d, ok := v.(decimal.Decimal); ok {
				return !d.IsZero()
			}
			return false
		},
	}
}

// parseAllTemplates parses every repository template (same pattern as
// main.go: templates/*.html plus templates/**/*.html).
func parseAllTemplates(t *testing.T) *template.Template {
	t.Helper()
	var files []string
	top, err := filepath.Glob("../../templates/*.html")
	if err != nil {
		t.Fatalf("glob top-level templates: %v", err)
	}
	nested, err := filepath.Glob("../../templates/*/*.html")
	if err != nil {
		t.Fatalf("glob nested templates: %v", err)
	}
	files = append(files, top...)
	files = append(files, nested...)
	if len(files) == 0 {
		t.Fatal("no templates found")
	}
	parsed, err := template.New("").Funcs(nfeTemplateFuncs()).ParseFiles(files...)
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	return parsed
}

func executeBlock(t *testing.T, parsed *template.Template, block string, data interface{}) string {
	t.Helper()
	var buf bytes.Buffer
	if err := parsed.ExecuteTemplate(&buf, block, data); err != nil {
		t.Fatalf("execute %s: %v", block, err)
	}
	return buf.String()
}

// TestNFeRascunhoFormSelectsSavedRecipient guards the editor's recipient
// <select>: the saved profile recipient must render as selected. Regression
// for the template-side coercion `eq (printf "%s" .Id) $.RecipientSelected`,
// where PersonOption.Id (*uint32) rendered as "%!s(*uint32=0x…)" under %s and
// never matched the string id, so no option was ever selected.
func TestNFeRascunhoFormSelectsSavedRecipient(t *testing.T) {
	parsed := parseAllTemplates(t)

	recipientID := uint32(7)
	otherID := uint32(8)
	profile := &nfe_model.DetachedProfile{
		Name:        "Rascunho de teste",
		RecipientID: &recipientID,
	}
	people := []entity_public.PersonOption{
		{Id: &otherID, Name: "Outra pessoa"},
		{Id: &recipientID, Name: "Itiquira"},
		{Name: "Sem id"},
	}

	view := detachedProfileFormView(profile, people, nil, cfop_view.CfopOptions{Default: "5101"}, "test-nonce")
	out := executeBlock(t, parsed, "nfe-rascunho-form", view)

	if !strings.Contains(out, `<option value="7" selected>Itiquira</option>`) {
		t.Errorf("saved recipient (id 7) must render selected; got:\n%s", out)
	}
	if strings.Contains(out, `<option value="8" selected`) {
		t.Error("option of a non-recipient person (id 8) must not render selected")
	}
	if strings.Contains(out, "Sem id") {
		t.Error("contact option without an id must be skipped")
	}
}

// TestNFeEmitPageSelectsSavedRecipient guards the emission page's recipient
// <select>: a rascunho's saved recipient must render as selected. Regression
// for the template-side coercion `eq (printf "%v" .Id) $.SelectedPersonID`,
// where PersonOption.Id (*uint32) rendered as a pointer address under %v and
// never matched the string id, so no option was ever selected.
func TestNFeEmitPageSelectsSavedRecipient(t *testing.T) {
	parsed := parseAllTemplates(t)

	recipientID := uint32(7)
	otherID := uint32(8)
	people := []entity_public.PersonOption{
		{Id: &otherID, Name: "Outra pessoa"},
		{Id: &recipientID, Name: "Itiquira"},
		{Name: "Sem id"},
	}

	view := detachedPageView{
		HasNFEFarmConfig:   true,
		People:             personOptionDisplays(people, &recipientID),
		RecipientType:      "existing",
		Items:              []detachedPageItem{{CFOP: "5101", Unit: "KG"}},
		CfopOptions:        cfop_view.CfopOptions{Default: "5101"},
		UseDefaultTaxRates: true,
		CSPNonce:           "test-nonce",
		TierKey:            "fiscal",
	}
	out := executeBlock(t, parsed, "nfe-emit.html", view)

	if !strings.Contains(out, `<option value="7" selected>Itiquira</option>`) {
		t.Errorf("saved recipient (id 7) must render selected; got:\n%s", out)
	}
	if strings.Contains(out, `<option value="8" selected`) {
		t.Error("option of a non-recipient person (id 8) must not render selected")
	}
	if strings.Contains(out, "Sem id") {
		t.Error("contact option without an id must be skipped")
	}
}

// TestNFeEmitPageRendersPerItemInfAdProd guards the per-item additional
// information input on the emission page: every item row renders it with the
// row-indexed name and test id, and a rascunho prefill value reaches the
// input.
func TestNFeEmitPageRendersPerItemInfAdProd(t *testing.T) {
	parsed := parseAllTemplates(t)

	view := detachedPageView{
		HasNFEFarmConfig: true,
		RecipientType:    "existing",
		Items: []detachedPageItem{
			{CFOP: "5101", Unit: "KG", InfAdProd: "SERIE ABC123"},
			{CFOP: "5102", Unit: "KG"},
		},
		CfopOptions:        cfop_view.CfopOptions{Default: "5101"},
		UseDefaultTaxRates: true,
		CSPNonce:           "test-nonce",
		TierKey:            "fiscal",
	}
	out := executeBlock(t, parsed, "nfe-emit.html", view)

	for _, want := range []string{
		`name="items[0].infAdProd"`,
		`data-test-id="item-inf-ad-prod-0"`,
		`value="SERIE ABC123"`,
		`name="items[1].infAdProd"`,
		`data-test-id="item-inf-ad-prod-1"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("emission page must render %s", want)
		}
	}
}

// TestNFeDetachedPreviewRendersInfAdProd guards the preview round-trip on the
// fragment side: the parsed per-item note must render into the hidden
// items[i].infAdProd input so the confirm step re-submits exactly the
// previewed text.
func TestNFeDetachedPreviewRendersInfAdProd(t *testing.T) {
	parsed := parseAllTemplates(t)

	view := detachedPreviewView{
		ItemCount: 2,
		Items: []detachedPreviewItem{
			{Index: 0, ProductName: "Soja", CFOP: "5101", Quantity: "10", GrossWeight: "10", UnitPrice: "50", Unit: "KG", InfAdProd: "SERIE ABC123"},
			{Index: 1, ProductName: "Milho", CFOP: "5102", Quantity: "5", GrossWeight: "5", UnitPrice: "30", Unit: "KG"},
		},
		TPag:     "90",
		CSPNonce: "test-nonce",
	}
	out := executeBlock(t, parsed, "nfe-detached-preview", view)

	if !strings.Contains(out, `name="items[0].infAdProd" value="SERIE ABC123"`) {
		t.Errorf("preview must round-trip the item note; got:\n%s", out)
	}
	if !strings.Contains(out, `name="items[1].infAdProd" value=""`) {
		t.Errorf("preview must round-trip a blank item note as empty; got:\n%s", out)
	}
}

// TestNFEDraftStatusLabelRendersNaoEnviada is the template sweep guard: every
// NF-e status rendering must show "Não enviada" for the internal 'draft'
// status and no NF-e surface may label that status "Rascunho".
func TestNFEDraftStatusLabelRendersNaoEnviada(t *testing.T) {
	parsed := parseAllTemplates(t)

	invoice := map[string]interface{}{
		"AccessKey":     "12345678901234567890123456789012345678901234",
		"Serie":         1,
		"Number":        10,
		"Status":        "draft",
		"TotalValue":    decimal.RequireFromString("100.00"),
		"XMLSigned":     nil,
		"XMLAuthorized": nil,
	}

	list := executeBlock(t, parsed, "nfe-list-item", map[string]interface{}{
		"Invoice": invoice,
		"Type":    "Avulsa",
	})
	if !strings.Contains(list, "Não enviada") {
		t.Errorf("nfe-list-item must render 'Não enviada' for draft; got:\n%s", list)
	}
	if strings.Contains(list, ">Rascunho<") {
		t.Error("nfe-list-item must not render a 'Rascunho' status label")
	}

	modal := executeBlock(t, parsed, "nfe-existing-modal", map[string]interface{}{
		"DepartureID": 1,
		"Invoice":     invoice,
		"CSPNonce":    "test-nonce",
	})
	if !strings.Contains(modal, "Não enviada") {
		t.Errorf("nfe-existing-modal must render 'Não enviada' for draft; got:\n%s", modal)
	}
	if strings.Contains(modal, ">Rascunho<") {
		t.Error("nfe-existing-modal must not render a 'Rascunho' status label")
	}
}

// TestNFETemplatesDraftSweep ensures no NF-e template still pairs the draft
// status with the word "Rascunho".
func TestNFETemplatesDraftSweep(t *testing.T) {
	entries, err := os.ReadDir("../../templates/nfe")
	if err != nil {
		t.Fatalf("read nfe templates: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("../../templates/nfe", entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		text := string(content)
		// The rascunho feature intentionally uses the word; only flag it when
		// paired with the draft status branch.
		if strings.Contains(text, `eq .Invoice.Status "draft"`) && strings.Contains(text, ">Rascunho<") {
			t.Errorf("%s renders the draft status as 'Rascunho'", entry.Name())
		}
	}
}

// TestRomaneioRascunhoWordingUnchanged guards the out-of-scope boundary: the
// entry/departure rascunho UI keeps its original wording.
func TestRomaneioRascunhoWordingUnchanged(t *testing.T) {
	entryDraft, err := os.ReadFile("../../templates/entry/entry-draft-list-item.html")
	if err != nil {
		t.Fatalf("read entry draft list item: %v", err)
	}
	if !strings.Contains(string(entryDraft), "Rascunho") {
		t.Error("entry rascunho wording must stay unchanged")
	}

	departureDraft, err := os.ReadFile("../../templates/departure/departure-draft-list-item.html")
	if err != nil {
		t.Fatalf("read departure draft list item: %v", err)
	}
	if !strings.Contains(string(departureDraft), "Rascunho") {
		t.Error("departure rascunho wording must stay unchanged")
	}
}
