# NF-e Package Maintenance Guide

> **Document:** Operational maintenance for `pkg/nfe` and the Armazenda NF-e integration  
> **Scope:** Legal, accounting, technical, and infrastructure upkeep  
> **Applies to:** Any state registered in `pkg/nfe/sefaz/endpoints.go`

---

## 1. Digital Certificate (Certificado Digital A1)

### 1.1 Renewal Cycle
- **A1 certificates expire annually** (12 months from issue date).
- **Action required:** Renew 30 days before expiry to avoid service interruption.
- **Where configured:** `nfe_farm_config.certificate_path` and `certificate_password_encrypted`.

### 1.2 Certificate Expiry Monitoring
```sql
-- Query to find certificates nearing expiry (if you track expiry date)
-- Recommended: add an `certificate_expires_at` column to `nfe_farm_config`
```

**Recommended:** Add a dashboard alert or cron job that:
1. Reads the certificate file periodically
2. Checks `NotAfter` field
3. Sends email/notification when < 30 days remain

### 1.3 Certificate Replacement Procedure
1. Obtain new `.pfx` from certifying authority (e.g., Serasa, Certisign, Valid)
2. Upload new certificate to server (update `certificate_path`)
3. Encrypt the new password with `NFE_CERT_KEY` env var
4. Update `nfe_farm_config.certificate_password_encrypted`
5. **Test in homologation first:** Send a test NF-e to verify
6. Update `certificate_expires_at` if tracked

### 1.4 Certificate Chain Issues
- SEFAZ rotates its root/intermediate certificates occasionally.
- If mTLS handshake fails after a SEFAZ maintenance window, the certificate chain may have changed.
- **Fix:** Download the new certificate chain from the state's SEFAZ portal and ensure it's included in the `.pfx` or system trust store.

---

## 2. SEFAZ Layout (Schema) Updates

### 2.1 Layout Version Lifecycle
- SEFAZ releases new NF-e layouts periodically (e.g., `4.00` → `4.01`).
- **Transition period:** Usually 3–6 months where both versions are accepted.
- **Mandatory switch:** Old layouts are rejected after the transition period.

### 2.2 What to Update
When a new layout is released:

| Component | Action |
|-----------|--------|
| `pkg/nfe/xml/builder.go` | Update XML element names, add new fields, remove deprecated ones |
| `pkg/nfe/entity/invoice.go` | Add/remove fields in structs |
| `pkg/nfe/defaults/agriculture.go` | Update `VersaoLayout` constant |
| `pkg/nfe/validate/` | Download new XSDs from SEFAZ, update validation logic |

### 2.3 Tracking Layout Changes
- Monitor: https://www.nfe.fazenda.gov.br/portal/principal.aspx
- SEFAZ publishes "Notas Técnicas" (technical notes) detailing changes.
- Subscribe to email alerts from your state's SEFAZ if available.

---

## 3. SEFAZ Endpoint URL Changes

### 3.1 When URLs Change
- Infrastructure migrations (e.g., state moves from Own → SVRS)
- Domain changes, load balancer rotations
- SSL/TLS certificate rotations affecting WSDL paths

### 3.2 What to Update
Edit `pkg/nfe/sefaz/endpoints.go`:
- Update `Production` or `Homologation` URLs in the affected `EndpointSet`
- Update `Namespace` if the WSDL namespace changes

### 3.3 Monitoring for URL Changes
- Run `CheckStatus()` daily against all registered states
- If a state returns 404 or connection refused, check the official SEFAZ portal
- The `frones/nfe` (Go) or `wmixvideo/nfe` (Java) open-source projects often update URLs quickly — monitor their commits

---

## 4. Tax Law & Fiscal Code Updates

### 4.1 NCM (Nomenclatura Comum do Mercosul)
- **Changes:** MERCOSUL periodically updates NCM codes for products.
- **Impact:** Invalid NCM = NF-e rejection by SEFAZ.
- **Action:**
  - Monitor: https://www.gov.br/produtividade-e-comercio-exterior/pt-br/assuntos/comercio-exterior/classificacao-fiscal
  - Update `nfe_product_config.ncm` in the database
  - Update `defaults/agriculture.go` default NCMs if changed

### 4.2 CFOP (Código Fiscal de Operações e Prestações)
- **Changes:** New CFOPs are created for new types of operations; existing ones may be deprecated.
- **Action:** Update `nfe_product_config.default_cfop` and `defaults/agriculture.go`

