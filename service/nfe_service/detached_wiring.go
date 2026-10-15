package nfe_service

import (
	"fmt"
	"strings"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
)

// detached_wiring.go holds the pure helpers that wire the NF-e builder's
// capability inputs into the detached (avulsa) flow: per-item discounts,
// finNFe derivation, referenced NF-e keys, effective CST validation, and the
// delivery-site (entrega) mapping. Spec:
// docs/sdd/detached_nfe_input_wiring/sdd.md.

// detPag/tPag codes accepted by the detached emission form.
const (
	// TPagDinheiro — Dinheiro.
	TPagDinheiro = "01"
	// TPagSemPagamento — Sem pagamento (default: the previous hardcoded
	// behavior; the builder forces vPag=0.00 for this tPag).
	TPagSemPagamento = "90"
	// TPagOutros — Outros (carries the corpus xPag default "Outros").
	TPagOutros = "99"
)

// resolveDetachedTPag resolves the effective detPag/tPag: an empty input
// keeps the legacy "90" (Sem pagamento) behavior.
func resolveDetachedTPag(input string) string {
	if input == "" {
		return TPagSemPagamento
	}
	return input
}

// ValidDetachedTPag is the form whitelist for tpag; values outside it are
// rejected by the router with a pt-br message.
func ValidDetachedTPag(value string) bool {
	switch value {
	case TPagDinheiro, TPagSemPagamento, TPagOutros:
		return true
	}
	return false
}

// applyRecipientEmail sets the recipient email (E20 dest/email) from an
// optional pointer, trimming surrounding whitespace and ignoring blanks —
// the builder omits the element when empty. Used by both detached recipient
// branches (person-backed and inline form).
func applyRecipientEmail(recipient *entity.RecipientData, email *string) {
	if email == nil {
		return
	}
	if trimmed := strings.TrimSpace(*email); trimmed != "" {
		recipient.Email = trimmed
	}
}

// validateItemDiscounts checks the per-item discounts of a detached input:
// each discount must be non-negative and smaller than the item gross total
// (quantity × unitPrice), so every item keeps a positive net value.
//
// Tax bases (ICMS/PIS/COFINS/IBS) deliberately stay on the gross total —
// the reference emitter corpus behaves this way (G3 devoluções carry
// vBC = vProd with prod/vDesc present); see buildDetachedItems.
func validateItemDiscounts(items []DetachedItemInput) []string {
	var errs []string
	for i, item := range items {
		if item.Discount.IsNegative() {
			errs = append(errs, fmt.Sprintf("Item %d: o desconto não pode ser negativo", i+1))
			continue
		}
		grossTotal := item.Quantity.Mul(item.UnitPrice)
		if item.Discount.Cmp(grossTotal) >= 0 {
			errs = append(errs, fmt.Sprintf("Item %d: o desconto deve ser menor que o valor bruto do item", i+1))
		}
	}
	return errs
}

// resolveDetachedFinNFe derives the emission's finNFe from the item CFOPs:
// "4" (devolução) when every item is a return CFOP, "1" (normal) otherwise.
// A mixed set is rejected with a pt-br message naming both kinds — the same
// homogeneity rule applied to naturezaOp.
func resolveDetachedFinNFe(items []DetachedItemInput) (string, []string) {
	if len(items) == 0 {
		return "1", nil
	}

	first := defaults.FinNFeForCFOP(items[0].CFOP)
	for i, item := range items {
		if got := defaults.FinNFeForCFOP(item.CFOP); got != first {
			if first == "4" {
				return "", []string{fmt.Sprintf(
					"Item %d (CFOP %s) não é uma devolução; emita a NF-e separadamente (CFOPs de devolução: %s…)",
					i+1, item.CFOP, firstFormingCFOPs(items))}
			}
			return "", []string{fmt.Sprintf(
				"Item %d (CFOP %s) é uma devolução; emita a NF-e separadamente (CFOPs de devolução na operação: ...)",
				i+1, item.CFOP)}
		}
	}
	return first, nil
}

