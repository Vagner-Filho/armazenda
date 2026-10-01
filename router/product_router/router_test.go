package product_router

import (
	"errors"
	"html/template"
	"testing"
)

// testFuncMap mirrors the app's template FuncMap (main.go) — templates use
// custom funcs like dict/deref.
func testFuncMap() template.FuncMap {
	return template.FuncMap{
		"deref": func(p *uint32) uint32 {
			if p == nil {
				return 0
			}
			return *p
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			dict := make(map[string]interface{}, len(values))
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				if i+1 >= len(values) {
					return nil, errors.New("dict needs an even number of arguments")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// TestProductTemplatesParse guards against template syntax errors, which
// would otherwise only surface on server startup.
func TestProductTemplatesParse(t *testing.T) {
	tmpl := template.New("test").Funcs(testFuncMap())
	templates := []string{
		"../../templates/product/product-list-item.html",
		"../../templates/product/product-form.html",
		"../../templates/pages/produto.html",
	}
	for _, path := range templates {
		if _, err := tmpl.ParseFiles(path); err != nil {
			t.Errorf("template %s failed to parse: %v", path, err)
		}
	}
}