### 4.3 CST / CSOSN Changes
- **Changes:** ICMS legislation changes can alter CST/CSOSN applicability.
- **Impact:** Wrong CST = NF-e rejection or incorrect tax calculation.
- **Action:** Update `defaults/agriculture.go` and `nfe_product_config` defaults.

### 4.4 ICMS Rate Changes
- **Changes:** States can change internal ICMS rates via state law.
- **Action:** Update tax calculation logic in `pkg/nfe/service/invoice_service.go` or the Armazenda bridge.

### 4.5 GTIN Requirements
- SEFAZ has progressively tightened GTIN (barcode) requirements.
- For bulk agricultural products without GTIN, the field must contain `"SEM GTIN"`.
- **Action:** Monitor SEFAZ technical notes for GTIN rule changes.

---

## 5. IBGE Municipality Updates

### 5.1 When Municipalities Change
- New municipalities are created (rare, but happens)
- Municipality codes may change (extremely rare)
- Municipality names may be officially altered

### 5.2 What to Update
1. Update `model/armazenda_database/seed_municipios.go`
2. Add new entries to `municipiosSeedData`
3. Run the seed function or execute manual `INSERT` into `ibge_municipio`

### 5.3 Source of Truth
- https://servicodados.ibge.gov.br/api/v1/localidades/estados/{uf}/municipios
- Check annually or when a user reports a missing municipality

---

## 6. Database Maintenance

### 6.1 Legal Retention Period
- Brazilian tax law requires **minimum 5 years** of NF-e XML storage.
- **Applies to:** `nfe_invoice.xml_signed`, `nfe_invoice.xml_authorized`, `nfe_invoice.xml_cancel_event`

### 6.2 Archive Strategy
```sql
-- Example: archive invoices older than 5 years
-- Move to cold storage (S3, Glacier, etc.) before deleting from DB

SELECT id, access_key, xml_signed, xml_authorized
FROM nfe_invoice
WHERE created_at < NOW() - INTERVAL '5 years';
```

### 6.3 Index Maintenance
```sql
-- Ensure indexes are healthy
REINDEX INDEX idx_nfe_invoice_farm_serie_number;
```

### 6.4 Cleanup
- **Pending invoices older than 30 days:** Investigate why they were never sent. May indicate a SEFAZ outage or a bug.
- **Denied invoices:** Keep for audit; do not delete.
- **Cancelled invoices:** Keep XML cancellation event forever (legal proof).

---

## 7. Security Maintenance

### 7.1 Certificate Password Encryption Key
- The certificate password is encrypted with a key from the `NFE_CERT_KEY` environment variable.
- **Rotation procedure:**
  1. Decrypt all existing passwords with old key
  2. Generate new encryption key
  3. Re-encrypt all passwords with new key
  4. Update `NFE_CERT_KEY` env var
  5. Restart application

### 7.2 Server Security
- Certificate `.pfx` files must be readable only by the application user (`chmod 600`)
- Store certificates outside the web root
- Rotate server TLS certificates independently of NF-e A1 certificates

### 7.3 Access Control
- Only admin users should be able to modify `nfe_farm_config`
- Log all invoice issuance, cancellation, and configuration changes

---

## 8. Monitoring & Alerting

### 8.1 Recommended Alerts

| Alert | Frequency | Action on Failure |
|-------|-----------|-------------------|
| SEFAZ status check | Every 5 min | Queue invoices, notify users |
| Certificate expiry | Daily | Alert if < 30 days |
| Pending invoice count | Every 15 min | Alert if > 10 pending for > 1 hour |
| Failed invoice rate | Every 1 hour | Alert if > 5% rejection rate |

### 8.2 Health Check Endpoint
```go
// Example health check for your monitoring system
func NFeHealthCheck() error {
    // Check SEFAZ status for each active farm
    // Check certificate validity
    // Check pending invoice queue depth
}
```

---

## 9. Contingency Procedures

### 9.1 SEFAZ Down (Planned Maintenance)
1. Invoices accumulate in `pending` status
2. Background retry worker should use exponential backoff
3. Notify users: "SEFAZ em manutenção. NF-e será enviada automaticamente quando o serviço voltar."

### 9.2 SEFAZ Down (Unexpected Outage)
1. Do not panic. Invoices in `draft` or `pending` status are safe.
2. User can download the signed XML manually (legally valid as backup)
3. When SEFAZ returns, retry queue processes automatically
4. **No invoice numbering is lost** — numbers are only allocated when building the XML

