package xml_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"

	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/xml"
)

// realworld_test.go replays the anonymized real-world NF-e corpus
// (pkg/nfe/xml/testdata/realworld) against Armazenda offline:
//   - corpus inventory checks (all files classified, counts sum to 115),
//   - parser round-trip over every file (ParseDANFEData),
//   - builder golden tests per use-case group with an explicit allowlist of
//     known gaps, mirrored in docs/research_nfe_use_case/USE_CASES.md.
//
// The fixtures are deterministically anonymized copies of the real corpus in
// docs/research_nfe_use_case (people/companies synthesized, business data
// preserved) — see testdata/realworld/README.md and generate.go.

const realWorldDir = "testdata/realworld"

type rwManifestEntry struct {
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

func loadManifest(t *testing.T) []rwManifestEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(realWorldDir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest.json not readable: %v", err)
	}
	var entries []rwManifestEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("manifest.json invalid: %v", err)
	}
	if len(entries) != 115 {
		t.Fatalf("expected 115 manifest entries, got %d", len(entries))
	}
	return entries
}

// TestRealWorld_CorpusInventory verifies the fixture set mirrors the
// real-world classification: 115 files, 7 groups (84/23/3/2/1/1/1), each file
// present on disk with a 44-digit access-key filename, and one representative
// per group.
func TestRealWorld_CorpusInventory(t *testing.T) {
	entries := loadManifest(t)

	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Group]++
		st, err := os.Stat(filepath.Join(realWorldDir, e.File))
		if err != nil {
			t.Errorf("missing fixture file %s: %v", e.File, err)
			continue
		}
		if st.IsDir() {
			t.Errorf("fixture %s is a directory", e.File)
		}
		key := strings.TrimSuffix(e.File, ".xml")
		if len([]rune(key)) != 44 {
			t.Errorf("fixture name %s is not a 44-char access key", e.File)
		}
		for _, r := range key {
			if r < '0' || r > '9' {
				t.Errorf("access key %s contains non-digit", key)
				break
			}
		}
		if e.Group == "" || e.Group == "G?" {
			t.Errorf("file %s unclassified", e.File)
		}
	}
	wantCounts := map[string]int{"G1": 84, "G2": 23, "G3": 3, "G4": 2, "G5": 1, "G6": 1, "G7": 1}
	for g, want := range wantCounts {
		if counts[g] != want {
			t.Errorf("group %s: got %d files, want %d", g, counts[g], want)
		}
	}
	if len(counts) != len(wantCounts) {
		t.Errorf("unexpected extra groups: %v", counts)
	}

	reps := map[string]bool{}
	for _, e := range entries {
		if e.Representative {
			if reps[e.Group] {
				t.Errorf("group %s has more than one representative", e.Group)
			}
			reps[e.Group] = true
		}
	}
	if len(reps) != 7 {
		t.Errorf("expected one representative per group (7), got %d: %v", len(reps), reps)
	}
}