// firstFormingCFOPs lists (up to three) input CFOPs of the first finNFe kind
// for a pt-br rejection message.
func firstFormingCFOPs(items []DetachedItemInput) string {
	var cfops []string
	for _, item := range items {
		if defaults.FinNFeForCFOP(item.CFOP) == "4" {
			cfops = append(cfops, item.CFOP)
			if len(cfops) >= 3 {
				break
			}
		}
	}
	return strings.Join(cfops, ", ")
}

// validateNFRefs checks the referenced NF-e keys: each entry must be a
// 44-digit numeric access key; failures name the offending value.
func validateNFRefs(refs []string) []string {
	var errs []string
	for _, ref := range refs {
		if len(ref) != 44 || !allDigits(ref) {
			errs = append(errs, fmt.Sprintf("Chave de NF-e de referência inválida: %q (informe 44 dígitos)", ref))
		}
	}
	return errs
}

// ParseNFRefs reads the nfRefs form textarea: keys are split on whitespace
// and/or commas/semicolons; every entry must be a 44-digit numeric access
// key or the whole input is rejected (errors message each bad value).
func ParseNFRefs(raw string) ([]string, []string) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	splitFn := func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', ',', ';':
			return true
		}
		return false
	}
	fields := strings.FieldsFunc(raw, splitFn)
	var refs, errs []string
	for _, field := range fields {
		if len(field) != 44 || !allDigits(field) {
			errs = append(errs, fmt.Sprintf("Chave de NF-e de referência inválida: %q (informe 44 dígitos)", field))
			continue
		}
		refs = append(refs, field)
	}
	return refs, errs
}

// allDigits reports whether the string is non-empty ASCII digits only.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// validDetachedICMSCSTs mirrors the regime-normal sets the XML builder emits
// (ICMS00/20/40/51/90). CFOP groups the builder cannot express (10, 70, …)
// keep falling back to the ICMS00 shape — hardening here instead means the
// service rejects an unsupported value instead of emitting an ICMS00 with a
// mismatched CST that SEFAZ rejects.
var validDetachedICMSCSTs = map[string]bool{
	"00": true, "20": true, "40": true, "41": true, "50": true, "51": true, "90": true,
}

// validDetachedPISCOFINSCSTs mirrors the builder's PIS/COFINS branches:
// 01/02 → Aliq, 04–09 → NT. CST 03 has no branch, so it is excluded.
var validDetachedPISCOFINSCSTs = map[string]bool{
	"01": true, "02": true, "04": true, "05": true, "06": true, "07": true, "08": true, "09": true,
}

// validateDetachedCSTs checks the resolved ICMS/PIS/COFINS CSTs (user
// override or farm default) against the effective-value whitelists above.
// Empty values are allowed: the resolution defaults (ICMS 00; PIS/COFINS 01)
// are filled by the builder paths.
func validateDetachedCSTs(codes *entity.InvoiceOverrides) []string {
	if codes == nil {
		return nil
	}
	var errs []string
	if codes.ICMSCST != nil && *codes.ICMSCST != "" && !validDetachedICMSCSTs[*codes.ICMSCST] {
		errs = append(errs, fmt.Sprintf(
			"CST ICMS inválido: %s (opções válidas: 00, 20, 40, 41, 50, 51 ou 90)", *codes.ICMSCST))
	}
	if codes.PISCST != nil && *codes.PISCST != "" && !validDetachedPISCOFINSCSTs[*codes.PISCST] {
		errs = append(errs, fmt.Sprintf(
			"CST PIS inválido: %s (opções válidas: 01, 02, 04, 05, 06, 07, 08 ou 09)", *codes.PISCST))
	}
	if codes.COFINSCST != nil && *codes.COFINSCST != "" && !validDetachedPISCOFINSCSTs[*codes.COFINSCST] {
		errs = append(errs, fmt.Sprintf(
			"CST COFINS inválido: %s (opções válidas: 01, 02, 04, 05, 06, 07, 08 ou 09)", *codes.COFINSCST))
	}
	return errs
}

