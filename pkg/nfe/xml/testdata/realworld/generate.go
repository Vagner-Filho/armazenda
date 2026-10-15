//go:build ignore

// generate.go — deterministically anonymizes the real-world NF-e corpus in
// docs/research_nfe_use_case/ into offline test fixtures.
//
// Run from the repo root:
//
//	go run pkg/nfe/xml/testdata/realworld/generate.go docs/research_nfe_use_case pkg/nfe/xml/testdata/realworld
//
// What it does (deterministic; same input -> byte-identical output):
//  1. Classifies every file into a use-case group (G1..G7, see
//     docs/research_nfe_use_case/USE_CASES.md).
//  2. Collects every person/company token across the corpus (names, CPF/CNPJ,
//     IE, emails, phones, addresses, driver names, plates) and maps each to a
//     stable synthetic identity with valid check digits.
//  3. Rewrites each XML: entities, infCpl free text (driver names, plates, CND
//     authentication codes, CND/contract numbers, punctuated CNPJs), rebuilds
//     the 44-digit access key with the synthetic emit document + recomputed
//     cDV (syncing infNFe Id and protNFe chNFe), and replaces the protocol
//     number and cryptographic material with synthetic values.
//  4. Business data is preserved unchanged: natOp, CFOP, ICMS/PIS/COFINS/IBS,
//     quantities, values, dates, modFrete, volumes, infAdProd, municipalities.
//  5. Verifies every output file (parses; key consistency) and that no
//     collected real token appears in any output; writes manifest.json +
//     README.md.
//
// Note: the fixtures' digital signatures are structurally intact but
// cryptographically invalid by construction (synthetic digest/cert bytes).
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// ---------------------------------------------------------------------------
// deterministic randomness

func h64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// nextByte advances an xorshift64 state and returns the low byte.
func nextByte(state *uint64) byte {
	*state ^= *state << 13
	*state ^= *state >> 7
	*state ^= *state << 17
	return byte(*state)
}

// digits returns n deterministic decimal digits from a seed.
func digits(seed uint64, n int) string {
	var b strings.Builder
	state := seed | 1
	for b.Len() < n {
		b.WriteString(strconv.FormatUint(uint64(nextByte(&state))%10, 10))
	}
	return b.String()
}

func digitsFromString(s string, n int) string { return digits(h64(s), n) }