// TestRealWorld_ParseDANFEData feeds every corpus file through the DANFE
// parser and asserts format-level invariants.
func TestRealWorld_ParseDANFEData(t *testing.T) {
	entries := loadManifest(t)

	for _, e := range entries {
		e := e
		name := e.Group + "/" + e.File[len(e.File)-20:]
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(realWorldDir, e.File))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			data, err := xml.ParseDANFEData(string(raw))
			if err != nil {
				t.Fatalf("ParseDANFEData failed: %v", err)
			}

			// authorization + identification
			wantKey := strings.TrimSuffix(e.File, ".xml")
			if data.AccessKey != wantKey {
				t.Errorf("AccessKey %s != filename %s", data.AccessKey, wantKey)
			}
			if data.CStat != "100" {
				t.Errorf("cStat = %q, want 100", data.CStat)
			}
			if data.Protocol == "" {
				t.Error("protocol empty")
			}
			if data.Serie != 921 {
				t.Errorf("serie = %d, want 921", data.Serie)
			}
			if data.Numero != e.NNF {
				t.Errorf("nNF = %d, want %d", data.Numero, e.NNF)
			}
			if data.TpEmis != "1" {
				t.Errorf("tpEmis = %q, want 1", data.TpEmis)
			}
			if data.TpAmb != "1" {
				t.Errorf("tpAmb = %q, want 1", data.TpAmb)
			}
			if data.TpNF != "1" {
				t.Errorf("tpNF = %q, want 1", data.TpNF)
			}
			if data.NaturezaOp != e.NatOp {
				t.Errorf("natOp mismatch: %q vs %q", data.NaturezaOp, e.NatOp)
			}

			// parties
			if data.EmitterCNPJ == "" {
				t.Error("emitter document empty")
			}
			switch e.DestType {
			case "CNPJ":
				if len(data.DestCNPJ) != 14 {
					t.Errorf("dest CNPJ length %d, want 14", len(data.DestCNPJ))
				}
			case "CPF":
				if len(data.DestCNPJ) != 11 {
					t.Errorf("dest CPF length %d, want 11", len(data.DestCNPJ))
				}
			}
			if data.DestName == "" {
				t.Error("dest name empty")
			}

			// items + taxes
			if len(data.Products) != 1 {
				t.Fatalf("expected exactly 1 det, got %d", len(data.Products))
			}
			p := data.Products[0]
			if p.CFOP != e.CFOP {
				t.Errorf("CFOP = %q, want %q", p.CFOP, e.CFOP)
			}
			if p.CST != e.ICMSCST {
				t.Errorf("ICMS CST = %q, want %q", p.CST, e.ICMSCST)
			}
			if !p.Quantity.IsPositive() {
				t.Error("item quantity not positive")
			}
			if !p.Total.IsPositive() {
				t.Error("item total not positive")
			}
			// the corpus is PISNT/COFINSNT (CST 08) everywhere: no computed
			// PIS/COFINS amounts.
			if !p.VPIS.IsZero() || !p.VCOFINS.IsZero() {
				t.Error("PIS/COFINS amounts should be zero (NT) in this corpus")
			}

			// totals + transport
			if !data.TotalValue.IsPositive() {
				t.Errorf("vNF not positive: %s", data.TotalValue)
			}
			if !data.VBC.IsZero() || !data.VICMS.IsZero() {
				t.Error("ICMS vBC/vICMS nonzero — corpus uses zero-valued ICMS")
			}
			if e.Discount && data.VDesc.IsZero() {
				t.Error("discounted file has zero vDesc total")
			}
			if !e.Discount && !data.VDesc.IsZero() {
				t.Errorf("non-discounted file has vDesc %s", data.VDesc)
			}
			if e.ModFrete != "" && data.ModFrete != e.ModFrete {
				t.Errorf("modFrete = %q, want %q", data.ModFrete, e.ModFrete)
			}
			if !data.PesoB.IsPositive() {
				t.Error("pesoB not positive")
			}
			if data.InfCpl == "" {
				t.Error("infCpl empty")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// builder goldens

var rwBaseAllowed = map[string]bool{
	// must-detect capability gaps (fail loudly once fixed — see
	// docs/research_nfe_use_case/USE_CASES.md gap table)
	"vTotTrib": true, // cosmetic: builder skips optional vTotTrib
	"iiGroup":  true, // cosmetic: builder skips optional <II> zero group
	// tolerated current-behavior diffs
	"dhEmiNow":     true, // builder uses time.Now()
	"accessKeyDV":  true, // built AAMM comes from time.Now()
	"verProc":      true, // "Armazenda-1.0" vs third-party "1.0.0"
	"gMonoTot":     true, // builder intentionally skips <gMono> in totals
	"cobr":         true, // builder emits cobr/fat; real corpus never does
	"indPag":       true, // builder omits indPag=0; the corpus files carrying indPag=0 become fixture-only diffs
	"ibsRateDP":    true, // builder writes pIBSMun with 4dp; real third-party uses 2dp
	"vItemFormula": true, // builder uses the full NT 2025.002 vItem formula (vProd − vDesc + vIBSUF + vIBSMun + vCBS) per the user decision in docs/sdd/nfe_builder_replay_gaps/sdd.md; fixtures carry the flat vProd − vDesc variant, so det/vItem and total/vNFTot intentionally diverge
}

// Retired required gaps (fixed by docs/sdd/nfe_builder_replay_gaps — the
// diff paths still map to keys via rwClassifyGap, so a regression in any of
// these behaviors fails as an unexpected gap):
//   - ICMSTag: buildICMS switches per CST (ICMS00/20/40/51/90).
//   - pisTag/cofinsTag: CST 04–09 emit PISNT/COFINSNT.
//   - vItem/vNFTot: emitted per the NT 2025.002 formula (key now
//     vItemFormula, an allowed divergence vs the flat-formula fixtures).
//   - dhSaiEnt, infRespTec, paisFields, destEmail, xPag, indPag (0 omitted),
//     veicTranspUFEmpty: emitted/omitted with the new pooled inputs
//     (DhSaiEnt, RespTec, Recipient.Email, PagamentoDetalhe.XPag).
//     Retired G3 required gaps: finNFe (FinNFe input, fixture=4 honored),
//     nfref (NFRefs input); retired G3 allowed keys: prodVDesc +
//     icmsTotDesc (Produto.VDesc summed into ICMSTot/vDesc); retired G2
//     allowed key: entrega (EntregaData input).

// rwAllowedPerGroup lists extra tolerated gaps per group.
var rwAllowedPerGroup = map[string]map[string]bool{
	"G3": {"pIBSUFDrift": true}, // builder derives per-item pIBSUF from rounded vIBSUF/vBC; real files write flat 0.1000
}

// rwRequiredGaps per group: gaps that MUST still appear until fixed; if a fix
// lands, update these sets together with USE_CASES.md.
var rwRequiredGaps = map[string][]string{
	"G1": {"vTotTrib", "iiGroup"},
	"G2": {"vTotTrib", "iiGroup"},
	"G3": {"vTotTrib", "iiGroup"},
	"G4": {"vTotTrib", "iiGroup"},
	"G5": {"vTotTrib", "iiGroup"},
	"G6": {"vTotTrib", "iiGroup"},
	"G7": {"vTotTrib", "iiGroup"},
}

// TestRealWorldBuilder_PerGroupGolden replays one representative per use-case
// group through the XML builder and diffs the generated document against the
// real anonymized fixture. Observed diffs must be inside the allowed gap
// registry, and every required gap must actually appear — so any fix (or any
// new divergence) makes this test fail and forces a registry update.
func TestRealWorldBuilder_PerGroupGolden(t *testing.T) {
	entries := loadManifest(t)
	reps := map[string]rwManifestEntry{}
	for _, e := range entries {
		if e.Representative {
			reps[e.Group] = e
		}
	}

	for _, g := range []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7"} {
		g := g
		e, ok := reps[g]
		if !ok {
			t.Fatalf("no representative for group %s", g)
		}
		t.Run(g, func(t *testing.T) {
			fixture, err := os.ReadFile(filepath.Join(realWorldDir, e.File))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			fdoc := etree.NewDocument()
			if err := fdoc.ReadFromBytes(fixture); err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			input := rwFixtureToInput(t, fdoc)

			doc, err := xml.NewBuilder().Build(input)
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}

			fin := rwInfNFeOf(fdoc.Root())
			bin := rwInfNFeOf(doc.Root())

			allowed := map[string]bool{}
			for k := range rwBaseAllowed {
				allowed[k] = true
			}
			for k := range rwAllowedPerGroup[g] {
				allowed[k] = true
			}

			observed := map[string]bool{}
			var unclassified []string
			rwDiffSkeletons(t, fin, bin, func(path string) {
				gap := rwClassifyGap(path)
				if gap == "" {
					unclassified = append(unclassified, path)
					return
				}
				if !allowed[gap] {
					t.Errorf("unexpected gap %q at path %s", gap, path)
				}
				observed[gap] = true
			})
			if len(unclassified) > 0 {
				sort.Strings(unclassified)
				t.Errorf("unclassified diff paths (add to gap registry): %v", unclassified)
			}

			for _, must := range rwRequiredGaps[g] {
				if !observed[must] {
					t.Errorf("required gap %q did not appear — behavior changed; update the gap registry and USE_CASES.md", must)
				}
			}
		})
	}
}