// mapDetachedEntregaInput validates and maps the delivery-address section of
// the detached form into entity.EntregaData. An all-empty section maps to
// nil; any non-empty field makes the mandatory set required. City (name) +
// State are resolved to the IBGE municipality code through resolveCity —
// production passes the nfe model's GetMunicipio; tests pass a stub.
//
// The document login type is inferred by digit length (13+ → CNPJ,
// otherwise CPF), the same rule as the detached recipient.
func mapDetachedEntregaInput(input *DetachedEntregaInput, resolveCity func(city, state string) (string, error)) (*entity.EntregaData, []string) {
	if input == nil {
		return nil, nil
	}

	if entregaAllEmpty(input) {
		return nil, nil
	}

	var errs []string
	if input.Name == "" {
		errs = append(errs, "Informe o nome / razão social do local de entrega")
	}
	if input.Document == "" {
		errs = append(errs, "Informe o CPF/CNPJ do local de entrega")
	}
	if input.Street == "" {
		errs = append(errs, "Informe o logradouro do local de entrega")
	}
	if input.Number == "" {
		errs = append(errs, "Informe o número do local de entrega (ou \"S/N\")")
	}
	if input.Neighborhood == "" {
		errs = append(errs, "Informe o bairro do local de entrega")
	}
	if input.City == nil || *input.City == "" || input.State == nil || *input.State == "" {
		errs = append(errs, "Informe cidade e UF do local de entrega")
	}
	if len(errs) > 0 {
		return nil, errs
	}

	// Document type: mirror the detached recipient rule (13+ chars → CNPJ).
	isCNPJ := detachedDocumentType(input.Document) == 2
	docDigits := digitsOnly(input.Document)
	if len(docDigits) != 11 && len(docDigits) != 14 {
		errs = append(errs, "CPF/CNPJ do local de entrega inválido: informe 11 ou 14 dígitos")
	}

	// Municipality code resolution (name → IBGE code), same source as the
	// inline recipient.
	var codigoMun string
	if resolveCity != nil {
		code, resolveErr := resolveCity(*input.City, *input.State)
		if resolveErr != nil {
			errs = append(errs, fmt.Sprintf("Falha ao resolver a cidade de entrega %s/%s", *input.City, *input.State))
		} else if code == "" {
			errs = append(errs, fmt.Sprintf("Cidade de entrega não encontrada: %s/%s (confira o nome e a UF)", *input.City, *input.State))
		} else {
			codigoMun = code
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}

	out := &entity.EntregaData{
		XNome:      input.Name,
		Logradouro: input.Street,
		Numero:     input.Number,
		Bairro:     input.Neighborhood,
		CodigoMun:  codigoMun,
		Municipio:  *input.City,
		UF:         *input.State,
	}
	if input.IE != nil && *input.IE != "" {
		out.IE = *input.IE
	}
	if input.CEP != nil && *input.CEP != "" {
		out.CEP = *input.CEP
	}
	if isCNPJ {
		out.Type = 1
		out.CNPJ = input.Document
	} else {
		out.Type = 2
		out.CPF = input.Document
	}
	return out, nil
}

// entregaAllEmpty reports whether the delivery section was skipped entirely.
func entregaAllEmpty(input *DetachedEntregaInput) bool {
	return input.Name == "" && input.Document == "" &&
		(input.IE == nil || *input.IE == "") &&
		input.Street == "" && input.Number == "" && input.Neighborhood == "" &&
		(input.City == nil || *input.City == "") &&
		(input.State == nil || *input.State == "") &&
		(input.CEP == nil || *input.CEP == "")
}

// digitsOnly strips non-digit characters — documents may carry mask chars.
func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
