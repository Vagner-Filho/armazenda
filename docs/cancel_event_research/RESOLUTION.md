# NF-e cancellation fix — resolution

**Status**: closed. Production homolog cancellation succeeds.

## Root cause

`pkg/nfe/xml/event.go:17` declared `VersaoEvento = "1.0"` instead of `"1.00"`. SEFAZ MT homolog enforces the schema pattern `pattern value="1\.00"` for cancellation events and rejects `1.0` with **cStat 595 — "A versao do leiaute da NF-e utilizada nao e mais valida"**. The same rejection occurs for `2.00` and any other value that doesn't match the pattern.

## Fix

Three changes:

1. **`pkg/nfe/xml/event.go:17`** — `VersaoEvento = "1.00"`. This constant flows into four XML emit sites (`envEvento/@versao`, `evento/@versao`, `infEvento/verEvento`, `detEvento/@versao`).

2. **`pkg/nfe/xml/envelope.go`** — added `BuildSOAPEnvelopeWithCabecMsg(serviceNamespace, bodyContent, cUF, versaoDados)`. Emits a SOAP 1.2 envelope with `<soap:Header><nfeCabecMsg xmlns="…"><cUF/><versaoDados/></nfeCabecMsg></soap:Header>` plus the `<soap:Body><nfeDadosMsg>` wrapping the inner payload. The legacy `BuildSOAPEnvelope` is unchanged so the other four flows (auth, query, status, SVC-status) are not affected.

3. **`pkg/nfe/service/invoice_service.go:159-160`** — cancellation site computes `cUF := defaults.UFCode(s.config.StateUF)` and calls `xml.BuildSOAPEnvelopeWithCabecMsg(ns, signedEventXML, cUF, "1.00")`.

The debug `fmt.Println`s that were added during the investigation have been removed.

## Why the cert-OID hypothesis was a red herring

We originally believed the cert's malformed `2.16.76.1.3.1` OID encoding was causing cStat 617 ("Chave de Acesso inválida (CNPJ zerado ou dígito inválido)"). It was not. The actual sequence was:

| State | Result |
|---|---|
| Original code (`versao="1.0"`, no nfeCabecMsg) | cStat 617 — **G04e/J02e rule fired first** (cert OID check appeared to fire because cert OID was the first thing we validated in our analysis) |
| After adding `<nfeCabecMsg>` to the envelope (`versao="1.0"` still wrong) | cStat 595 — **D05 rule fired** (verEvento version check), shifting the failure mode |
| After correcting `versao="1.00"` (with nfeCabecMsg) | **cStat 135 — cancellation succeeded** |

The cert-OID-vs-access-key mismatch that we hypothesized (G04e/J02e) was a misinterpretation of why the validation order changed. The cert owner's success was simply because their library/tool defaulted `verEvento` to `"1.00"`.

The cert OID is still non-conformant per DOC-ICP-04 v4.0 (length byte `0x07` instead of `0x0B`, BCD value `19768350346310` instead of the holder's CPF `83503463100`). SEFAZ MT homolog doesn't enforce this for cancellation events, but other flows/states may.

## Tests added

`pkg/nfe/xml/envelope_test.go` (new file):

- `TestVersaoEventoVersion` — locks `VersaoEvento` to `"1.00"` and rejects `"1.0"`. Prevents regression to the short form.
- `TestBuildSOAPEnvelopeWithCabecMsg` — locks the envelope structure: SOAP 1.2 namespace, `<soap:Header>` with `<nfeCabecMsg>` carrying the right `cUF`/`versaoDados`, `<soap:Body>` with `<nfeDadosMsg>` carrying the inner payload, and ordering (header before body).
- `TestBuildSOAPEnvelope_LegacyUnchanged` — regression guard for the other four flows: legacy envelope stays on SOAP 1.2 with no `<soap:Header>` and no `<nfeCabecMsg>`.

## Follow-up (not done in this PR)

- **Cert OID audit** — the cert's `2.16.76.1.3.1` OtherName encoding is still non-conformant. SEFAZ MT homolog didn't enforce it for this flow, but other states or other event types might. Worth a separate investigation.
- **Per-UF envelope architecture** — the `BuildSOAPEnvelopeWithCabecMsg` is currently hard-coded for MT. As the user noted, the 27 states have quirks; a per-UF configuration or strategy pattern is the long-term path.
- **`BuildSOAPEnvelopeWithCabecMsg` empty-body panic** — both this and the legacy `BuildSOAPEnvelope` will panic on empty body input (`etree.AddChild(nil)`). Pre-existing, out of scope for this PR. Documented in the skipped test `TestBuildSOAPEnvelopeWithCabecMsg_EmptyBody`.

## Cleanup of investigation artefacts

- The `~/cancel_nfe_investigation/opencode/` directory (PoC scripts, captured wire dumps, PyNFe reproduction) is left in place as evidence and a starting point for future investigations. It is not under version control in the Armazenda repo.
- `docs/cancel_event_research/diff_PL_010d_v1.03_vs_PL_010d_v1.01.md` remains — it documents the XSD changes between PL v1.01 and PL_010d_v1.03, useful background for future schema migration work.
- `docs/cancel_event_research/tier_1_and_tier_2/` (PDF + markdown dumps of NT 2011.006c, NT 2026.004, NT 2025.001 Conjunta, PL_010d, PL_010e) remains as reference.
