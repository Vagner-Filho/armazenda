package nfe_router

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	entity_public "armazenda/entity/public"
	cfop_view "armazenda/view/cfop"

	"github.com/gin-gonic/gin"
)

// assertCfopToast asserts the response carries the expected status and a
// HX-Trigger toast whose message/hint contains want.
func assertCfopToast(t *testing.T, c *gin.Context, status int, want string) {
	t.Helper()
	if c.Writer.Status() != status {
		t.Fatalf("status = %d, want %d", c.Writer.Status(), status)
	}
	header := c.Writer.Header().Get("HX-Trigger")
	if header == "" {
		t.Fatal("expected an HX-Trigger toast header")
	}
	if !strings.Contains(header, want) {
		t.Fatalf("toast %q does not contain %q", header, want)
	}
}

func TestAddCfop_RejectsMalformedCodes(t *testing.T) {
	cases := []string{"510", "51011", "51A1", "0101", "4101", "8101", "9101", ""}
	for _, code := range cases {
		t.Run("code_"+code, func(t *testing.T) {
			values := url.Values{
				"code":              {code},
				"description":       {"Operação teste"},
				"originDestination": {"Mesmo estado"},
			}
			c := newFormContext(t, values)
			addCfop(c)
			assertCfopToast(t, c, http.StatusBadRequest, "CFOP inválido")
		})
	}
}

func TestAddCfop_RejectsEmptyDescription(t *testing.T) {
	values := url.Values{
		"code":              {"5101"},
		"description":       {"   "},
		"originDestination": {"Mesmo estado"},
	}
	c := newFormContext(t, values)
	addCfop(c)
	assertCfopToast(t, c, http.StatusBadRequest, "Descrição obrigatória")
}

func TestAddCfop_RejectsLongDescription(t *testing.T) {
	values := url.Values{
		"code":              {"5101"},
		"description":       {strings.Repeat("a", cfopDescriptionMaxLength+1)},
		"originDestination": {"Mesmo estado"},
	}
	c := newFormContext(t, values)
	addCfop(c)
	assertCfopToast(t, c, http.StatusBadRequest, "Descrição muito longa")
}

func TestAddCfop_RejectsOriginDestinationMismatch(t *testing.T) {
	cases := []url.Values{
		{"code": {"5101"}, "description": {"Operação teste"}, "originDestination": {"Outro estado"}},
		{"code": {"5101"}, "description": {"Operação teste"}, "originDestination": {"Exterior"}},
		{"code": {"5101"}, "description": {"Operação teste"}, "originDestination": {""}},
		{"code": {"6102"}, "description": {"Operação teste"}, "originDestination": {"Mesmo estado"}},
	}
	for i, values := range cases {
		c := newFormContext(t, values)
		addCfop(c)
		if c.Writer.Status() != http.StatusBadRequest {
			t.Fatalf("case %d: status = %d, want 400", i, c.Writer.Status())
		}
		if header := c.Writer.Header().Get("HX-Trigger"); !strings.Contains(header, "Origem/Destino inválido") {
			t.Fatalf("case %d: unexpected toast %q", i, header)
		}
	}
}

