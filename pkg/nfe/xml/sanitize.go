package xml

import (
	"strings"

	"github.com/beevik/etree"
)

// The NF-e schema types free-text fields as TString, whose pattern
// "[!-ÿ]{1}[ -ÿ]{0,}[!-ÿ]{1}|[!-ÿ]{1}" allows printable Latin-1 characters
// (U+0020–U+00FF) with no leading/trailing spaces. Control characters
// (newlines, tabs, carriage returns) and characters above U+00FF (e.g., emoji)
// cause a schema rejection at SEFAZ (cvc-type.3.1.3).
//
// SEFAZ-MT additionally rejects accented Latin-1 code points in the área de
// dados with cStat 402 ("XML da área de dados com codificação diferente de
// UTF-8"). To stay compatible across states, SanitizeSchemaString folds
// accented characters to their ASCII base letter, so the emitted content is a
// strict ASCII subset of UTF-8.

// SanitizeSchemaString normalizes a free-text value so it is safe to emit to
// SEFAZ: line breaks become "; ", tabs become spaces, accented Latin-1
// characters are folded to their ASCII base letter (ç→c, ã→a, ...), any
// character that is still outside U+0020–U+007E is dropped, and leading/
// trailing spaces are trimmed. Empty input yields empty output.
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

// schemaRune folds an accented Latin-1 rune to its ASCII base and keeps only
// printable ASCII, converting tabs to spaces.
func schemaRune(r rune) rune {
	if r == '\t' {
		return ' '
	}
	if folded, ok := accentFold[r]; ok {
		return folded
	}
	if r < 0x20 || r > 0x7E {
		return -1
	}
	return r
}

// accentFold maps Latin-1 accented letters to their ASCII base letter. It
// mirrors the "CaracteresRemoverAcentos" set used across Brazilian NF-e
// emitters. Combining marks (decomposed input) fall outside U+0020–U+007E and
// are dropped by schemaRune, which folds that input too.
var accentFold = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o',
	'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'ç': 'c', 'Ç': 'C',
	'ñ': 'n', 'Ñ': 'N',
	'ý': 'y', 'ÿ': 'y', 'Ý': 'Y',
	'º': 'o', 'ª': 'a',
}

// setSchemaText creates a child element with text content sanitized for the
// schema TString pattern. Use it for all free-text fields (xNome, xLgr,
// infCpl, xJust, etc.). Coded/pattern fields (cUF, CNPJ, CEP, dates, numeric
// values) should keep using CreateElement(...).SetText(...) directly.
func setSchemaText(parent *etree.Element, tag, value string) {
	parent.CreateElement(tag).SetText(SanitizeSchemaString(value))
}
