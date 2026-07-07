package nfe_model

import (
	"context"
	"errors"
	"fmt"

	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type NFeModel struct {
	pool *pgxpool.Pool
}

var nfeModelImpl *NFeModel

func InitNFeModel(pool *pgxpool.Pool) (*NFeModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if nfeModelImpl == nil {
		nfeModelImpl = &NFeModel{
			pool: pool,
		}
	}

	return nfeModelImpl, nil
}

func GetNFeModel() *NFeModel {
	if nfeModelImpl == nil {
		panic("nfe model hasnt been initialized")
	}
	return nfeModelImpl
}

// FarmConfig represents the NFe configuration for a farm.
type FarmConfig struct {
	ID                           int
	FarmID                       int
	CertificatePath              string
	CertificateData              []byte
	CertificatePasswordEncrypted string
	Environment                  int
	Serie                        int
	NextNumber                   int
	TaxRegime                    int
	EmitterType                  int
	CNPJEmitter                  *string
	CPFEmitter                   *string
	IEEmitter                    string
	EmitterUF                    string
	DefaultModFrete              int
	DefaultCFOP                  string
	DefaultCEST                  *string
	DefaultUnit                  string
	DefaultICMSCST               *string
	DefaultPISCST                *string
	DefaultCOFINSCST             *string
	DefaultNaturezaOp            *string
	ICMSRate                     decimal.Decimal
	PISRate                      decimal.Decimal
	COFINSRate                   decimal.Decimal
}

// GetFarmConfig returns the NFe configuration for a farm.
func (m *NFeModel) GetFarmConfig(farmID uint32) (*FarmConfig, error) {
	query := `
		SELECT id, farm_id, certificate_path, certificate_data, certificate_password_encrypted,
		       environment, serie, next_number, tax_regime, emitter_type,
		       cnpj_emitter, cpf_emitter, ie_emitter, emitter_uf, default_mod_frete,
		       default_cfop, default_cest, default_unit,
		       default_icms_cst, default_pis_cst, default_cofins_cst, default_natureza_op,
		       icms_rate, pis_rate, cofins_rate
		FROM nfe_farm_config
		WHERE farm_id = $1
	`
	row := m.pool.QueryRow(context.Background(), query, farmID)

	var cfg FarmConfig
	var cnpj, cpf *string
	var defaultCest, defaultIcmsCst, defaultPisCst, defaultCofinsCst, defaultNaturezaOp *string
	err := row.Scan(
		&cfg.ID, &cfg.FarmID, &cfg.CertificatePath, &cfg.CertificateData, &cfg.CertificatePasswordEncrypted,
		&cfg.Environment, &cfg.Serie, &cfg.NextNumber, &cfg.TaxRegime, &cfg.EmitterType,
		&cnpj, &cpf, &cfg.IEEmitter, &cfg.EmitterUF, &cfg.DefaultModFrete,
		&cfg.DefaultCFOP, &defaultCest, &cfg.DefaultUnit,
		&defaultIcmsCst, &defaultPisCst, &defaultCofinsCst, &defaultNaturezaOp,
		&cfg.ICMSRate, &cfg.PISRate, &cfg.COFINSRate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get farm config: %w", err)
	}

	cfg.CNPJEmitter = cnpj
	cfg.CPFEmitter = cpf
	cfg.DefaultCEST = defaultCest
	cfg.DefaultICMSCST = defaultIcmsCst
	cfg.DefaultPISCST = defaultPisCst
	cfg.DefaultCOFINSCST = defaultCofinsCst
	cfg.DefaultNaturezaOp = defaultNaturezaOp
	return &cfg, nil
}

// UpsertFarmConfig inserts or updates the NFe configuration for a farm.
func (m *NFeModel) UpsertFarmConfig(cfg FarmConfig) error {
	query := `
		INSERT INTO nfe_farm_config (
			farm_id, certificate_path, certificate_data, certificate_password_encrypted, environment,
			serie, next_number, tax_regime, emitter_type, cnpj_emitter, cpf_emitter,
			ie_emitter, emitter_uf, default_mod_frete,
			default_cfop, default_cest, default_unit,
			default_icms_cst, default_pis_cst, default_cofins_cst, default_natureza_op,
			icms_rate, pis_rate, cofins_rate
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)
		ON CONFLICT (farm_id) DO UPDATE SET
			certificate_path = EXCLUDED.certificate_path,
			certificate_data = EXCLUDED.certificate_data,
			certificate_password_encrypted = EXCLUDED.certificate_password_encrypted,
			environment = EXCLUDED.environment,
			serie = EXCLUDED.serie,
			next_number = EXCLUDED.next_number,
			tax_regime = EXCLUDED.tax_regime,
			emitter_type = EXCLUDED.emitter_type,
			cnpj_emitter = EXCLUDED.cnpj_emitter,
			cpf_emitter = EXCLUDED.cpf_emitter,
			ie_emitter = EXCLUDED.ie_emitter,
			emitter_uf = EXCLUDED.emitter_uf,
			default_mod_frete = EXCLUDED.default_mod_frete,
			default_cfop = EXCLUDED.default_cfop,
			default_cest = EXCLUDED.default_cest,
			default_unit = EXCLUDED.default_unit,
			default_icms_cst = EXCLUDED.default_icms_cst,
			default_pis_cst = EXCLUDED.default_pis_cst,
			default_cofins_cst = EXCLUDED.default_cofins_cst,
			default_natureza_op = EXCLUDED.default_natureza_op,
			icms_rate = EXCLUDED.icms_rate,
			pis_rate = EXCLUDED.pis_rate,
			cofins_rate = EXCLUDED.cofins_rate,
			modified_at = CURRENT_TIMESTAMP
	`
	_, err := m.pool.Exec(context.Background(), query,
		cfg.FarmID, cfg.CertificatePath, cfg.CertificateData, cfg.CertificatePasswordEncrypted,
		cfg.Environment, cfg.Serie, cfg.NextNumber, cfg.TaxRegime, cfg.EmitterType,
		cfg.CNPJEmitter, cfg.CPFEmitter, cfg.IEEmitter, cfg.EmitterUF, cfg.DefaultModFrete,
		cfg.DefaultCFOP, cfg.DefaultCEST, cfg.DefaultUnit,
		cfg.DefaultICMSCST, cfg.DefaultPISCST, cfg.DefaultCOFINSCST, cfg.DefaultNaturezaOp,
		cfg.ICMSRate, cfg.PISRate, cfg.COFINSRate,
	)
	return err
}

// AllocateNumber atomically allocates a new invoice number for a farm/serie.
func (m *NFeModel) AllocateNumber(farmID uint32, serie int) (int, error) {
	var number int
	err := m.pool.QueryRow(context.Background(), "SELECT nfe_allocate_number($1, $2)", farmID, serie).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("failed to allocate number: %w", err)
	}
	return number, nil
}