// rwClassifyGap maps a diff path to a gap-registry key.
func rwClassifyGap(path string) string {
	switch {
	case path == "ide/dhSaiEnt":
		return "dhSaiEnt"
	case path == "ide/dhEmi":
		return "dhEmiNow"
	case path == "ide/cDV":
		return "accessKeyDV"
	case path == "ide/verProc":
		return "verProc"
	case path == "ide/finNFe":
		return "finNFe"
	case strings.HasPrefix(path, "det/imposto/ICMS/"):
		return "ICMSTag"
	case strings.HasPrefix(path, "det/imposto/PIS"):
		return "pisTag"
	case strings.HasPrefix(path, "det/imposto/COFINS"):
		return "cofinsTag"
	case path == "det/imposto/vTotTrib":
		return "vTotTrib"
	case strings.HasPrefix(path, "det/imposto/II"):
		return "iiGroup"
	// vItem/vNFTot now use the full NT 2025.002 formula (vProd − vDesc +
	// vIBSUF + vIBSMun + vCBS) while the fixtures carry the flat vProd −
	// vDesc variant — an allowed divergence (vItemFormula), see
	// docs/sdd/nfe_builder_replay_gaps/sdd.md.
	case path == "det/vItem", path == "total/vNFTot":
		return "vItemFormula"
	case strings.HasPrefix(path, "total/IBSCBSTot/gMono"):
		return "gMonoTot"
	case path == "total/ICMSTot/vDesc":
		return "icmsTotDesc"
	case path == "det/imposto/IBSCBS/gIBSCBS/gIBSUF/pIBSUF":
		return "pIBSUFDrift"
	case path == "det/imposto/IBSCBS/gIBSCBS/gIBSMun/pIBSMun":
		return "ibsRateDP"
	case path == "det/prod/vDesc":
		return "prodVDesc"
	case path == "cobr/fat/nFat", path == "cobr/fat/vOrig", path == "cobr/fat/vLiq":
		return "cobr"
	case path == "dest/email":
		return "destEmail"
	case path == "pag/detPag/xPag":
		return "xPag"
	case path == "pag/detPag/indPag":
		return "indPag"
	case path == "transp/veicTransp/UF":
		return "veicTranspUFEmpty"
	case strings.HasPrefix(path, "infRespTec"):
		return "infRespTec"
	case strings.HasPrefix(path, "entrega"):
		return "entrega"
	case strings.HasPrefix(path, "ide/NFref"):
		return "nfref"
	case path == "emit/enderEmit/cPais", path == "emit/enderEmit/xPais",
		path == "dest/enderDest/cPais", path == "dest/enderDest/xPais":
		return "paisFields"
	}
	return ""
}

