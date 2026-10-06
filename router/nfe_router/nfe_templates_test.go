package nfe_router

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
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