// Invoice represents a tracked NF-e invoice.
type Invoice struct {
	ID                   int
	DepartureID          uint32
	AccessKey            string
	Serie                int
	Number               int
	Status               string
	CFOP                 string
	NCM                  string
	QuantityKG           decimal.Decimal
	UnitPrice            decimal.Decimal
	TotalValue           decimal.Decimal
	ICMSValue            *decimal.Decimal
	XMLSigned            *string
	XMLAuthorized        *string
	Protocol             *string
	SefazStatusCode      *string
	SefazMotive          *string
	RejectionReason      *string
	CancellationReason   *string
	RetryCount           int
	TpEmis               int
	DhCont               interface{}
	XJust                *string
	ContingencyParentID  *int
	SVCEndpointUsed      *string
	CreatedAt            interface{}
	SignedAt             interface{}
	SentAt               interface{}
	AuthorizedAt         interface{}
	TaxRates             *entity.TaxRates
}

// CreateInvoice creates a new invoice record with default tpEmis=1 (normal).
func (m *NFeModel) CreateInvoice(departureID uint32, accessKey string, serie, number int, cfop, ncm string, quantityKG, unitPrice, totalValue decimal.Decimal, taxRates *entity.TaxRates) (int, error) {
	return m.CreateInvoiceWithEmission(departureID, accessKey, serie, number, cfop, ncm, quantityKG, unitPrice, totalValue, 1, nil, "", taxRates)
}

