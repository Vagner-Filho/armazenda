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

// TestNFeEmitPageRendersStepperContract guards the markup contract the stepper
// (assets/js/detachedNFeTabs.js) and its validation
// (assets/js/detachedNFeValidation.js) depend on:
//   - the four step chips, each with its live status slot;
//   - the Voltar / Avançar footer buttons;
//   - the preview submit starting hidden AND disabled — the only network action
//     of the form, and disabled off the last step so Enter stays a no-op;
//   - the rascunho form opting out of native validation, so a draft always
//     reaches the server and is reported with the server's pt-br toast.
func TestNFeEmitPageRendersStepperContract(t *testing.T) {
	parsed := parseAllTemplates(t)

	view := detachedPageView{
		HasNFEFarmConfig:   true,
		RecipientType:      "existing",
		Items:              []detachedPageItem{{CFOP: "5101", Unit: "KG"}},
		CfopOptions:        cfop_view.CfopOptions{Default: "5101"},
		UseDefaultTaxRates: true,
		CSPNonce:           "test-nonce",
		TierKey:            "fiscal",
	}
	out := executeBlock(t, parsed, "nfe-emit.html", view)

	for _, want := range []string{
		`id="wizard-back"`,
		`id="wizard-next"`,
		`data-step-status="dest"`,
		`data-step-status="transac"`,
		`data-step-status="transp"`,
		`data-step-status="trib"`,
		`data-test-id="preview-detached-nfe"`,
		`id="preview-nfe"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("emission page must render %s", want)
		}
	}

	preview := elementByID(t, out, "preview-nfe")
	if !strings.Contains(preview, "disabled") {
		t.Error("the preview submit must render disabled; only the last step enables it")
	}
	if !strings.Contains(preview, "hidden") {
		t.Error("the preview submit must render hidden; it only exists on the last step")
	}

	rascunho := elementByID(t, out, "save-rascunho-form")
	if !strings.Contains(rascunho, "novalidate") {
		t.Error("the rascunho form must render with novalidate so a draft always reaches the server")
	}

	// The recipient select is the one the page script toggles `required` on
	// (it is the recipient while "Selecionar existente" is active), so it must
	// render with its id and its placeholder option.
	if !strings.Contains(out, `id="recipientPersonId"`) {
		t.Error("emission page must render the recipient person select")
	}
	if !strings.Contains(out, `<option value="">Selecione um destinatário</option>`) {
		t.Error("the recipient person select must render its placeholder option")
	}
}

// elementByID returns the opening tag of the element carrying the given id,
// so attribute-level assertions do not depend on attribute ordering.
func elementByID(t *testing.T, out, id string) string {
	t.Helper()
	start := strings.Index(out, `id="`+id+`"`)
	if start < 0 {
		t.Fatalf("element with id %q not found; got:\n%s", id, out)
	}
	open := strings.LastIndex(out[:start], "<")
	if open < 0 {
		t.Fatalf("no opening tag before id %q; got:\n%s", id, out)
	}
	end := strings.Index(out[start:], ">")
	if end < 0 {
		t.Fatalf("unterminated tag for id %q; got:\n%s", id, out)
	}
	return out[open : start+end]
}

// transportInlineControlIDs are the inline Transportadora controls that the
// source switch hides and disables in "existing" mode.
var transportInlineControlIDs = []string{
	"transportNome",
	"transportDocumento",
	"transportIE",
	"transportLogradouro",
	"transportNumero",
	"transportMunicipio",
	"transportUF",
}

// recipientInlineControlIDs are the inline Destinatário controls that the
// source switch hides and disables in "existing" mode.
var recipientInlineControlIDs = []string{
	"recipientName",
	"recipientDocument",
	"recipientIE",
	"recipientPhone",
	"recipientEmail",
	"recipientStreet",
	"recipientNumber",
	"recipientNeighborhood",
	"recipientCity",
	"recipientState",
	"recipientCEP",
}

// emitPageBaseView returns a minimal detached page view for the template
// source-mode tests.
func emitPageBaseView() detachedPageView {
	return detachedPageView{
		HasNFEFarmConfig:   true,
		RecipientType:      "existing",
		TransportType:      "existing",
		Items:              []detachedPageItem{{CFOP: "5101", Unit: "KG"}},
		CfopOptions:        cfop_view.CfopOptions{Default: "5101"},
		UseDefaultTaxRates: true,
		CSPNonce:           "test-nonce",
		TierKey:            "fiscal",
	}
}

// TestNFeEmitPageTransportSource guards the Transporte source switch: the
// default is "Selecionar existente", the contact select is enabled with the
// "Nenhuma transportadora" option first, contacts without an id are skipped,
// and the inline fields are hidden and disabled. "Novo transportador" is the
// inverse. Every new control has a <label for>.
func TestNFeEmitPageTransportSource(t *testing.T) {
	parsed := parseAllTemplates(t)

	personID := uint32(7)
	otherID := uint32(8)
	people := []entity_public.PersonOption{
		{Id: &otherID, Name: "Transportadora B"},
		{Id: &personID, Name: "Transportadora A"},
		{Name: "Sem id"},
	}

	existingView := emitPageBaseView()
	existingView.TransportPeople = personOptionDisplays(people, nil)
	out := executeBlock(t, parsed, "nfe-emit.html", existingView)

	selectTag := elementByID(t, out, "transportPersonId")
	if strings.Contains(selectTag, "disabled") {
		t.Error("the contact select must be enabled in 'Selecionar existente' mode")
	}
	existingSection := elementByID(t, out, "existing-transport-section")
	if strings.Contains(existingSection, "hidden") {
		t.Error("the contact select section must be visible in 'Selecionar existente' mode")
	}
	newSection := elementByID(t, out, "new-transport-section")
	if !strings.Contains(newSection, "hidden") {
		t.Error("the inline section must be hidden in 'Selecionar existente' mode")
	}
	for _, id := range transportInlineControlIDs {
		tag := elementByID(t, out, id)
		if !strings.Contains(tag, "disabled") {
			t.Errorf("inline control %s must be disabled in 'Selecionar existente' mode", id)
		}
		if !strings.Contains(out, `for="`+id+`"`) {
			t.Errorf("inline control %s must have a <label for>", id)
		}
	}

	if !strings.Contains(out, `<option value="">Nenhuma transportadora</option>`) {
		t.Error("the contact select must render the 'Nenhuma transportadora' option first")
	}
	if strings.Index(out, `<option value="">Nenhuma transportadora</option>`) > strings.Index(out, `<option value="7"`) {
		t.Error("the empty option must come before the contact options")
	}
	if !strings.Contains(out, `<option value="7">Transportadora A</option>`) ||
		!strings.Contains(out, `<option value="8">Transportadora B</option>`) {
		t.Error("the contact select must list the farm contacts")
	}
	if strings.Contains(out, "Sem id") {
		t.Error("contacts without an id must be skipped")
	}
	if !strings.Contains(out, `for="transportPersonId"`) {
		t.Error("the contact select must have a <label for>")
	}

	newView := emitPageBaseView()
	newView.TransportType = "new"
	newView.TransportPeople = personOptionDisplays(people, nil)
	out = executeBlock(t, parsed, "nfe-emit.html", newView)

	if !strings.Contains(elementByID(t, out, "transportPersonId"), "disabled") {
		t.Error("the contact select must be disabled in 'Novo transportador' mode")
	}
	if !strings.Contains(elementByID(t, out, "existing-transport-section"), "hidden") {
		t.Error("the contact select section must be hidden in 'Novo transportador' mode")
	}
	if strings.Contains(elementByID(t, out, "new-transport-section"), "hidden") {
		t.Error("the inline section must be visible in 'Novo transportador' mode")
	}
	for _, id := range transportInlineControlIDs {
		if strings.Contains(elementByID(t, out, id), "disabled") {
			t.Errorf("inline control %s must be enabled in 'Novo transportador' mode", id)
		}
	}

	// The vehicle and volume fields stay enabled in both modes.
	for _, id := range []string{"transportPlaca", "volQtd", "volEsp", "volMarca", "volPesoL", "volPesoB"} {
		if strings.Contains(elementByID(t, out, id), "disabled") {
			t.Errorf("%s must stay enabled in every transport source mode", id)
		}
	}
}

// TestNFeEmitPageRecipientSource guards the Destinatário source switch: the
// person select is enabled only in "Selecionar existente" and the inline
// controls only in "Novo destinatário". Both states are rendered by the
// server.
func TestNFeEmitPageRecipientSource(t *testing.T) {
	parsed := parseAllTemplates(t)

	view := emitPageBaseView()
	out := executeBlock(t, parsed, "nfe-emit.html", view)

	if strings.Contains(elementByID(t, out, "recipientPersonId"), "disabled") {
		t.Error("the recipient select must be enabled in 'Selecionar existente' mode")
	}
	if strings.Contains(elementByID(t, out, "new-recipient-section"), "hidden") == false {
		t.Error("the inline recipient section must be hidden in 'Selecionar existente' mode")
	}
	for _, id := range recipientInlineControlIDs {
		if !strings.Contains(elementByID(t, out, id), "disabled") {
			t.Errorf("inline recipient control %s must be disabled in 'Selecionar existente' mode", id)
		}
		if !strings.Contains(out, `for="`+id+`"`) {
			t.Errorf("inline recipient control %s must have a <label for>", id)
		}
	}

	view = emitPageBaseView()
	view.RecipientType = "new"
	out = executeBlock(t, parsed, "nfe-emit.html", view)

	if !strings.Contains(elementByID(t, out, "recipientPersonId"), "disabled") {
		t.Error("the recipient select must be disabled in 'Novo destinatário' mode")
	}
	if strings.Contains(elementByID(t, out, "new-recipient-section"), "hidden") {
		t.Error("the inline recipient section must be visible in 'Novo destinatário' mode")
	}
	for _, id := range recipientInlineControlIDs {
		if strings.Contains(elementByID(t, out, id), "disabled") {
			t.Errorf("inline recipient control %s must be enabled in 'Novo destinatário' mode", id)
		}
	}
	if !strings.Contains(out, `for="recipientPersonId"`) {
		t.Error("the recipient select must have a <label for>")
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

// TestDetachedPreviewTransportPerson guards the preview transport round-trip:
// a picked contact carries only the hidden contact id plus the summary line,
// while inline data carries the inline hidden inputs as before.
func TestDetachedPreviewTransportPerson(t *testing.T) {
	parsed := parseAllTemplates(t)

	contactView := detachedPreviewView{
		ItemCount:      1,
		TransportLabel: "Transportes XYZ Ltda",
		Transport: detachedTransportFields{
			PersonID: "42",
			Placa:    "ABC1D23",
		},
		TPag:     "90",
		CSPNonce: "test-nonce",
	}
	out := executeBlock(t, parsed, "nfe-detached-preview", contactView)

	if !strings.Contains(out, `name="transportPersonId" value="42"`) {
		t.Errorf("a picked contact must round-trip its id; got:\n%s", out)
	}
	if strings.Contains(out, `name="transportNome"`) {
		t.Error("a picked contact must not round-trip the inline transportadora fields")
	}
	if !strings.Contains(out, "Transportadora:</strong> Transportes XYZ Ltda") {
		t.Errorf("the summary must show the transportadora name; got:\n%s", out)
	}
	if !strings.Contains(out, `name="transportPlaca" value="ABC1D23"`) {
		t.Error("the vehicle fields must round-trip for a picked contact")
	}

	inlineView := detachedPreviewView{
		ItemCount:      1,
		TransportLabel: "Carreteiro Fulano",
		Transport: detachedTransportFields{
			Nome:     "Carreteiro Fulano",
			Document: "529.982.247-25",
			UF:       "MT",
		},
		TPag:     "90",
		CSPNonce: "test-nonce",
	}
	out = executeBlock(t, parsed, "nfe-detached-preview", inlineView)

	if strings.Contains(out, `name="transportPersonId"`) {
		t.Error("inline data must not render a transport contact id")
	}
	if !strings.Contains(out, `name="transportNome" value="Carreteiro Fulano"`) ||
		!strings.Contains(out, `name="transportDocumento" value="529.982.247-25"`) ||
		!strings.Contains(out, `name="transportUF" value="MT"`) {
		t.Errorf("inline data must round-trip its fields; got:\n%s", out)
	}
	if !strings.Contains(out, "Transportadora:</strong> Carreteiro Fulano") {
		t.Errorf("the summary must show the typed transportadora name; got:\n%s", out)
	}

	emptyView := detachedPreviewView{ItemCount: 1, TPag: "90", CSPNonce: "test-nonce"}
	out = executeBlock(t, parsed, "nfe-detached-preview", emptyView)
	if strings.Contains(out, "Transportadora:</strong>") {
		t.Error("the summary must omit the transportadora line when none is informed")
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