### 9.3 Invoice Rejection
- SEFAZ returns a rejection code and motive
- User must correct the data (e.g., wrong CNPJ, invalid NCM, incorrect ICMS)
- The invoice status becomes `denied`
- User can rebuild with corrections — a **new invoice number** is allocated automatically

### 9.4 Network Issues
- mTLS requires stable HTTPS connection
- If your server is behind a proxy/firewall, ensure outbound HTTPS to SEFAZ endpoints is allowed
- Whitelist SEFAZ IP ranges if using strict firewall rules

---

## 10. Testing & Validation

### 10.1 Homologation Testing
- **Before any production deployment:** Test against the homologation environment
- Send at least one test NF-e per state you support
- Verify: authorization, consultation, cancellation, status check

### 10.2 After Any Code Change
Run:
```bash
go build ./...
go test ./...
```

### 10.3 After Any SEFAZ Change
- Update endpoints or schema
- Test in homologation for 24 hours
- Monitor production closely for the first week

---

## 11. Dependency Updates

### 11.1 Critical Dependencies
| Package | Purpose | Update Frequency |
|---------|---------|-----------------|
| `github.com/russellhaering/goxmldsig` | XML-DSig | Monthly check |
| `software.sslmate.com/src/go-pkcs12` | A1 certificate loading | Monthly check |
| `github.com/beevik/etree` | XML manipulation | Monthly check |
| `github.com/signintech/gopdf` | DANFE PDF | Monthly check |
| Standard library `crypto/tls` | TLS/mTLS | With Go upgrades |

### 11.2 Update Procedure
```bash
go get -u ./...
go mod tidy
go build ./...
go test ./...
```

**Caution:** Major version bumps in crypto libraries may change behavior. Always test in homologation after dependency updates.

---

## 12. Legal & Accounting Checklist

### 12.1 Monthly
- [ ] Verify no `authorized` invoices are missing XML backup
- [ ] Check rejection rate (should be < 1%)
- [ ] Verify certificate validity
- [ ] Review pending queue

### 12.2 Quarterly
- [ ] Test homologation environment for all supported states
- [ ] Review SEFAZ technical notes for layout/schema changes
- [ ] Review NCM/CFOP/CST changes with accountant
- [ ] Verify IBGE municipality list is current

### 12.3 Annually
- [ ] Renew A1 certificate before expiry
- [ ] Review tax regime configuration (Simples Nacional, Lucro Real, Lucro Presumido)
- [ ] Archive invoices older than 5 years to cold storage
- [ ] Full disaster recovery test: restore DB, verify all XMLs are intact

### 12.4 On-Demand
- [ ] When SEFAZ announces new layout version
- [ ] When state ICMS legislation changes
- [ ] When a new farm is onboarded in a new state
- [ ] When certificate chain issues arise

---

## 13. Emergency Contacts & Resources

| Resource | URL |
|----------|-----|
| SEFAZ National Portal | https://www.nfe.fazenda.gov.br/portal/ |
| SEFAZ MT | https://www.sefaz.mt.gov.br/ |
| NCM Lookup | https://www.gov.br/produtividade-e-comercio-exterior/pt-br/assuntos/comercio-exterior/classificacao-fiscal |
| IBGE API | https://servicodados.ibge.gov.br/api/docs/localidades |
| Open Source Ref (Java) | https://github.com/wmixvideo/nfe |
| Open Source Ref (Go) | https://github.com/frones/nfe |

---

## 14. Common Error Codes & Quick Fixes

| SEFAZ Code | Meaning | Quick Fix |
|-----------|---------|-----------|
| `215` | Rejeicao: Falha no schema XML | Layout version mismatch — update XML builder |
| `245` | Rejeicao: CNPJ Emitente nao cadastrado | Verify CNPJ in `nfe_farm_config` |
| `254` | Rejeicao: Sigla da UF do Emitente diverge | `emitter_uf` does not match certificate UF |
| `267` | Rejeicao: Chave de Acesso invalida | Check access key generation logic |
| `377` | Rejeicao: CFOP invalido | Update CFOP in product config |
| `400` | Rejeicao: NCM invalido | Update NCM in product config |
| `600` | Rejeicao: Chave de Acesso ja existe | Invoice already authorized — do not resend |
| `694` | Rejeicao: CSOSN invalido | Check Simples Nacional CSOSN codes |

Full list: https://www.nfe.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=/ku8KWY0Kgc=

---

*Last updated: 15/05/2026*  
*Maintainers: Update this document whenever a new maintenance pattern is discovered.*