// CreateInvoiceWithEmission creates a new invoice record with a specific emission type.
// If taxRates is non-nil and contains at least one non-nil rate, a row is also
// inserted into nfe_invoice_tax_rates recording the user-provided rates.
func (m *NFeModel) CreateInvoiceWithEmission(departureID uint32, accessKey string, serie, number int, cfop, ncm string, quantityKG, unitPrice, totalValue decimal.Decimal, tpEmis int, dhCont interface{}, xJust string, taxRates *entity.TaxRates) (int, error) {
	query := `
		INSERT INTO nfe_invoice (
			departure_id, access_key, serie, number, status,
			cfop, ncm, quantity_kg, unit_price, total_value,
			tp_emis, dh_cont, x_just
		)
		VALUES ($1, $2, $3, $4, 'draft', $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	var id int
	err := m.pool.QueryRow(context.Background(), query,
		departureID, accessKey, serie, number, cfop, ncm, quantityKG, unitPrice, totalValue,
		tpEmis, dhCont, xJust).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create invoice: %w", err)
	}

	if err := m.insertInvoiceTaxRates(id, taxRates); err != nil {
		return 0, err
	}
	return id, nil
}

// insertInvoiceTaxRates persists a row in nfe_invoice_tax_rates when at least
// one rate was provided by the user. Nil inputs are written as NULL columns,
// preserving the per-axis "user provided / not provided" distinction.
func (m *NFeModel) insertInvoiceTaxRates(invoiceID int, taxRates *entity.TaxRates) error {
	if taxRates == nil {
		return nil
	}
	if taxRates.ICMSRate == nil && taxRates.PISRate == nil && taxRates.COFINSRate == nil {
		return nil
	}
	_, err := m.pool.Exec(context.Background(),
		`INSERT INTO nfe_invoice_tax_rates (invoice_id, icms_rate, pis_rate, cofins_rate)
		 VALUES ($1, $2, $3, $4)`,
		invoiceID, taxRates.ICMSRate, taxRates.PISRate, taxRates.COFINSRate,
	)
	if err != nil {
		return fmt.Errorf("failed to insert invoice tax rates: %w", err)
	}
	return nil
}

// GetInvoiceTaxRates returns the user-provided tax rate overrides for an
// invoice. Returns (nil, nil) when no row exists (legacy invoice or emission
// where the user did not provide any rate).
func (m *NFeModel) GetInvoiceTaxRates(invoiceID int) (*entity.TaxRates, error) {
	var tr entity.TaxRates
	err := m.pool.QueryRow(context.Background(),
		`SELECT icms_rate, pis_rate, cofins_rate
		 FROM nfe_invoice_tax_rates
		 WHERE invoice_id = $1`,
		invoiceID,
	).Scan(&tr.ICMSRate, &tr.PISRate, &tr.COFINSRate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice tax rates: %w", err)
	}
	return &tr, nil
}

// DeleteInvoiceTaxRates removes the tax rate row for an invoice.
func (m *NFeModel) DeleteInvoiceTaxRates(invoiceID int) error {
	_, err := m.pool.Exec(context.Background(),
		`DELETE FROM nfe_invoice_tax_rates WHERE invoice_id = $1`, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice tax rates: %w", err)
	}
	return nil
}

// UpdateInvoiceXML updates the signed XML of an invoice and transitions status.
func (m *NFeModel) UpdateInvoiceXML(id int, xmlSigned string) error {
	query := `
		UPDATE nfe_invoice
		SET xml_signed = $2, status = 'signed', signed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, xmlSigned)
	return err
}