// ---------------------------------------------------------------------------
// helpers

func rwInfNFeOf(root *etree.Element) *etree.Element {
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

// rwDiffSkeletons walks both infNFe trees, mapping terminal element paths to
// text values, and reports each diff path via the callback.
func rwDiffSkeletons(t *testing.T, a, b *etree.Element, onDiff func(path string)) {
	t.Helper()
	ms := rwSkeleton("", a)
	mb := rwSkeleton("", b)
	if os.Getenv("RW_DEBUG") != "" {
		keys := make([]string, 0, len(ms)+len(mb))
		for k := range ms {
			keys = append(keys, k)
		}
		for k := range mb {
			if _, ok := ms[k]; !ok {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			va, oa := ms[k]
			vb, ob := mb[k]
			mark := "OK"
			if !oa && ob {
				mark = "BUILT-ONLY"
			} else if oa && !ob {
				mark = "FIXTURE-ONLY"
			} else if va != vb {
				mark = "VALUE-DIFF"
			}
			t.Logf("%-12s %-50s fixture=%-12q built=%q", mark, k, va, vb)
		}
	}
	paths := map[string]bool{}
	for p := range ms {
		paths[p] = true
	}
	for p := range mb {
		paths[p] = true
	}
	sorted := make([]string, 0, len(paths))
	for p := range paths {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)
	for _, p := range sorted {
		va, okA := ms[p]
		vb, okB := mb[p]
		switch {
		case okA && !okB, !okA && okB, va != vb:
			onDiff(p)
		}
	}
}

func rwSkeleton(prefix string, el *etree.Element) map[string]string {
	m := map[string]string{}
	if el == nil {
		return m
	}
	for _, ch := range el.ChildElements() {
		p := ch.Tag
		if prefix != "" {
			p = prefix + "/" + ch.Tag
		}
		if len(ch.ChildElements()) == 0 {
			m[p] = ch.Text()
		} else {
			for k, v := range rwSkeleton(p, ch) {
				m[k] = v
			}
		}
	}
	return m
}

func rwElText(parent *etree.Element, tag string) string {
	if parent == nil {
		return ""
	}
	e := parent.SelectElement(tag)
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.Text())
}

func rwDec(t *testing.T, s string) decimal.Decimal {
	t.Helper()
	if s == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		t.Fatalf("bad decimal %q: %v", s, err)
	}
	return d
}

func rwInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// rwFixtureToInput maps an anonymized real-world nfeProc fixture into the
// builder's entity.InvoiceInput as faithfully as the entity supports (the
// entity-side gaps the replay surfaced get mapped here as the builder gains
// the corresponding inputs; remaining divergences live in the gap registry).
func rwFixtureToInput(t *testing.T, doc *etree.Document) entity.InvoiceInput {
	t.Helper()
	root := doc.Root()
	infNFe := rwInfNFeOf(root)
	ide := infNFe.SelectElement("ide")
	emitEl := infNFe.SelectElement("emit")
	destEl := infNFe.SelectElement("dest")
	det := infNFe.SelectElement("det")
	prod := det.SelectElement("prod")
	imp := det.SelectElement("imposto")

	input := entity.InvoiceInput{
		Serie:       rwInt(rwElText(ide, "serie")),
		Numero:      rwInt(rwElText(ide, "nNF")),
		CNF:         rwElText(ide, "cNF"),
		NaturezaOp:  rwElText(ide, "natOp"),
		Environment: rwInt(rwElText(ide, "tpAmb")),
		FinNFe:      rwElText(ide, "finNFe"),
	}
	if input.FinNFe == "" {
		input.FinNFe = "1"
	}
	if dhSaiEnt := rwElText(ide, "dhSaiEnt"); dhSaiEnt != "" {
		parsed, err := time.Parse(time.RFC3339, dhSaiEnt)
		if err != nil {
			t.Fatalf("dhSaiEnt %q not RFC3339: %v", dhSaiEnt, err)
		}
		input.DhSaiEnt = &parsed
	}
	for _, refEl := range ide.SelectElements("NFref") {
		if ref := rwElText(refEl, "refNFe"); ref != "" {
			input.NFRefs = append(input.NFRefs, ref)
		}
	}

	// emitter
	etypedoc := rwElText(emitEl, "CPF")
	etypedocTag := "CPF"
	if etypedoc == "" {
		etypedoc = rwElText(emitEl, "CNPJ")
		etypedocTag = "CNPJ"
	}
	if ender := emitEl.SelectElement("enderEmit"); ender != nil {
		input.Emitter = entity.EmitterData{
			Type:       map[bool]int{true: 2, false: 1}[etypedocTag == "CPF"],
			Document:   etypedoc,
			IE:         rwElText(emitEl, "IE"),
			XNome:      rwElText(emitEl, "xNome"),
			XFant:      rwElText(emitEl, "xFant"),
			Logradouro: rwElText(ender, "xLgr"),
			Numero:     rwElText(ender, "nro"),
			Bairro:     rwElText(ender, "xBairro"),
			CodigoMun:  rwElText(ender, "cMun"),
			Municipio:  rwElText(ender, "xMun"),
			UF:         rwElText(ender, "UF"),
			CEP:        rwElText(ender, "CEP"),
			Fone:       rwElText(ender, "fone"),
			CRT:        rwElText(emitEl, "CRT"),
		}
	}

	// recipient
	dtypedoc := rwElText(destEl, "CPF")
	dtypedocTag := "CPF"
	if dtypedoc == "" {
		dtypedoc = rwElText(destEl, "CNPJ")
		dtypedocTag = "CNPJ"
	}
	if ender := destEl.SelectElement("enderDest"); ender != nil {
		rec := entity.RecipientData{
			Type:       map[bool]int{true: 2, false: 1}[dtypedocTag == "CPF"],
			IE:         rwElText(destEl, "IE"),
			Email:      rwElText(destEl, "email"),
			XNome:      rwElText(destEl, "xNome"),
			Logradouro: rwElText(ender, "xLgr"),
			Numero:     rwElText(ender, "nro"),
			Bairro:     rwElText(ender, "xBairro"),
			CodigoMun:  rwElText(ender, "cMun"),
			Municipio:  rwElText(ender, "xMun"),
			UF:         rwElText(ender, "UF"),
			CEP:        rwElText(ender, "CEP"),
			IndIEDest:  rwElText(destEl, "indIEDest"),
		}
		if dtypedocTag == "CPF" {
			rec.CPF = dtypedoc
		} else {
			rec.CNPJ = dtypedoc
		}
		input.Recipient = rec
	}

	// product
	input.Items = []entity.ItemData{{
		Numero: 1,
		Produto: entity.ProdutoData{
			Codigo:   rwElText(prod, "cProd"),
			CEAN:     rwElText(prod, "cEAN"),
			XProd:    rwElText(prod, "xProd"),
			NCM:      rwElText(prod, "NCM"),
			CFOP:     rwElText(prod, "CFOP"),
			UCom:     rwElText(prod, "uCom"),
			QCom:     rwDec(t, rwElText(prod, "qCom")),
			VUnCom:   rwDec(t, rwElText(prod, "vUnCom")),
			VProd:    rwDec(t, rwElText(prod, "vProd")),
			CEANTrib: rwElText(prod, "cEANTrib"),
			UTrib:    rwElText(prod, "uTrib"),
			QTrib:    rwDec(t, rwElText(prod, "qTrib")),
			VUnTrib:  rwDec(t, rwElText(prod, "vUnTrib")),
			VDesc:    rwDec(t, rwElText(prod, "vDesc")),
			IndTot:   rwInt(rwElText(prod, "indTot")),
		},
		InfAdProd: rwElText(det, "infAdProd"),
	}}

	// taxes
	icms := imp.SelectElement("ICMS")
	for _, ch := range icms.ChildElements() {
		if strings.HasPrefix(ch.Tag, "ICMS") {
			input.Items[0].Imposto.ICMS = entity.ICMSData{
				Origem:  rwElText(ch, "orig"),
				CST:     rwElText(ch, "CST"),
				ModBC:   rwElText(ch, "modBC"),
				VBC:     rwDec(t, rwElText(ch, "vBC")),
				PICMS:   rwDec(t, rwElText(ch, "pICMS")),
				VICMS:   rwDec(t, rwElText(ch, "vICMS")),
				PRedBC:  rwDec(t, rwElText(ch, "pRedBC")),
				ModBCST: rwElText(ch, "modBCST"),
				VBCST:   rwDec(t, rwElText(ch, "vBCST")),
				PICMSST: rwDec(t, rwElText(ch, "pICMSST")),
				VICMSST: rwDec(t, rwElText(ch, "vICMSST")),
			}
			break
		}
	}
	if pisEl := imp.SelectElement("PIS"); pisEl != nil {
		for _, ch := range pisEl.ChildElements() {
			if strings.HasPrefix(ch.Tag, "PIS") {
				input.Items[0].Imposto.PIS = entity.PISData{CST: rwElText(ch, "CST")}
				break
			}
		}
	}
	if cofinsEl := imp.SelectElement("COFINS"); cofinsEl != nil {
		for _, ch := range cofinsEl.ChildElements() {
			if strings.HasPrefix(ch.Tag, "COFINS") {
				input.Items[0].Imposto.COFINS = entity.COFINSData{CST: rwElText(ch, "CST")}
				break
			}
		}
	}
	if ibs := imp.SelectElement("IBSCBS"); ibs != nil {
		if g := ibs.SelectElement("gIBSCBS"); g != nil {
			ibsData := entity.IBSCBSData{
				CST:        rwElText(ibs, "CST"),
				CClassTrib: rwElText(ibs, "cClassTrib"),
				VBC:        rwDec(t, rwElText(g, "vBC")),
				VIBS:       rwDec(t, rwElText(g, "vIBS")),
			}
			// per-item IBS/CBS values live inside the nested groups
			if suf := g.SelectElement("gIBSUF"); suf != nil {
				ibsData.VIBSUF = rwDec(t, rwElText(suf, "vIBSUF"))
			}
			if mun := g.SelectElement("gIBSMun"); mun != nil {
				ibsData.VIBSMun = rwDec(t, rwElText(mun, "vIBSMun"))
			}
			if cbs := g.SelectElement("gCBS"); cbs != nil {
				ibsData.PCBS = rwDec(t, rwElText(cbs, "pCBS"))
				ibsData.VCBS = rwDec(t, rwElText(cbs, "vCBS"))
			}
			input.Items[0].Imposto.IBSCBS = ibsData
		}
	}

	// delivery-site group (G2 vendas a ordem)
	if ent := infNFe.SelectElement("entrega"); ent != nil {
		edoc := rwElText(ent, "CPF")
		etype := 2
		if edoc == "" {
			edoc = rwElText(ent, "CNPJ")
			etype = 1
		}
		ed := entity.EntregaData{
			Type:       etype,
			XNome:      rwElText(ent, "xNome"),
			Logradouro: rwElText(ent, "xLgr"),
			Numero:     rwElText(ent, "nro"),
			Bairro:     rwElText(ent, "xBairro"),
			CodigoMun:  rwElText(ent, "cMun"),
			Municipio:  rwElText(ent, "xMun"),
			UF:         rwElText(ent, "UF"),
			CEP:        rwElText(ent, "CEP"),
			IE:         rwElText(ent, "IE"),
		}
		if etype == 2 {
			ed.CPF = edoc
		} else {
			ed.CNPJ = edoc
		}
		input.Entrega = &ed
	}

	// transport
	if transp := infNFe.SelectElement("transp"); transp != nil {
		input.Transport.ModFrete = rwInt(rwElText(transp, "modFrete"))
		if tra := transp.SelectElement("transporta"); tra != nil {
			traDoc := rwElText(tra, "CPF")
			traType := 2
			if traDoc == "" {
				traDoc = rwElText(tra, "CNPJ")
				traType = 1
			}
			tra := &entity.TransportadoraData{
				Type:      traType,
				XNome:     rwElText(tra, "xNome"),
				IE:        rwElText(tra, "IE"),
				Endereco:  rwElText(tra, "xEnder"),
				Municipio: rwElText(tra, "xMun"),
				UF:        rwElText(tra, "UF"),
			}
			if traType == 2 {
				tra.CPF = traDoc
			} else {
				tra.CNPJ = traDoc
			}
			input.Transport.Transportadora = tra
		}
		if veic := transp.SelectElement("veicTransp"); veic != nil {
			input.Transport.Veiculo = &entity.VeiculoData{
				Placa: rwElText(veic, "placa"),
				UF:    rwElText(veic, "UF"),
			}
		}
		if vol := transp.SelectElement("vol"); vol != nil {
			input.Transport.Volumes = []entity.VolumeData{{
				QVol:  rwInt(rwElText(vol, "qVol")),
				Esp:   rwElText(vol, "esp"),
				Marca: rwElText(vol, "marca"),
				NVol:  rwElText(vol, "nVol"),
				PesoL: rwDec(t, rwElText(vol, "pesoL")),
				PesoB: rwDec(t, rwElText(vol, "pesoB")),
			}}
		}
	}

	// payment
	input.Payment.IndPag = 0
	if pag := infNFe.SelectElement("pag"); pag != nil {
		for _, detPag := range pag.SelectElements("detPag") {
			input.Payment.IndPag = rwInt(rwElText(detPag, "indPag"))
			input.Payment.Detalhes = append(input.Payment.Detalhes, entity.PagamentoDetalhe{
				IndPag: rwInt(rwElText(detPag, "indPag")),
				TPag:   rwElText(detPag, "tPag"),
				XPag:   rwElText(detPag, "xPag"),
				VPag:   rwDec(t, rwElText(detPag, "vPag")),
			})
		}
	}

	// totals + additional info
	if total := infNFe.SelectElement("total"); total != nil {
		if icmsTot := total.SelectElement("ICMSTot"); icmsTot != nil {
			input.TotalValue = rwDec(t, rwElText(icmsTot, "vNF"))
		}
	}
	input.InformacoesAdicionais = rwElText(infNFe.SelectElement("infAdic"), "infCpl")

	// technical responsible (last child of infNFe in the corpus)
	if resp := infNFe.SelectElement("infRespTec"); resp != nil {
		input.RespTec = &entity.RespTecData{
			CNPJ:     rwElText(resp, "CNPJ"),
			XContato: rwElText(resp, "xContato"),
			Email:    rwElText(resp, "email"),
			Fone:     rwElText(resp, "fone"),
		}
	}

	return input
}
