package nfe_service

import (
	"errors"
	"strings"
	"testing"

	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// detached_wiring_test.go covers the pure wiring helpers asserted by
// docs/sdd/detached_nfe_input_wiring: payment composition, item discounts,
// finNFe derivation, NF-e reference keys, CST whitelists and the entrega
// mapping. All offline; no model singletons involved.

func TestResolveDetachedTPag(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", "90"},   // empty keeps the legacy Sem pagamento behavior
		{"01", "01"}, // Dinheiro
		{"90", "90"}, // Sem pagamento
		{"99", "99"}, // Outros
		{"02", "02"}, // pass-through — whitelist enforcement happens in the router
	}
	for _, tt := range tests {
		if got := resolveDetachedTPag(tt.input); got != tt.want {
			t.Errorf("resolveDetachedTPag(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidDetachedTPag(t *testing.T) {
	valid := []string{"01", "90", "99"}
	for _, tpag := range valid {
		if !ValidDetachedTPag(tpag) {
			t.Errorf("ValidDetachedTPag(%q) = false, want true", tpag)
		}
	}
	for _, tpag := range []string{"", "02", "03", "15", "x9"} {
		if ValidDetachedTPag(tpag) {
			t.Errorf("ValidDetachedTPag(%q) = true, want false", tpag)
		}
	}
}

func TestValidateItemDiscounts(t *testing.T) {
	item := func(discount, quantity, unitPrice string) DetachedItemInput {
		v, err := decimal.NewFromString(discount)
		if err != nil {
			t.Fatalf("bad discount %q", discount)
		}
		q, err := decimal.NewFromString(quantity)
		if err != nil {
			t.Fatalf("bad quantity %q", quantity)
		}
		p, err := decimal.NewFromString(unitPrice)
		if err != nil {
			t.Fatalf("bad unitPrice %q", unitPrice)
		}
		return DetachedItemInput{Quantity: q, UnitPrice: p, Discount: v}
	}

	if errs := validateItemDiscounts([]DetachedItemInput{item("0", "100", "10")}); len(errs) != 0 {
		t.Errorf("zero discount: unexpected errors %v", errs)
	}
	// Discount smaller than gross → ok.
	if errs := validateItemDiscounts([]DetachedItemInput{item("50", "100", "10")}); len(errs) != 0 {
		t.Errorf("valid discount: unexpected errors %v", errs)
	}
	// Negative → named message.
	errs := validateItemDiscounts([]DetachedItemInput{item("-1", "100", "10")})
	if len(errs) != 1 || !strings.Contains(errs[0], "não pode ser negativo") {
		t.Errorf("negative discount: got %v", errs)
	}
	// Discount ≥ gross → named message (equality also rejects: net must
	// stay positive).
	for _, bad := range []string{"1000", "1000.01"} {
		errs = validateItemDiscounts([]DetachedItemInput{item(bad, "100", "10")})
		if len(errs) != 1 || !strings.Contains(errs[0], "menor que o valor bruto") {
			t.Errorf("discount %s: got %v", bad, errs)
		}
	}
	// Multi-item: the message indexes the offending item (1-based).
	errs = validateItemDiscounts([]DetachedItemInput{
		item("50", "10", "10"), item("-3", "1", "10"),
	})
	if len(errs) != 1 || !strings.Contains(errs[0], "Item 2:") {
		t.Errorf("multi-item errors: got %v", errs)
	}
}

func TestResolveDetachedFinNFe(t *testing.T) {
	item := func(cfop string) DetachedItemInput { return DetachedItemInput{CFOP: cfop} }

	// All normal → 1.
	if got, errs := resolveDetachedFinNFe([]DetachedItemInput{item("5101"), item("5102")}); got != "1" || len(errs) != 0 {
		t.Errorf("normal items: finNFe %q errs %v", got, errs)
	}
	// All devolução → 4.
	if got, errs := resolveDetachedFinNFe([]DetachedItemInput{item("5556"), item("6556")}); got != "4" || len(errs) != 0 {
		t.Errorf("devolução items: finNFe %q errs %v", got, errs)
	}
	// Mixed → rejection naming the mixed item.
	got, errs := resolveDetachedFinNFe([]DetachedItemInput{item("5556"), item("5101"), item("6202")})
	if got != "" {
		t.Errorf("mixed set should not resolve, got %q", got)
	}
	if len(errs) != 1 || !strings.Contains(errs[0], "Item 2") || !strings.Contains(errs[0], "CFOP 5101") {
		t.Errorf("mixed set message: got %v", errs)
	}
	// Empty input → default normal.
	if got, errs := resolveDetachedFinNFe(nil); got != "1" || len(errs) != 0 {
		t.Errorf("empty items: finNFe %q errs %v", got, errs)
	}
}

func TestParseNFRefs(t *testing.T) {
	if refs, errs := ParseNFRefs(""); refs != nil || errs != nil {
		t.Errorf("empty input: refs %v errs %v", refs, errs)
	}
	if refs, errs := ParseNFRefs("   \n\t "); refs != nil || errs != nil {
		t.Errorf("blank input: refs %v errs %v", refs, errs)
	}

	// Separators (space, comma, semicolon, newline) plus duplicates of none.
	refs, errs := ParseNFRefs(
		"51260437218116000101550010001202441561693345, 51260437218116000101550010001202441561693346;\n51260437218116000101550010001202441561693347")
	if len(errs) != 0 {
		t.Fatalf("valid separators: unexpected errors %v", errs)
	}
	if len(refs) != 3 || refs[0][43:] != "5" || refs[2][43:] != "7" {
		t.Errorf("parsed refs: %v", refs)
	}

	// Invalid entries reject the list naming the bad value.
	_, errs = ParseNFRefs("5126043721811600010155001000120244156169334") // 43 digits
	if len(errs) != 1 || !strings.Contains(errs[0], "44 dígitos") {
		t.Errorf("43-digit entry: got %v", errs)
	}
	_, errs = ParseNFRefs("51260437218116000101550010001202441561693X5") // letter
	if len(errs) != 1 || !strings.Contains(errs[0], "561693X5") {
		t.Errorf("non-digit entry: got %v", errs)
	}
}

func TestValidateNFRefs(t *testing.T) {
	if errs := validateNFRefs([]string{"51260437218116000101550010001202441561693345"}); len(errs) != 0 {
		t.Errorf("valid ref: got %v", errs)
	}
	errs := validateNFRefs([]string{"5126043721811600010155001000120244156169334", "nope"})
	if len(errs) != 2 {
		t.Errorf("invalid refs: got %v", errs)
	}
}

func TestValidateDetachedCSTs(t *testing.T) {
	// nil codes / empty strings pass.
	if errs := validateDetachedCSTs(nil); len(errs) != 0 {
		t.Errorf("nil codes: got %v", errs)
	}
	if errs := validateDetachedCSTs(&entity.InvoiceOverrides{}); len(errs) != 0 {
		t.Errorf("empty codes: got %v", errs)
	}

	// Every whitelisted ICMS CST passes.
	for _, cst := range []string{"00", "20", "40", "41", "50", "51", "90"} {
		errs := validateDetachedCSTs(&entity.InvoiceOverrides{ICMSCST: &cst})
		if len(errs) != 0 {
			t.Errorf("ICMS %s: got %v", cst, errs)
		}
	}
	// Non-whitelisted ICMS rejected (builder falls back to ICMS00 shape —
	// the service must refuse instead).
	errs := validateDetachedCSTs(&entity.InvoiceOverrides{ICMSCST: setPtr("10")})
	if len(errs) != 1 || !strings.Contains(errs[0], "CST ICMS inválido") {
		t.Errorf("ICMS 10: got %v", errs)
	}

	// PIS/COFINS: whitelisted values pass, 03 and 10 reject.
	for _, cst := range []string{"01", "02", "04", "05", "06", "07", "08", "09"} {
		errs := validateDetachedCSTs(&entity.InvoiceOverrides{PISCST: setPtr(cst), COFINSCST: setPtr(cst)})
		if len(errs) != 0 {
			t.Errorf("PIS/COFINS %s: got %v", cst, errs)
		}
	}
	errs = validateDetachedCSTs(&entity.InvoiceOverrides{PISCST: setPtr("03")})
	if len(errs) != 1 || !strings.Contains(errs[0], "CST PIS inválido") {
		t.Errorf("PIS 03: got %v", errs)
	}
	errs = validateDetachedCSTs(&entity.InvoiceOverrides{COFINSCST: setPtr("10")})
	if len(errs) != 1 || !strings.Contains(errs[0], "CST COFINS inválido") {
		t.Errorf("COFINS 10: got %v", errs)
	}
}

func TestMapDetachedEntregaInput(t *testing.T) {
	// nil section → nil (no <entrega> group).
	if got, errs := mapDetachedEntregaInput(nil, nil); got != nil || len(errs) != 0 {
		t.Errorf("nil section: got %v errs %v", got, errs)
	}

	// All-empty section → nil.
	empty := &DetachedEntregaInput{}
	if got, errs := mapDetachedEntregaInput(empty, nil); got != nil || len(errs) != 0 {
		t.Errorf("empty section: got %v errs %v", got, errs)
	}

	// Missing mandatory fields listed.
	partial := &DetachedEntregaInput{Document: "76061198000185"}
	got, errs := mapDetachedEntregaInput(partial, nil)
	if got != nil {
		t.Fatalf("partial section should not map, got %+v", got)
	}
	joined := strings.Join(errs, "; ")
	for _, want := range []string{"nome", "logradouro", "número", "bairro", "cidade"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing-field list lacks %q: %v", want, errs)
		}
	}

	resolver := func(city, state string) (string, error) {
		if city == "NOVA SANTA HELENA" && state == "MT" {
			return "5106190", nil
		}
		return "", nil
	}

	// Complete CNPJ-facing section → Type 1 + IBGE code; IE/CEP optional.
	complete := &DetachedEntregaInput{
		Name:         "TRANSPORTES DEMONSTRATIVOS LTDA",
		Document:     "76061198000185",
		IE:           setPtr("31841627867"),
		Street:       "FAZENDA MODELO DOIS, GLEBA LOTE 292",
		Number:       "S/N",
		Neighborhood: "PERIMETRO URBANO",
		City:         setPtr("NOVA SANTA HELENA"),
		State:        setPtr("MT"),
		CEP:          setPtr("78548000"),
	}
	got, errs = mapDetachedEntregaInput(complete, resolver)
	if len(errs) != 0 {
		t.Fatalf("complete section: unexpected errors %v", errs)
	}
	if got.Type != 1 || got.CNPJ != "76061198000185" {
		t.Errorf("CNPJ mapping: got %+v", got)
	}
	if got.CodigoMun != "5106190" || got.XNome != "TRANSPORTES DEMONSTRATIVOS LTDA" ||
		got.IE != "31841627867" || got.CEP != "78548000" {
		t.Errorf("complete mapping: got %+v", got)
	}

	// CPF (11 digits) → Type 2.
	cpf := &DetachedEntregaInput{
		Name: "Fulano", Document: "52998224725",
		Street: "FAZENDA", Number: "1", Neighborhood: "Centro",
		City: setPtr("NOVA SANTA HELENA"), State: setPtr("MT"),
	}
	got, errs = mapDetachedEntregaInput(cpf, resolver)
	if len(errs) != 0 || got.Type != 2 || got.CPF != "52998224725" {
		t.Errorf("CPF mapping: errs %v got %+v", errs, got)
	}

	// Masked document (only digits counted).
	masked := &DetachedEntregaInput{
		Name: "Fulano", Document: "52.998.224/725",
		Street: "FAZENDA", Number: "1", Neighborhood: "Centro",
		City: setPtr("NOVA SANTA HELENA"), State: setPtr("MT"),
	}
	if _, errs = mapDetachedEntregaInput(masked, resolver); len(errs) != 0 {
		t.Errorf("masked CPF: got %v", errs)
	}

	// Invalid document length rejected.
	badDoc := &DetachedEntregaInput{
		Name: "X", Document: "123", Street: "A", Number: "1", Neighborhood: "B",
		City: setPtr("NOVA SANTA HELENA"), State: setPtr("MT"),
	}
	if got, errs = mapDetachedEntregaInput(badDoc, resolver); got != nil || len(errs) == 0 {
		t.Errorf("bad document: got %v errs %v", got, errs)
	}

	// City resolution failure (service error).
	badResolver := func(city, state string) (string, error) {
		return "", errors.New("db down")
	}
	cityFail := &DetachedEntregaInput{
		Name: "Fulano", Document: "76061198000185",
		Street: "FAZENDA", Number: "1", Neighborhood: "Centro",
		City: setPtr("NOVA SANTA HELENA"), State: setPtr("MT"),
	}
	if got, errs = mapDetachedEntregaInput(cityFail, badResolver); got != nil || len(errs) != 1 {
		t.Errorf("resolver error: got %v errs %v", got, errs)
	}

	// City not found (resolver returns empty).
	notFound := &DetachedEntregaInput{
		Name: "Fulano", Document: "76061198000185",
		Street: "FAZENDA", Number: "1", Neighborhood: "Centro",
		City: setPtr("CIDADE FOFA"), State: setPtr("MT"),
	}
	got, errs = mapDetachedEntregaInput(notFound, resolver)
	if got != nil || len(errs) != 1 || !strings.Contains(errs[0], "CIDADE FOFA") {
		t.Errorf("city not found: got %v errs %v", got, errs)
	}
}

func setPtr(v string) *string {
	return &v
}

// TestApplyRecipientEmail verifies the detached-only E20 dest/email
// propagation shared by both recipient branches: blank/null → untouched
// (builder omits the element), whitespace trimmed.
func TestApplyRecipientEmail(t *testing.T) {
	var recipient entity.RecipientData

	applyRecipientEmail(&recipient, nil)
	if recipient.Email != "" {
		t.Errorf("nil email must leave recipient untouched, got %q", recipient.Email)
	}

	applyRecipientEmail(&recipient, setPtr("   "))
	if recipient.Email != "" {
		t.Errorf("blank email must be ignored, got %q", recipient.Email)
	}

	applyRecipientEmail(&recipient, setPtr("  contato@exemplo.com.br "))
	if recipient.Email != "contato@exemplo.com.br" {
		t.Errorf("email = %q, want trimmed value", recipient.Email)
	}
}

// TestMapInlineRecipientEmail drives the inline recipient mapping (City/State
// nil keeps the municipality lookup offline) and asserts the form's
// recipientEmail lands in RecipientData.Email.
func TestMapInlineRecipientEmail(t *testing.T) {
	svc := NewNFeService()

	input := DetachedRecipient{
		Name:     "TRANSPORTES DEMONSTRATIVOS LTDA",
		Document: "37218116000101",
		Email:    setPtr("  contato744265@exemplo.com.br "),
	}

	recipient := svc.mapInlineRecipientData(input, nil)
	if got, want := recipient.Email, "contato744265@exemplo.com.br"; got != want {
		t.Errorf("inline recipient email = %q, want %q", got, want)
	}

	input.Email = nil
	recipient = svc.mapInlineRecipientData(input, nil)
	if recipient.Email != "" {
		t.Errorf("absent email must stay empty, got %q", recipient.Email)
	}

	// Person-backed branch shape: applyRecipientEmail(&recipient, person.Email)
	// is the same helper verified in TestApplyRecipientEmail — the service
	// wiring is a one-line call (see resolveOrCreateRecipient).
}