// UpdateInvoiceXMLWithStatus updates the signed XML and sets a custom status.
func (m *NFeModel) UpdateInvoiceXMLWithStatus(id int, xmlSigned, status string) error {
	query := `
		UPDATE nfe_invoice
		SET xml_signed = $2, status = $3, signed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, xmlSigned, status)
	return err
}

// UpdateInvoiceStatus updates the status of an invoice.
func (m *NFeModel) UpdateInvoiceStatus(id int, status, protocol, sefazCode, sefazMotive string) error {
	query := `
		UPDATE nfe_invoice
		SET status = $2, protocol = $3, sefaz_status_code = $4, sefaz_motive = $5,
		    authorized_at = CASE WHEN $2 = 'authorized' THEN CURRENT_TIMESTAMP ELSE authorized_at END
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, status, protocol, sefazCode, sefazMotive)
	return err
}

// UpdateInvoiceAuthorizedXML stores the <nfeProc> wrapper XML (signed NFe +
// protocol) in the xml_authorized column.
func (m *NFeModel) UpdateInvoiceAuthorizedXML(id int, xmlAuthorized string) error {
	query := `UPDATE nfe_invoice SET xml_authorized = $2 WHERE id = $1`
	_, err := m.pool.Exec(context.Background(), query, id, xmlAuthorized)
	return err
}

