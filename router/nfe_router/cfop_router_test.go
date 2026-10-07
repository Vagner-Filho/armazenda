package nfe_router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
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
				"code":        {code},
				"description": {"Operação teste"},
			}
			c := newFormContext(t, values)
			addCfop(c)
			assertCfopToast(t, c, http.StatusBadRequest, "CFOP inválido")
		})
	}
}

func TestAddCfop_RejectsEmptyDescription(t *testing.T) {
	values := url.Values{
		"code":        {"5101"},
		"description": {"   "},
	}
	c := newFormContext(t, values)
	addCfop(c)
	assertCfopToast(t, c, http.StatusBadRequest, "Descrição obrigatória")
}

func TestAddCfop_RejectsLongDescription(t *testing.T) {
	values := url.Values{
		"code":        {"5101"},
		"description": {strings.Repeat("a", cfopDescriptionMaxLength+1)},
	}
	c := newFormContext(t, values)
	addCfop(c)
	assertCfopToast(t, c, http.StatusBadRequest, "Descrição muito longa")
}

type recordingCfopRegistrar struct {
	addedFarmID      uint32
	addedCode        string
	addedDescription string
	addedOriginDest  string
	catalogHasCode   bool
}

func (m *recordingCfopRegistrar) CatalogHasCfop(string) (bool, *model_error.ModelError) {
	return m.catalogHasCode, nil
}

func (m *recordingCfopRegistrar) AddFarmCfop(farmID uint32, code, description, originDestination string) (entity_public.CfopOption, *model_error.ModelError) {
	m.addedFarmID = farmID
	m.addedCode = code
	m.addedDescription = description
	m.addedOriginDest = originDestination
	return entity_public.CfopOption{
		Code:              code,
		Description:       description,
		OriginDestination: originDestination,
		IsFarm:            true,
	}, nil
}

func TestAddCfop_DerivesOriginDestinationAndIgnoresPostedValue(t *testing.T) {
	cases := []struct {
		name           string
		postedOrigin   string
		expectedOrigin string
	}{
		{name: "origin omitted", expectedOrigin: "Outro estado"},
		{name: "forged origin ignored", postedOrigin: "Exterior", expectedOrigin: "Outro estado"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			values := url.Values{
				"code":        {"6102"},
				"description": {"Operação teste"},
			}
			if testCase.postedOrigin != "" {
				values.Set("originDestination", testCase.postedOrigin)
			}

			registrar := &recordingCfopRegistrar{}
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.SetHTMLTemplate(parseAllTemplates(t))
			router.POST("/nfe/cfop", func(c *gin.Context) {
				addCfopWithProvider(c, func() (cfopRegistrar, uint32) {
					return registrar, 42
				})
			})

			req := httptest.NewRequest(http.MethodPost, "/nfe/cfop", strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			if response.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
			}
			if registrar.addedFarmID != 42 {
				t.Errorf("farm ID = %d, want 42", registrar.addedFarmID)
			}
			if registrar.addedOriginDest != testCase.expectedOrigin {
				t.Errorf("persisted origin = %q, want %q", registrar.addedOriginDest, testCase.expectedOrigin)
			}
			if !strings.Contains(response.Body.String(), `data-origin-destination="Outro estado"`) {
				t.Errorf("response option missing derived origin metadata: %s", response.Body.String())
			}
		})
	}
}

func TestCfopOptionsTemplate_RendersFarmGroups(t *testing.T) {
	parsed := parseAllTemplates(t)
	data := map[string]interface{}{
		"Options": cfop_view.CfopOptions{
			MostUsed: []entity_public.CfopOption{
				{Code: "5101", Description: "Venda de produção do estabelecimento", OriginDestination: "Mesmo estado", UseCount: 3},
			},
			All: []entity_public.CfopOption{
				{Code: "5101", Description: "Venda de produção do estabelecimento", OriginDestination: "Mesmo estado", UseCount: 3},
				{Code: "6102", Description: "Venda de mercadoria adquirida ou recebida de terceiros", OriginDestination: "Outro estado"},
				{Code: "7999", Description: "Operação da fazenda", OriginDestination: "Exterior", IsFarm: true},
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
		`value="6102" data-origin-destination="Outro estado" selected`,
		`value="7999" data-origin-destination="Exterior" data-cfop-farm="true"`,
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
			All: []entity_public.CfopOption{{Code: "5101", Description: "Venda de produção do estabelecimento", OriginDestination: "Mesmo estado"}},
		},
		"Selected": "5101",
	})

	if strings.Contains(out, "Mais utilizados") {
		t.Errorf("empty most-used group must not render; got:\n%s", out)
	}
	if !strings.Contains(out, `value="5101" data-origin-destination="Mesmo estado" selected`) {
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
		`role="combobox"`,
		`data-cfop-trigger`,
		`popover="manual"`,
		`data-cfop-options-list`,
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
		`<output`,
		`id="cfop-origin"`,
		`aria-live="polite"`,
		`Informe um CFOP válido.`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("nfe-cfop-form missing %q; got:\n%s", want, out)
		}
	}
	if strings.Contains(out, `name="originDestination"`) {
		t.Errorf("origin destination must not be submitted as an editable form field; got:\n%s", out)
	}
}

func TestCfopOptionTemplate_Contract(t *testing.T) {
	parsed := parseAllTemplates(t)

	farm := executeBlock(t, parsed, "cfop-option", entity_public.CfopOption{
		Code: "7999", Description: "Operação da fazenda", OriginDestination: "Exterior", IsFarm: true,
	})
	if !strings.Contains(farm, `value="7999"`) || !strings.Contains(farm, `data-origin-destination="Exterior"`) || !strings.Contains(farm, `data-cfop-farm="true"`) {
		t.Errorf("farm option fragment missing farm marker; got:\n%s", farm)
	}

	catalog := executeBlock(t, parsed, "cfop-option", entity_public.CfopOption{
		Code: "5101", Description: "Venda de produção do estabelecimento",
	})
	if strings.Contains(catalog, "data-cfop-farm") {
		t.Errorf("catalog option must not render the farm marker; got:\n%s", catalog)
	}
}