// syntheticBase64 returns base64 of n deterministic bytes — shaped like the
// real digest/signature/certificate fields.
func syntheticBase64(seed string, n int) string {
	state := h64(seed) | 1
	b := make([]byte, n)
	for i := range b {
		b[i] = nextByte(&state)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// syntheticAlnum returns a deterministic uppercase alnum string of length n.
func syntheticAlnum(s string, n int) string {
	alphabet := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	state := h64(s) | 1
	out := make([]byte, 0, n)
	for len(out) < n {
		out = append(out, alphabet[uint64(nextByte(&state))%uint64(len(alphabet))])
	}
	return string(out)
}

// ---------------------------------------------------------------------------
// check-digit math

func intDigits(s string) []int {
	out := make([]int, 0, len(s))
	for _, c := range s {
		if c >= '0' && c <= '9' {
			out = append(out, int(c-'0'))
		}
	}
	return out
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// cnpjDV computes the two CNPJ check digits for a 13-digit prefix (base + d1).
func cnpjDV(vals []int) int {
	weights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	if len(vals) == 13 {
		weights = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	}
	sum := 0
	for i, v := range vals {
		sum += v * weights[i]
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

func syntheticCNPJ(real string) string {
	root := digitsFromString("CNPJ:"+digitsOnly(real), 8)
	base := root + "0001"
	baseDigits := intDigits(base)
	d1 := cnpjDV(baseDigits)
	d2 := cnpjDV(append(append([]int{}, baseDigits...), d1))
	return fmt.Sprintf("%s%d%d", base, d1, d2)
}

func cpfDV(vals []int) int {
	w := len(vals) + 1
	sum := 0
	for _, v := range vals {
		sum += v * w
		w--
	}
	r := (sum * 10) % 11
	if r == 10 {
		r = 0
	}
	return r
}

func syntheticCPF(real string) string {
	body := digitsFromString("CPF:"+digitsOnly(real), 9)
	bodyDigits := intDigits(body)
	d1 := cpfDV(bodyDigits)
	d2 := cpfDV(append(append([]int{}, bodyDigits...), d1))
	return fmt.Sprintf("%s%d%d", body, d1, d2)
}

// accessKeyDV: mod 11 with weights 2..9 cycling from the right over the first
// 43 digits of the NF-e access key.
func accessKeyDV(key43 string) int {
	sum := 0
	weight := 2
	for i := len(key43) - 1; i >= 0; i-- {
		sum += int(key43[i]-'0') * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

// rebuildAccessKey replaces the emitter document inside a 44-digit access key
// (positions 7..20, zero-padded for CPF emitters) and recomputes the cDV.
func rebuildAccessKey(originalKey, syntheticEmitDoc string) string {
	if len(originalKey) != 44 {
		return originalKey
	}
	doc14 := digitsOnly(syntheticEmitDoc)
	if len(doc14) > 14 || doc14 == "" {
		return originalKey
	}
	doc14 = strings.Repeat("0", 14-len(doc14)) + doc14
	prefix := originalKey[:6] + doc14 + originalKey[20:43]
	return prefix + strconv.Itoa(accessKeyDV(prefix))
}

// ---------------------------------------------------------------------------
// synthetic identity pools (ASCII-only, matching SanitizeSchemaString rules)

var companyPool = []string{
	"AGROINDUSTRIA MODELO SINTETICA LTDA", "COMERCIAL EXEMPLO SUL LTDA",
	"IMPORTADORA FICTICIA NORTE S/A", "INDUSTRIA DEMONSTRATIVA CENTRO LTDA",
	"DISTRIBUIDORA ILUSTRATIVA S/A", "AGROPECUARIA EXEMPLO OESTE LTDA",
	"TRANSPORTES DEMONSTRATIVOS LTDA", "TRADER EXEMPLO GRANEIS S/A",
	"HIDROVIARIA FICTICIA LTDA", "COOPERATIVA DEMONSTRATIVA S/A",
}

var personPool = []string{
	"JOAO RICARDO ALMEIDA", "MARIA CLARA OLIVEIRA", "PEDRO HENRIQUE SANTOS",
	"ANA BEATRIZ PEREIRA", "LUCAS MARTINS ROCHA", "JULIANA COSTA LIMA",
	"CARLOS EDUARDO FERREIRA", "RENATA GOMES BARROS",
}

var driverPool = []string{
	"CARTEIRO SINTETICO UM", "CARTEIRO SINTETICO DOIS", "CARTEIRO SINTETICO TRES",
	"CARTEIRO SINTETICO QUATRO", "CARTEIRO SINTETICO CINCO",
}

var farmNamePool = []string{
	"FAZENDA SINTETICA UM", "FAZENDA MODELO DOIS", "GLEBA EXEMPLO TRES",
	"FAZENDA DEMONSTRATIVA QUATRO", "FAZENDA ILUSTRATIVA CINCO",
}

var plateLetters = "ABCDEFGHJKLMNPQRSTUVWXYZ"

func nameFor(pool []string, salt, real string) string {
	return pool[int(h64(salt+"|"+digitsOnly(real))%uint64(len(pool)))]
}

func emailFor(real string) string {
	return fmt.Sprintf("contato%d%d@exemplo.com.br", h64(real)%1000, 100+h64("E|"+real)%900)
}

func phoneFor(real string) string {
	return "65" + digitsFromString("FONE|"+real, 9)
}

func ieFor(real string) string {
	n := len(digitsOnly(real))
	if n < 4 {
		n = 8
	}
	return digitsFromString("IE|"+digitsOnly(real), n)
}

func addressFor(real string) string {
	return farmNamePool[int(h64("ADDR|"+real)%uint64(len(farmNamePool)))] +
		", GLEBA LOTE " + digitsFromString("ADDRN|"+real, 3)
}

// syntheticPlate: Mercosul format AAA9A99
func syntheticPlate(s string) string {
	state := h64("PLACA|" + s)
	pick := func(i int) byte { return plateLetters[(state>>uint(i*5))%uint64(len(plateLetters))] }
	l := []byte{pick(1), pick(2), pick(3)}
	d := digitsFromString("PLACAD|"+s, 4)
	return string(l) + d[:1] + string(l[0]) + d[1:3]
}

// ---------------------------------------------------------------------------
// tokens

type tokenType int

const (
	tokDoc tokenType = iota
	tokPersonName
	tokCompanyName
	tokEmail
	tokPhone
	tokIE
	tokAddress
	tokDriver
	tokAuthCode
	tokPlate
)

type token struct {
	kind tokenType
	real string
}

func syntheticValue(t token) string {
	switch t.kind {
	case tokDoc:
		n := len(digitsOnly(t.real))
		if n == 11 {
			return syntheticCPF(t.real)
		}
		if n == 14 {
			return syntheticCNPJ(t.real)
		}
		// length-preserving synthetic digits for numeric codes (CND numbers,
		// contract numbers)
		return digitsFromString("CODE|"+digitsOnly(t.real), n)
	case tokPersonName:
		return nameFor(personPool, "NAME", t.real)
	case tokCompanyName:
		return nameFor(companyPool, "NAME", t.real)
	case tokDriver:
		return nameFor(driverPool, "DRIVER", t.real)
	case tokEmail:
		return emailFor(t.real)
	case tokPhone:
		return phoneFor(t.real)
	case tokIE:
		return ieFor(t.real)
	case tokAddress:
		return addressFor(t.real)
	case tokAuthCode:
		return syntheticAlnum("AUTHCODE|"+t.real, len(t.real))
	case tokPlate:
		return syntheticPlate(t.real)
	}
	return t.real
}

// ---------------------------------------------------------------------------
// corpus: collection + per-file rewrite

type corpus struct {
	tokens map[string]token
	emit12 string // real emit document digits (11 for CPF, 14 for CNPJ)
}

func newCorpus() *corpus {
	return &corpus{tokens: map[string]token{}}
}

func (c *corpus) add(kind tokenType, real string) {
	real = strings.TrimSpace(real)
	if real == "" || len([]rune(real)) < 4 {
		return // nothing identify-worthy is this short (avoids tag-name collisions)
	}
	if _, ok := c.tokens[real]; !ok {
		c.tokens[real] = token{kind: kind, real: real}
	}
}

func (c *corpus) syntheticFor(real string) string {
	if t, ok := c.tokens[real]; ok {
		return syntheticValue(t)
	}
	return real
}

func (c *corpus) syntheticEmitDoc() string {
	if c.emit12 == "" {
		return ""
	}
	if len(c.emit12) == 11 {
		return syntheticCPF(c.emit12)
	}
	return syntheticCNPJ(c.emit12)
}

// ---------------------------------------------------------------------------
// regexes for free text (Go RE2: no lookaheads)

var (
	driverRegex         = regexp.MustCompile(`MOTORISTA:\s*([A-Z][A-Z ]{2,60})`)
	punctuatedCNPJRegex = regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}|\d{8}/\d{4}-\d{2}`)
	authCodeRegex       = regexp.MustCompile(`AUTENTICACAO ([A-Z0-9]{10,25})`)
	cndNumberRegex      = regexp.MustCompile(`(NUMERO|N) (\d{6,15})`)
	contractNumberRegex = regexp.MustCompile(`(?:CONTRATO)[A-Z ]*(\d{4,10})`)
	ctrNumberRegex      = regexp.MustCompile(`CTR[ ](\d{4,10})`)
	dashNumberRegex     = regexp.MustCompile(`(\d{6,12})-(\d{4,8})`)
	barePlateRegex      = regexp.MustCompile(`\b([A-Z]{3}[0-9][A-Z0-9][0-9]{2}|[A-Z]{3}-[0-9]{4})\b`)
	docRefRegex         = regexp.MustCompile(`[0-9]{44}`)
)

// ---------------------------------------------------------------------------
// helpers

func elText(parent *etree.Element, tag string) string {
	if parent == nil {
		return ""
	}
	e := parent.SelectElement(tag)
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.Text())
}

// infNFeOf returns the <infNFe> element for either an <nfeProc> root or a
// bare <NFe> document.
func infNFeOf(root *etree.Element) *etree.Element {
	if root == nil {
		return nil
	}
	if root.Tag == "NFe" {
		return root.SelectElement("infNFe")
	}
	if nfe := root.SelectElement("NFe"); nfe != nil {
		return nfe.SelectElement("infNFe")
	}
	return nil
}

func findTag(parent *etree.Element, tag string) *etree.Element {
	if parent == nil {
		return nil
	}
	if parent.Tag == tag {
		return parent
	}
	for _, ch := range parent.ChildElements() {
		if e := findTag(ch, tag); e != nil {
			return e
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// classification

type natOpInfo struct {
	natOp, cfop, finNFe string
}

func readFinance(root *etree.Element) natOpInfo {
	f := natOpInfo{}
	if infNFe := infNFeOf(root); infNFe != nil {
		if i := infNFe.SelectElement("ide"); i != nil {
			f.natOp = elText(i, "natOp")
			f.finNFe = elText(i, "finNFe")
		}
		if det := infNFe.SelectElement("det"); det != nil {
			if prod := det.SelectElement("prod"); prod != nil {
				f.cfop = elText(prod, "CFOP")
			}
		}
	}
	return f
}

func classify(cfop, finNFe string) string {
	switch {
	case cfop == "5101" && finNFe == "1":
		return "G1"
	case cfop == "5118":
		return "G2"
	case cfop == "5556":
		return "G3"
	case cfop == "5913":
		return "G4"
	case cfop == "5915":
		return "G5"
	case cfop == "5551":
		return "G6"
	case cfop == "5949":
		return "G7"
	default:
		return "G?"
	}
}

var representativeNNF = map[int]bool{946: true, 972: true, 1060: true, 1064: true, 1067: true, 1098: true, 1104: true}

// ---------------------------------------------------------------------------
// collection pass

func collect(files []string) (*corpus, map[string]*etree.Document, map[string]natOpInfo, error) {
	c := newCorpus()
	docs := map[string]*etree.Document{}
	finances := map[string]natOpInfo{}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(raw); err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", path, err)
		}
		root := doc.Root()
		ownKey := strings.TrimSuffix(filepath.Base(path), ".xml")
		key := ownKey

		// connected-document references (refNFe / infCpl CHAVE): any 44-digit
		// run that is not this file's own key embeds a third-party document
		for _, m := range docRefRegex.FindAllString(string(raw), -1) {
			if m != ownKey {
				c.add(tokDoc, m[6:20])
			}
		}
		infNFe := infNFeOf(root)

		if emit := infNFe.SelectElement("emit"); emit != nil {
			docTxt := elText(emit, "CNPJ")
			if docTxt == "" {
				docTxt = elText(emit, "CPF")
			}
			c.add(tokDoc, docTxt)
			if c.emit12 == "" {
				c.emit12 = digitsOnly(docTxt)
			}
			c.add(tokPersonName, elText(emit, "xNome"))
			c.add(tokPersonName, elText(emit, "xFant"))
			c.add(tokIE, elText(emit, "IE"))
			if ender := emit.SelectElement("enderEmit"); ender != nil {
				c.add(tokAddress, elText(ender, "xLgr"))
				c.add(tokPhone, elText(ender, "fone"))
			}
		}
		if dest := infNFe.SelectElement("dest"); dest != nil {
			c.add(tokDoc, elText(dest, "CNPJ"))
			c.add(tokDoc, elText(dest, "CPF"))
			c.add(tokCompanyName, elText(dest, "xNome"))
			c.add(tokIE, elText(dest, "IE"))
			c.add(tokEmail, elText(dest, "email"))
			if ender := dest.SelectElement("enderDest"); ender != nil {
				c.add(tokAddress, elText(ender, "xLgr"))
			}
		}
		if transp := infNFe.SelectElement("transp"); transp != nil {
			if tra := transp.SelectElement("transporta"); tra != nil {
				c.add(tokDoc, elText(tra, "CNPJ"))
				c.add(tokDoc, elText(tra, "CPF"))
				c.add(tokCompanyName, elText(tra, "xNome"))
				c.add(tokIE, elText(tra, "IE"))
				c.add(tokAddress, elText(tra, "xEnder"))
			}
			if veic := transp.SelectElement("veicTransp"); veic != nil {
				c.add(tokPlate, elText(veic, "placa"))
			}
		}
		if resp := infNFe.SelectElement("infRespTec"); resp != nil {
			c.add(tokDoc, elText(resp, "CNPJ"))
			c.add(tokCompanyName, elText(resp, "xContato"))
			c.add(tokEmail, elText(resp, "email"))
			c.add(tokPhone, elText(resp, "fone"))
		}
		// entrega/retirada delivery site groups (flat shape)
		for _, tag := range []string{"entrega", "retirada"} {
			if grp := infNFe.SelectElement(tag); grp != nil {
				c.add(tokDoc, elText(grp, "CNPJ"))
				c.add(tokDoc, elText(grp, "CPF"))
				c.add(tokCompanyName, elText(grp, "xNome"))
				c.add(tokIE, elText(grp, "IE"))
				c.add(tokAddress, elText(grp, "xLgr"))
			}
		}

		if infAdic := infNFe.SelectElement("infAdic"); infAdic != nil {
			if infCpl := infAdic.SelectElement("infCpl"); infCpl != nil {
				text := infCpl.Text()
				for _, m := range driverRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokDriver, strings.TrimSpace(m[1]))
				}
				for _, m := range punctuatedCNPJRegex.FindAllString(text, -1) {
					if nd := digitsOnly(m); len(nd) == 14 {
						c.add(tokDoc, nd)
					}
				}
				for _, m := range authCodeRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokAuthCode, m[1])
				}
				for _, m := range cndNumberRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokDoc, m[2])
				}
				for _, m := range contractNumberRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokDoc, m[1])
				}
				for _, m := range ctrNumberRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokDoc, m[1])
				}
				for _, m := range dashNumberRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokDoc, m[1])
					c.add(tokDoc, m[2])
				}
				for _, m := range barePlateRegex.FindAllStringSubmatch(text, -1) {
					c.add(tokPlate, m[1])
				}
			}
		}

		fin := readFinance(root)
		finances[key] = fin
		docs[key] = doc
	}
	return c, docs, finances, nil
}

// ---------------------------------------------------------------------------
// per-file rewrite

type manifestEntry struct {
	File           string `json:"file"`
	Group          string `json:"group"`
	OriginalFolder string `json:"originalFolder"`
	NNF            int    `json:"nnf"`
	NatOp          string `json:"natOp"`
	CFOP           string `json:"cfop"`
	FinNFe         string `json:"finNFe"`
	ICMSTag        string `json:"icmsTag"`
	ICMSCST        string `json:"icmsCst"`
	PISCST         string `json:"pisCst"`
	COFINSCST      string `json:"cofinsCst"`
	TPag           string `json:"tPag"`
	ModFrete       string `json:"modFrete"`
	DestType       string `json:"destType"`
	Discount       bool   `json:"discount"`
	InfAdProd      bool   `json:"infAdProd"`
	Representative bool   `json:"representative"`
}

func replaceIdentityDoc(parent *etree.Element, tag string) {
	e := parent.SelectElement(tag)
	if e == nil {
		return
	}
	txt := strings.TrimSpace(e.Text())
	if txt == "" {
		return
	}
	switch len(digitsOnly(txt)) {
	case 11:
		e.SetText(syntheticCPF(txt))
	case 14:
		e.SetText(syntheticCNPJ(txt))
	}
}

func replaceNamed(parent *etree.Element, tag string, kind tokenType) {
	e := parent.SelectElement(tag)
	if e == nil {
		return
	}
	txt := strings.TrimSpace(e.Text())
	if txt == "" {
		return
	}
	e.SetText(syntheticValue(token{kind: kind, real: txt}))
}

func scrubInfCpl(c *corpus, text string) string {
	text = driverRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := driverRegex.FindStringSubmatch(m)
		return "MOTORISTA: " + c.syntheticFor(strings.TrimSpace(g[1]))
	})
	text = punctuatedCNPJRegex.ReplaceAllStringFunc(text, func(m string) string {
		nd := digitsOnly(m)
		if len(nd) != 14 {
			return m
		}
		syn := syntheticCNPJ(nd)
		return syn[:8] + "/" + syn[8:12] + "-" + syn[12:14]
	})
	text = authCodeRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := authCodeRegex.FindStringSubmatch(m)
		return "AUTENTICACAO " + syntheticAlnum("AUTHCODE|"+g[1], len(g[1]))
	})
	text = cndNumberRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := cndNumberRegex.FindStringSubmatch(m)
		return g[1] + " " + digitsFromString("CODE|"+g[2], len(g[2]))
	})
	text = contractNumberRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := contractNumberRegex.FindStringSubmatch(m)
		return "CONTRATO " + digitsFromString("CODE|"+g[1], len(g[1]))
	})
	text = ctrNumberRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := ctrNumberRegex.FindStringSubmatch(m)
		return "CTR " + digitsFromString("CODE|"+g[1], len(g[1]))
	})
	text = dashNumberRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := dashNumberRegex.FindStringSubmatch(m)
		return digitsFromString("CODE|"+g[1], len(g[1])) + "-" + digitsFromString("CODE|"+g[2], len(g[2]))
	})
	text = barePlateRegex.ReplaceAllStringFunc(text, func(m string) string {
		g := barePlateRegex.FindStringSubmatch(m)
		return syntheticPlate(g[1])
	})
	// farm-name fragments of structured addresses that recur inside infCpl
	text = replaceAddressFragments(c, text)
	return text
}

// replaceAddressFragments replaces each comma-separated segment (>= 8 chars)
// of any known address token wherever it appears in free text.
func replaceAddressFragments(c *corpus, text string) string {
	type repl struct{ old, new string }
	var reps []repl
	for real, t := range c.tokens {
		if t.kind != tokAddress {
			continue
		}
		for _, seg := range strings.Split(real, ",") {
			seg = strings.TrimSpace(seg)
			if len(seg) >= 8 {
				reps = append(reps, repl{old: seg, new: syntheticValue(token{kind: tokAddress, real: seg})})
			}
		}
	}
	sort.Slice(reps, func(i, j int) bool { return len(reps[i].old) > len(reps[j].old) })
	for _, r := range reps {
		text = strings.ReplaceAll(text, r.old, r.new)
	}
	return text
}

func transform(doc *etree.Document, c *corpus) (string, error) {
	root := doc.Root()
	infNFe := infNFeOf(root)
	ide := infNFe.SelectElement("ide")
	oldKey := strings.TrimPrefix(infNFe.SelectAttrValue("Id", ""), "NFe")

	if emit := infNFe.SelectElement("emit"); emit != nil {
		replaceIdentityDoc(emit, "CNPJ")
		replaceIdentityDoc(emit, "CPF")
		replaceNamed(emit, "xNome", tokPersonName)
		replaceNamed(emit, "xFant", tokPersonName)
		replaceNamed(emit, "IE", tokIE)
		if ender := emit.SelectElement("enderEmit"); ender != nil {
			replaceNamed(ender, "xLgr", tokAddress)
			replaceNamed(ender, "fone", tokPhone)
		}
	}
	if dest := infNFe.SelectElement("dest"); dest != nil {
		replaceIdentityDoc(dest, "CNPJ")
		replaceIdentityDoc(dest, "CPF")
		replaceNamed(dest, "xNome", tokCompanyName)
		replaceNamed(dest, "IE", tokIE)
		replaceNamed(dest, "email", tokEmail)
		if ender := dest.SelectElement("enderDest"); ender != nil {
			replaceNamed(ender, "xLgr", tokAddress)
		}
	}
	if transp := infNFe.SelectElement("transp"); transp != nil {
		if tra := transp.SelectElement("transporta"); tra != nil {
			replaceIdentityDoc(tra, "CNPJ")
			replaceIdentityDoc(tra, "CPF")
			replaceNamed(tra, "xNome", tokCompanyName)
			replaceNamed(tra, "IE", tokIE)
			replaceNamed(tra, "xEnder", tokAddress)
		}
		if veic := transp.SelectElement("veicTransp"); veic != nil {
			replaceNamed(veic, "placa", tokPlate)
		}
	}
	if resp := infNFe.SelectElement("infRespTec"); resp != nil {
		replaceNamed(resp, "CNPJ", tokDoc)
		replaceNamed(resp, "xContato", tokCompanyName)
		replaceNamed(resp, "email", tokEmail)
		replaceNamed(resp, "fone", tokPhone)
	}
	// entrega/retirada delivery site groups (flat shape: no enderDest wrapper)
	for _, tag := range []string{"entrega", "retirada"} {
		if grp := infNFe.SelectElement(tag); grp != nil {
			replaceIdentityDoc(grp, "CNPJ")
			replaceIdentityDoc(grp, "CPF")
			replaceNamed(grp, "xNome", tokCompanyName)
			replaceNamed(grp, "IE", tokIE)
			replaceNamed(grp, "xLgr", tokAddress)
		}
	}

	if infAdic := infNFe.SelectElement("infAdic"); infAdic != nil {
		if infCpl := infAdic.SelectElement("infCpl"); infCpl != nil {
			infCpl.SetText(scrubInfCpl(c, infCpl.Text()))
		}
	}

	// access-key rebuild + protocol sync
	newKey := rebuildAccessKey(oldKey, c.syntheticEmitDoc())
	if idAttr := infNFe.SelectAttr("Id"); idAttr != nil {
		idAttr.Value = "NFe" + newKey
	}
	if cDV := ide.SelectElement("cDV"); cDV != nil {
		cDV.SetText(strconv.Itoa(accessKeyDV(newKey[:43])))
	}
	if prot := root.SelectElement("protNFe"); prot != nil {
		if ip := prot.SelectElement("infProt"); ip != nil {
			if ch := ip.SelectElement("chNFe"); ch != nil {
				ch.SetText(newKey)
			}
			if np := ip.SelectElement("nProt"); np != nil {
				np.SetText("15" + digitsFromString("NPROT|"+oldKey, 13))
			}
		}
	}

	// cryptographic material: structurally intact, cryptographically void
	for _, dv := range doc.FindElements("//DigestValue") {
		dv.SetText(syntheticBase64(oldKey+"DIGEST", 20))
	}
	for _, dv := range doc.FindElements("//digVal") {
		dv.SetText(syntheticBase64(oldKey+"DIGEST", 20))
	}
	for _, sv := range doc.FindElements("//SignatureValue") {
		sv.SetText(syntheticBase64(oldKey+"SIG", 256))
	}
	for _, cert := range doc.FindElements("//X509Certificate") {
		cert.SetText(syntheticBase64(oldKey+"CERT", 1725))
	}

	doc.WriteSettings.CanonicalEndTags = true
	out, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	// Any remnant of the old key (e.g. ds:Reference URI="#NFe<key>") must be
	// replaced as well — it embeds the real emitter document.
	out = strings.ReplaceAll(out, oldKey, newKey)
	// Connected-document keys (refNFe, infCpl CHAVE): embed third-party
	// documents' CNPJs — rebuild them with synthetic documents and a valid DV.
	// The file's own (already synthetic) key is excluded.
	ownNew := newKey
	out = docRefRegex.ReplaceAllStringFunc(out, func(m string) string {
		if m == ownNew {
			return m
		}
		syn := syntheticCNPJ(m[6:20])
		prefix := m[:6] + syn + m[20:43]
		return prefix + strconv.Itoa(accessKeyDV(prefix))
	})
	return out, nil
}

// ---------------------------------------------------------------------------
// verification

func verify(path, wantKey string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return err
	}
	root := doc.Root()
	if root == nil || root.Tag != "nfeProc" {
		return fmt.Errorf("root tag unexpected")
	}
	infNFe := infNFeOf(root)
	id := strings.TrimPrefix(infNFe.SelectAttrValue("Id", ""), "NFe")
	if id != wantKey {
		return fmt.Errorf("infNFe Id %s != key %s", id, wantKey)
	}
	prot := root.SelectElement("protNFe")
	ch := elText(prot.SelectElement("infProt"), "chNFe")
	if ch != wantKey {
		return fmt.Errorf("protNFe chNFe %s != key %s", ch, wantKey)
	}
	if len(wantKey) != 44 {
		return fmt.Errorf("key length %d", len(wantKey))
	}
	return nil
}

func loadOutputs(outDir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	paths, err := filepath.Glob(filepath.Join(outDir, "*.xml"))
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		out[filepath.Base(p)] = b
	}
	return out, nil
}

func checkLeaks(outputs map[string][]byte, c *corpus) []string {
	// Each output legitimately contains its own (synthetic) access key; remove
	// it first so digit-run collisions with short tokens are not false
	// positives. Any remnant with the OLD key (real document digits) is then
	// detectable — e.g. a missed ds:Reference URI.
	var leaks []string
	for real, t := range c.tokens {
		if syntheticValue(t) == real {
			continue
		}
		for fname, raw := range outputs {
			content := strings.ReplaceAll(string(raw), strings.TrimSuffix(fname, ".xml"), "")
			idx := strings.Index(content, real)
			if idx >= 0 {
				start, end := idx-30, idx+len(real)+30
				if start < 0 {
					start = 0
				}
				if end > len(content) {
					end = len(content)
				}
				leaks = append(leaks, fmt.Sprintf("token %q leaked in %s: ...%s...",
					real, fname, content[start:end]))
			}
		}
	}
	return leaks
}

// ---------------------------------------------------------------------------
// main

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run generate.go <inputDir> <outputDir>")
		os.Exit(2)
	}
	inDir, outDir := os.Args[1], os.Args[2]

	var files []string
	_ = filepath.WalkDir(inDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".xml") {
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no input files found in", inDir)
		os.Exit(1)
	}
	sort.Strings(files)

	c, docs, finances, err := collect(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "collect:", err)
		os.Exit(1)
	}
	if c.syntheticEmitDoc() == "" {
		fmt.Fprintln(os.Stderr, "emit document not found; cannot rebuild access keys")
		os.Exit(1)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	manifest := make([]manifestEntry, 0, len(files))
	for _, path := range files {
		key := strings.TrimSuffix(filepath.Base(path), ".xml")
		newKey := rebuildAccessKey(key, c.syntheticEmitDoc())
		out, err := transform(docs[key], c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "transform:", key, err)
			os.Exit(1)
		}
		outPath := filepath.Join(outDir, newKey+".xml")
		if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := verify(outPath, newKey); err != nil {
			fmt.Fprintln(os.Stderr, "verify:", key, err)
			os.Exit(1)
		}

		infNFe := infNFeOf(docs[key].Root())
		fin := finances[key]
		icmsTag, icmsCst, pisCst, cofinsCst := scanTaxGroups(infNFe)
		origFolder := filepath.Base(filepath.Dir(filepath.Dir(path)))
		nnf := 0
		if n, err := strconv.Atoi(elText(infNFe.SelectElement("ide"), "nNF")); err == nil {
			nnf = n
		}
		manifest = append(manifest, manifestEntry{
			File:           newKey + ".xml",
			Group:          classify(fin.cfop, fin.finNFe),
			OriginalFolder: origFolder,
			NNF:            nnf,
			NatOp:          fin.natOp,
			CFOP:           fin.cfop,
			FinNFe:         fin.finNFe,
			ICMSTag:        icmsTag,
			ICMSCST:        icmsCst,
			PISCST:         pisCst,
			COFINSCST:      cofinsCst,
			TPag:           elText(findTag(infNFe, "detPag"), "tPag"),
			ModFrete:       elText(infNFe.SelectElement("transp"), "modFrete"),
			DestType:       destDocType(infNFe),
			Discount:       vDescNonZero(infNFe),
			InfAdProd:      infNFe.SelectElement("det").SelectElement("infAdProd") != nil,
			Representative: representativeNNF[nnf],
		})
	}

	outputs, err := loadOutputs(outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	leaks := checkLeaks(outputs, c)
	sort.Strings(leaks)
	if len(leaks) > 0 {
		fmt.Fprintln(os.Stderr, "LEAKS detected:")
		for _, l := range leaks[:min(10, len(leaks))] {
			fmt.Fprintln(os.Stderr, "  ", l)
		}
		os.Exit(1)
	}

	sort.Slice(manifest, func(i, j int) bool { return manifest[i].File < manifest[j].File })
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(outDir, "manifest.json"), append(mb, '\n'), 0o644)
	writeREADME(outDir, manifest)

	orders := []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G?"}
	counts := map[string]int{}
	for _, m := range manifest {
		counts[m.Group]++
	}
	var parts []string
	for _, g := range orders {
		if counts[g] > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", g, counts[g]))
		}
	}
	fmt.Printf("wrote %d fixtures, groups: %v, tokens synthetic: %d\n", len(files), parts, len(c.tokens))
}

func scanTaxGroups(infNFe *etree.Element) (string, string, string, string) {
	icmsTag, icmsCst, pisCst, cofinsCst := "", "", "", ""
	if det := infNFe.SelectElement("det"); det != nil {
		if imp := det.SelectElement("imposto"); imp != nil {
			if icms := imp.SelectElement("ICMS"); icms != nil {
				for _, ch := range icms.ChildElements() {
					if strings.HasPrefix(ch.Tag, "ICMS") {
						icmsTag = ch.Tag
						icmsCst = elText(ch, "CST")
						break
					}
				}
			}
			if pis := imp.SelectElement("PIS"); pis != nil {
				for _, ch := range pis.ChildElements() {
					if strings.HasPrefix(ch.Tag, "PIS") {
						pisCst = elText(ch, "CST")
						break
					}
				}
			}
			if cofins := imp.SelectElement("COFINS"); cofins != nil {
				for _, ch := range cofins.ChildElements() {
					if strings.HasPrefix(ch.Tag, "COFINS") {
						cofinsCst = elText(ch, "CST")
						break
					}
				}
			}
		}
	}
	return icmsTag, icmsCst, pisCst, cofinsCst
}

func destDocType(infNFe *etree.Element) string {
	dest := infNFe.SelectElement("dest")
	if dest == nil {
		return ""
	}
	if dest.SelectElement("CPF") != nil {
		return "CPF"
	}
	if dest.SelectElement("CNPJ") != nil {
		return "CNPJ"
	}
	return ""
}

// vDescNonZero: first vDesc anywhere in the tree starting with a nonzero digit.
// Document order places det/prod before total, so item-level discounts win.
func vDescNonZero(infNFe *etree.Element) bool {
	v := findTag(infNFe, "vDesc")
	if v == nil {
		return false
	}
	t := strings.TrimSpace(v.Text())
	return len(t) > 0 && t[0] >= '1' && t[0] <= '9'
}

var _ = base64.StdEncoding // referenced via syntheticBase64

func writeREADME(outDir string, manifest []manifestEntry) {
	counts := map[string]int{}
	for _, m := range manifest {
		counts[m.Group]++
	}
	var b strings.Builder
	b.WriteString(`# Real-world NF-e replay fixtures (anonymized)

Deterministically anonymized copies of the full corpus in
` + "`docs/research_nfe_use_case/`" + ` (115 authorized ` + "`nfeProc`" + ` 4.00 files, serie 921,
farm in MT). People and companies are SYNTHETIC: names, CPF/CNPJ (valid check
digits), IE, emails, phones, addresses, driver names, plates, CND codes,
protocol numbers, certificates and digests. Digital signatures are structurally
intact but cryptographically invalid by construction.

All business content is preserved from the real files: natOp, CFOP, finNFe,
ICMS tag/CST, PIS/COFINS CSTs, tPag, modFrete, transport data shapes, weights,
values, dates, IBS/CBS blocks, infCpl structure (with synthetic identity
tokens). Municipal names and nNF numbering are public/business data and kept.

Regenerate with:

    go run pkg/nfe/xml/testdata/realworld/generate.go docs/research_nfe_use_case pkg/nfe/xml/testdata/realworld

Group taxonomy: docs/research_nfe_use_case/USE_CASES.md.
Per-file metadata (group + discriminants + provenance): manifest.json.

## Groups
`)
	for _, g := range []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G?"} {
		if counts[g] > 0 {
			fmt.Fprintf(&b, "- %s: %d files\n", g, counts[g])
		}
	}
	b.WriteString("\nRepresentatives per group (fixture access key):\n")
	for _, m := range manifest {
		if m.Representative {
			fmt.Fprintf(&b, "- %s: %s\n", m.Group, strings.TrimSuffix(m.File, ".xml"))
		}
	}
	_ = os.WriteFile(filepath.Join(outDir, "README.md"), []byte(b.String()), 0o644)
}