func TestCfopOptionsTemplate_RendersFarmGroups(t *testing.T) {
	parsed := parseAllTemplates(t)
	data := map[string]interface{}{
		"Options": cfop_view.CfopOptions{
			MostUsed: []entity_public.CfopOption{
				{Code: "5101", Description: "Venda de produção do estabelecimento", UseCount: 3},
			},
			All: []entity_public.CfopOption{
				{Code: "5101", Description: "Venda de produção do estabelecimento", UseCount: 3},
				{Code: "6102", Description: "Venda de mercadoria adquirida ou recebida de terceiros"},
				{Code: "7999", Description: "Operação da fazenda", IsFarm: true},
			},
		},
		"Selected": "6102",
	}

	out := executeBlock(t, parsed, "cfop-options", data)
	for _, want := range []string{
		`<optgroup label="Mais utilizados">`,
		`<optgroup label="Todos os CFOPs">`,
		`value="5101"`,
		`5101 — Venda de produção do estabelecimento`,
		`value="6102" selected`,
		`value="7999" data-cfop-farm="true"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("cfop-options missing %q; got:\n%s", want, out)
		}
	}
}

func TestCfopOptionsTemplate_OmitsEmptyMostUsedGroup(t *testing.T) {
	parsed := parseAllTemplates(t)
	out := executeBlock(t, parsed, "cfop-options", map[string]interface{}{
		"Options": cfop_view.CfopOptions{
			All: []entity_public.CfopOption{{Code: "5101", Description: "Venda de produção do estabelecimento"}},
		},
		"Selected": "5101",
	})

	if strings.Contains(out, "Mais utilizados") {
		t.Errorf("empty most-used group must not render; got:\n%s", out)
	}
	if !strings.Contains(out, `value="5101" selected`) {
		t.Errorf("selected option missing; got:\n%s", out)
	}
}

func TestCfopOptionsTemplate_KeepsUnknownSelectedCode(t *testing.T) {
	parsed := parseAllTemplates(t)
	out := executeBlock(t, parsed, "cfop-options", map[string]interface{}{
		"Options": cfop_view.CfopOptions{
			All: []entity_public.CfopOption{{Code: "5101", Description: "Venda de produção do estabelecimento"}},
		},
		"Selected": "9999",
	})

	if !strings.Contains(out, `value="9999" selected`) {
		t.Errorf("a selected code outside the catalog must survive; got:\n%s", out)
	}
}

func TestCfopSelectorTemplate_Contract(t *testing.T) {
	parsed := parseAllTemplates(t)
	out := executeBlock(t, parsed, "cfop-selector", map[string]interface{}{
		"Options": cfop_view.CfopOptions{
			All:     []entity_public.CfopOption{{Code: "5101", Description: "Venda de produção do estabelecimento"}},
			Default: "5101",
		},
		"Selected": "5101",
		"Name":     "items[0].cfop",
	})

	for _, want := range []string{
		`name="items[0].cfop"`,
		`required`,
		`data-cfop-select`,
		`class="cfop-search`,
		`hx-get="/nfe/cfop/form"`,
		`hx-target="body"`,
		`hx-swap="beforeend"`,
		`data-test-id="cfop-add"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("cfop-selector missing %q; got:\n%s", want, out)
		}
	}
}

func TestCfopFormTemplate_Contract(t *testing.T) {
	parsed := parseAllTemplates(t)
	out := executeBlock(t, parsed, "nfe-cfop-form", map[string]interface{}{"CSPNonce": "test-nonce"})

	for _, want := range []string{
		`hx-post="/nfe/cfop"`,
		`name="code"`,
		`pattern="[0-9]{4}"`,
		`maxlength="4"`,
		`name="description"`,
		`maxlength="200"`,
		`name="originDestination"`,
		`Mesmo estado`,
		`Outro estado`,
		`Exterior`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("nfe-cfop-form missing %q; got:\n%s", want, out)
		}
	}
}

func TestCfopOptionTemplate_Contract(t *testing.T) {
	parsed := parseAllTemplates(t)

	farm := executeBlock(t, parsed, "cfop-option", entity_public.CfopOption{
		Code: "9999", Description: "Operação da fazenda", IsFarm: true,
	})
	if !strings.Contains(farm, `value="9999"`) || !strings.Contains(farm, `data-cfop-farm="true"`) {
		t.Errorf("farm option fragment missing farm marker; got:\n%s", farm)
	}

	catalog := executeBlock(t, parsed, "cfop-option", entity_public.CfopOption{
		Code: "5101", Description: "Venda de produção do estabelecimento",
	})
	if strings.Contains(catalog, "data-cfop-farm") {
		t.Errorf("catalog option must not render the farm marker; got:\n%s", catalog)
	}
}