// GetInvoiceByDeparture returns the invoice for a departure.
func (m *NFeModel) GetInvoiceByDeparture(departureID uint32) (*Invoice, error) {
	query := `
		SELECT i.id, i.departure_id, i.access_key, i.serie, i.number, i.status, i.cfop, i.ncm,
		       i.quantity_kg, i.unit_price, i.total_value, i.icms_value, i.xml_signed, i.xml_authorized,
		       i.protocol, i.sefaz_status_code, i.sefaz_motive, i.rejection_reason, i.cancellation_reason,
		       i.retry_count, i.tp_emis, i.dh_cont, i.x_just, i.contingency_parent_id, i.svc_endpoint_used,
		       i.created_at, i.signed_at, i.sent_at, i.authorized_at,
		       t.icms_rate, t.pis_rate, t.cofins_rate
		FROM nfe_invoice i
		LEFT JOIN nfe_invoice_tax_rates t ON t.invoice_id = i.id
		WHERE i.departure_id = $1
		ORDER BY i.id DESC
		LIMIT 1
	`
	row := m.pool.QueryRow(context.Background(), query, departureID)

	inv, err := scanInvoice(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return inv, nil
}

// GetMunicipio returns an IBGE municipality by name and UF.
func (m *NFeModel) GetMunicipio(name, uf string) (string, error) {
	query := `
		SELECT code FROM ibge_municipio
		WHERE name ILIKE $1 AND uf = $2
		LIMIT 1
	`
	var code string
	err := m.pool.QueryRow(context.Background(), query, name, uf).Scan(&code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return code, nil
}

// InvoiceForRetry represents an invoice ready for retry processing.
type InvoiceForRetry struct {
	ID          int
	DepartureID uint32
	AccessKey   string
	Status      string
	XMLSigned   string
	RetryCount  int
	TpEmis      int
	LastRetryAt interface{}
}

// GetPendingInvoicesForRetry returns invoices with status 'pending' that need status polling,
// respecting capped exponential backoff.
func (m *NFeModel) GetPendingInvoicesForRetry() ([]InvoiceForRetry, error) {
	query := `
		SELECT id, departure_id, access_key, status, xml_signed, retry_count, tp_emis, last_retry_at
		FROM nfe_invoice
		WHERE status = 'pending'
		  AND retry_count < 10
		  AND (
		    last_retry_at IS NULL
		    OR last_retry_at <= CURRENT_TIMESTAMP - (INTERVAL '5 minutes' * LEAST(POWER(2, retry_count), 12))
		  )
		ORDER BY created_at ASC
		LIMIT 50
	`
	rows, err := m.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invoices: %w", err)
	}
	defer rows.Close()

	var invoices []InvoiceForRetry
	for rows.Next() {
		var inv InvoiceForRetry
		var xmlSigned *string
		var lastRetryAt interface{}
		if err := rows.Scan(&inv.ID, &inv.DepartureID, &inv.AccessKey, &inv.Status, &xmlSigned, &inv.RetryCount, &inv.TpEmis, &lastRetryAt); err != nil {
			return nil, fmt.Errorf("failed to scan pending invoice: %w", err)
		}
		if xmlSigned != nil {
			inv.XMLSigned = *xmlSigned
		}
		inv.LastRetryAt = lastRetryAt
		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

// GetDraftInvoicesForRetry returns draft invoices that are eligible for auto-retry
// when SVC becomes active. Only returns invoices created within the last 24 hours.
func (m *NFeModel) GetDraftInvoicesForRetry() ([]InvoiceForRetry, error) {
	query := `
		SELECT id, departure_id, access_key, status, xml_signed, retry_count, tp_emis, last_retry_at
		FROM nfe_invoice
		WHERE status = 'draft'
		  AND retry_count < 5
		  AND created_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
		  AND (
		    last_retry_at IS NULL
		    OR last_retry_at <= CURRENT_TIMESTAMP - (INTERVAL '10 minutes' * LEAST(POWER(2, retry_count), 6))
		  )
		ORDER BY created_at ASC
		LIMIT 50
	`
	rows, err := m.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get draft invoices: %w", err)
	}
	defer rows.Close()

	var invoices []InvoiceForRetry
	for rows.Next() {
		var inv InvoiceForRetry
		var xmlSigned *string
		var lastRetryAt interface{}
		if err := rows.Scan(&inv.ID, &inv.DepartureID, &inv.AccessKey, &inv.Status, &xmlSigned, &inv.RetryCount, &inv.TpEmis, &lastRetryAt); err != nil {
			return nil, fmt.Errorf("failed to scan draft invoice: %w", err)
		}
		if xmlSigned != nil {
			inv.XMLSigned = *xmlSigned
		}
		inv.LastRetryAt = lastRetryAt
		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

// ResetPendingBackoff clears last_retry_at for all pending invoices
// so they become immediately eligible for status polling on the next worker run.
func (m *NFeModel) ResetPendingBackoff() error {
	query := `
		UPDATE nfe_invoice
		SET last_retry_at = NULL
		WHERE status = 'pending'
	`
	_, err := m.pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to reset pending backoff: %w", err)
	}
	return nil
}

// IncrementRetryCount increments the retry count and sets last_retry_at.
func (m *NFeModel) IncrementRetryCount(id int) error {
	query := `
		UPDATE nfe_invoice
		SET retry_count = retry_count + 1, last_retry_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id)
	return err
}

// UpdateInvoiceSent marks an invoice as sent and records the timestamp.
func (m *NFeModel) UpdateInvoiceSent(id int) error {
	query := `
		UPDATE nfe_invoice
		SET sent_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id)
	return err
}

// SupersedeInvoice marks an old invoice as superseded and links a new contingency invoice to it.
func (m *NFeModel) SupersedeInvoice(oldID, newID int) error {
	query := `
		UPDATE nfe_invoice
		SET status = 'superseded', contingency_parent_id = $2
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, oldID, newID)
	if err != nil {
		return fmt.Errorf("failed to supersede invoice: %w", err)
	}
	return nil
}

// GetInvoicesByFarm returns invoices for a farm with pagination, ordered by creation date descending.
func (m *NFeModel) GetInvoicesByFarm(farmID uint32, page int) ([]Invoice, int, error) {
	pageSize := 10
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM nfe_invoice i
		JOIN departure d ON d.id = i.departure_id
		WHERE d.farm = $1
	`
	var total int
	if err := m.pool.QueryRow(context.Background(), countQuery, farmID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	query := `
		SELECT i.id, i.departure_id, i.access_key, i.serie, i.number, i.status, i.cfop, i.ncm,
		       i.quantity_kg, i.unit_price, i.total_value, i.icms_value, i.xml_signed, i.xml_authorized,
		       i.protocol, i.sefaz_status_code, i.sefaz_motive, i.rejection_reason, i.cancellation_reason,
		       i.retry_count, i.tp_emis, i.dh_cont, i.x_just, i.contingency_parent_id, i.svc_endpoint_used,
		       i.created_at, i.signed_at, i.sent_at, i.authorized_at,
		       t.icms_rate, t.pis_rate, t.cofins_rate
		FROM nfe_invoice i
		JOIN departure d ON d.id = i.departure_id
		LEFT JOIN nfe_invoice_tax_rates t ON t.invoice_id = i.id
		WHERE d.farm = $1
		ORDER BY i.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.pool.Query(context.Background(), query, farmID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}
	defer rows.Close()

	var invoices []Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, *inv)
	}

	return invoices, total, rows.Err()
}

// GetInvoiceByAccessKey returns an invoice by its access key.
func (m *NFeModel) GetInvoiceByAccessKey(accessKey string) (*Invoice, error) {
	query := `
		SELECT i.id, i.departure_id, i.access_key, i.serie, i.number, i.status, i.cfop, i.ncm,
		       i.quantity_kg, i.unit_price, i.total_value, i.icms_value, i.xml_signed, i.xml_authorized,
		       i.protocol, i.sefaz_status_code, i.sefaz_motive, i.rejection_reason, i.cancellation_reason,
		       i.retry_count, i.tp_emis, i.dh_cont, i.x_just, i.contingency_parent_id, i.svc_endpoint_used,
		       i.created_at, i.signed_at, i.sent_at, i.authorized_at,
		       t.icms_rate, t.pis_rate, t.cofins_rate
		FROM nfe_invoice i
		LEFT JOIN nfe_invoice_tax_rates t ON t.invoice_id = i.id
		WHERE i.access_key = $1
		LIMIT 1
	`
	row := m.pool.QueryRow(context.Background(), query, accessKey)

	inv, err := scanInvoice(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return inv, nil
}

// invoiceScanner is the common surface for pgx.Row and pgx.Rows so scanInvoice
// can serve both single-row and multi-row queries.
type invoiceScanner interface {
	Scan(dest ...any) error
}

// scanInvoice reads a row produced by a query that LEFT JOINs
// nfe_invoice_tax_rates and materializes the optional tax rate columns into
// the Invoice.TaxRates field.
func scanInvoice(row invoiceScanner) (*Invoice, error) {
	var inv Invoice
	var icmsValue *decimal.Decimal
	var xmlSigned, xmlAuthorized, protocol, sefazCode, sefazMotive, rejectionReason, cancellationReason, xJust, svcEndpoint *string
	var contingencyParentID *int
	var icmsRate, pisRate, cofinsRate *decimal.Decimal
	err := row.Scan(
		&inv.ID, &inv.DepartureID, &inv.AccessKey, &inv.Serie, &inv.Number, &inv.Status,
		&inv.CFOP, &inv.NCM, &inv.QuantityKG, &inv.UnitPrice, &inv.TotalValue, &icmsValue,
		&xmlSigned, &xmlAuthorized, &protocol, &sefazCode, &sefazMotive, &rejectionReason,
		&cancellationReason, &inv.RetryCount, &inv.TpEmis, &inv.DhCont, &xJust, &contingencyParentID, &svcEndpoint,
		&inv.CreatedAt, &inv.SignedAt, &inv.SentAt, &inv.AuthorizedAt,
		&icmsRate, &pisRate, &cofinsRate,
	)
	if err != nil {
		return nil, err
	}

	inv.ICMSValue = icmsValue
	inv.XMLSigned = xmlSigned
	inv.XMLAuthorized = xmlAuthorized
	inv.Protocol = protocol
	inv.SefazStatusCode = sefazCode
	inv.SefazMotive = sefazMotive
	inv.RejectionReason = rejectionReason
	inv.CancellationReason = cancellationReason
	inv.XJust = xJust
	inv.ContingencyParentID = contingencyParentID
	inv.SVCEndpointUsed = svcEndpoint

	if icmsRate != nil || pisRate != nil || cofinsRate != nil {
		inv.TaxRates = &entity.TaxRates{
			ICMSRate:   icmsRate,
			PISRate:    pisRate,
			COFINSRate: cofinsRate,
		}
	}
	return &inv, nil
}
