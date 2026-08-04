package xml

import (
	"strings"

	"github.com/beevik/etree"
)

// The NF-e schema types free-text fields as TString, whose pattern
// "[!-ÿ]{1}[ -ÿ]{0,}[!-ÿ]{1}|[!-ÿ]{1}" only allows printable Latin-1
// characters (U+0020–U+00FF) with no leading/trailing spaces. Control
// characters (newlines, tabs, carriage returns) and characters above U+00FF
// (e.g., emoji) cause a schema rejection at SEFAZ (cvc-type.3.1.3).

// SanitizeSchemaString normalizes a free-text value so it conforms to the
// schema TString pattern: line breaks become "; ", tabs become spaces, any
// character outside U+0020–U+00FF is dropped, and leading/trailing spaces are
// trimmed. Empty input yields empty output.
func SanitizeSchemaString(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	parts := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(strings.Map(schemaRune, ln))
		if ln != "" {
			parts = append(parts, ln)
		}
	}
	return strings.Join(parts, "; ")
}

// schemaRune keeps only characters allowed by the TString pattern,
// converting tabs to spaces.
func schemaRune(r rune) rune {
	if r == '\t' {
		return ' '
	}
	if r < 0x20 || r > 0xFF {
		return -1
	}
	return r
}

// setSchemaText creates a child element with text content sanitized for the
// schema TString pattern. Use it for all free-text fields (xNome, xLgr,
// infCpl, xJust, etc.). Coded/pattern fields (cUF, CNPJ, CEP, dates, numeric
// values) should keep using CreateElement(...).SetText(...) directly.
func setSchemaText(parent *etree.Element, tag, value string) {
	parent.CreateElement(tag).SetText(SanitizeSchemaString(value))
}
