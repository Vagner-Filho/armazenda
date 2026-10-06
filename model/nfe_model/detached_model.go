package nfe_model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// DetachedInvoiceItem represents a single item in a detached invoice
type DetachedInvoiceItem struct {
	FarmProductID *uint16         `json:"farm_product_id,omitempty"`
	ProductName   string          `json:"product_name"`
	NCM           string          `json:"ncm"`
	CEST          *string         `json:"cest,omitempty"`
	CFOP          string          `json:"cfop"`
	Unit          string          `json:"unit"`
	Quantity      decimal.Decimal `json:"quantity"`
	GrossWeight   decimal.Decimal `json:"gross_weight"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	TotalValue    decimal.Decimal `json:"total_value"`
}

// DetachedInvoice represents a detached NF-e (not tied to any departure)
type DetachedInvoice struct {
	ID                  int
	FarmID              uint32
	RecipientID         uint32
	AccessKey           string
	Serie               int
	Number              int
	Status              string
	NaturezaOp          *string
	ModFrete            *int
	TotalValue          decimal.Decimal
	IBSValue            decimal.Decimal
	CBSValue            decimal.Decimal
	Items               []DetachedInvoiceItem
	XMLSigned           *string
	XMLAuthorized       *string
	XMLCancelEvent      *string
	Protocol            *string
	SefazStatusCode     *string
	SefazMotive         *string
	RejectionReason     *string
	CancellationReason  *string
	TpEmis              int
	DhCont              interface{}
	XJust               *string
	ContingencyParentID *int
	SVCEndpointUsed     *string
	ICMSCST             *string
	PISCST              *string
	COFINSCST           *string
	IBSCST              *string
	CBSCST              *string
	CClassTrib          *string
	InfCpl              *string
	RetryCount          int
	LastRetryAt         interface{}
	CreatedAt           interface{}
	SignedAt            interface{}
	SentAt              interface{}
	AuthorizedAt        interface{}
	CancelledAt         interface{}
	TaxRates            *entity.TaxRates
}

// CreateDetachedInvoice creates a new detached invoice record
func (m *NFeModel) CreateDetachedInvoice(
	farmID uint32,
	recipientID uint32,
	accessKey string,
	serie, number int,
	naturezaOp *string,
	modFrete *int,
	totalValue, ibsValue, cbsValue decimal.Decimal,
	tpEmis int,
	taxRates *entity.TaxRates,
	overrides *entity.InvoiceOverrides,
	items []DetachedInvoiceItem,
) (int, error) {
	// Serialize items to JSON
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		INSERT INTO detached_nfe_invoice (
			farm_id, recipient_id, access_key, serie, number, status,
			natureza_op, mod_frete,
			total_value, ibs_value, cbs_value,
			items_json,
			tp_emis,
			icms_cst, pis_cst, cofins_cst,
			ibs_cst, cbs_cst, c_class_trib,
			inf_cpl
		)
		VALUES ($1, $2, $3, $4, $5, 'draft',
			$6, $7,
			$8, $9, $10,
			$11,
			$12,
			$13, $14, $15,
			$16, $17, $18,
			$19
		)
		RETURNING id
	`
	var id int
	err = m.pool.QueryRow(context.Background(), query,
		farmID, recipientID, accessKey, serie, number,
		naturezaOp, modFrete,
		totalValue, ibsValue, cbsValue,
		itemsJSON,
		tpEmis,
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.ICMSCST }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.PISCST }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.COFINSCST }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.IBSCST }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.CBSCST }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.CClassTrib }),
		safeString(overrides, func(o *entity.InvoiceOverrides) *string { return o.InfCpl }),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create detached invoice: %w", err)
	}

	// Insert tax rates if provided
	if err := m.insertDetachedInvoiceTaxRates(id, taxRates); err != nil {
		return 0, err
	}

	return id, nil
}

// insertDetachedInvoiceTaxRates persists tax rate overrides for a detached invoice
func (m *NFeModel) insertDetachedInvoiceTaxRates(invoiceID int, taxRates *entity.TaxRates) error {
	if taxRates == nil {
		return nil
	}
	if taxRates.ICMSRate == nil && taxRates.PISRate == nil && taxRates.COFINSRate == nil &&
		taxRates.IBSRate == nil && taxRates.CBSRate == nil {
		return nil
	}
	_, err := m.pool.Exec(context.Background(),
		`INSERT INTO detached_nfe_tax_rates (invoice_id, icms_rate, pis_rate, cofins_rate, ibs_rate, cbs_rate)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		invoiceID, taxRates.ICMSRate, taxRates.PISRate, taxRates.COFINSRate,
		taxRates.IBSRate, taxRates.CBSRate,
	)
	if err != nil {
		return fmt.Errorf("failed to insert detached invoice tax rates: %w", err)
	}
	return nil
}

// GetDetachedInvoicesByFarm returns detached invoices for a farm with pagination
func (m *NFeModel) GetDetachedInvoicesByFarm(farmID uint32, page int) ([]DetachedInvoice, int, error) {
	pageSize := 10
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM detached_nfe_invoice
		WHERE farm_id = $1
	`
	var total int
	if err := m.pool.QueryRow(context.Background(), countQuery, farmID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count detached invoices: %w", err)
	}

	query := `
		SELECT d.id, d.farm_id, d.recipient_id, d.access_key, d.serie, d.number, d.status,
			d.natureza_op, d.mod_frete,
			d.total_value, d.ibs_value, d.cbs_value,
			d.items_json,
			d.xml_signed, d.xml_authorized, d.xml_cancel_event,
			d.protocol, d.sefaz_status_code, d.sefaz_motive, d.rejection_reason, d.cancellation_reason,
			d.tp_emis, d.dh_cont, d.x_just, d.contingency_parent_id, d.svc_endpoint_used,
			d.icms_cst, d.pis_cst, d.cofins_cst, d.ibs_cst, d.cbs_cst, d.c_class_trib, d.inf_cpl,
			d.retry_count, d.last_retry_at,
			d.created_at, d.signed_at, d.sent_at, d.authorized_at, d.cancelled_at,
			t.icms_rate, t.pis_rate, t.cofins_rate, t.ibs_rate, t.cbs_rate
		FROM detached_nfe_invoice d
		LEFT JOIN detached_nfe_tax_rates t ON t.invoice_id = d.id
		WHERE d.farm_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.pool.Query(context.Background(), query, farmID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get detached invoices: %w", err)
	}
	defer rows.Close()

	var invoices []DetachedInvoice
	for rows.Next() {
		inv, err := scanDetachedInvoice(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan detached invoice: %w", err)
		}
		invoices = append(invoices, *inv)
	}

	return invoices, total, rows.Err()
}

// GetDetachedInvoiceByAccessKey returns a detached invoice by access key
func (m *NFeModel) GetDetachedInvoiceByAccessKey(accessKey string) (*DetachedInvoice, error) {
	query := `
		SELECT d.id, d.farm_id, d.recipient_id, d.access_key, d.serie, d.number, d.status,
			d.natureza_op, d.mod_frete,
			d.total_value, d.ibs_value, d.cbs_value,
			d.items_json,
			d.xml_signed, d.xml_authorized, d.xml_cancel_event,
			d.protocol, d.sefaz_status_code, d.sefaz_motive, d.rejection_reason, d.cancellation_reason,
			d.tp_emis, d.dh_cont, d.x_just, d.contingency_parent_id, d.svc_endpoint_used,
			d.icms_cst, d.pis_cst, d.cofins_cst, d.ibs_cst, d.cbs_cst, d.c_class_trib, d.inf_cpl,
			d.retry_count, d.last_retry_at,
			d.created_at, d.signed_at, d.sent_at, d.authorized_at, d.cancelled_at,
			t.icms_rate, t.pis_rate, t.cofins_rate, t.ibs_rate, t.cbs_rate
		FROM detached_nfe_invoice d
		LEFT JOIN detached_nfe_tax_rates t ON t.invoice_id = d.id
		WHERE d.access_key = $1
		LIMIT 1
	`
	row := m.pool.QueryRow(context.Background(), query, accessKey)

	inv, err := scanDetachedInvoice(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get detached invoice: %w", err)
	}
	return inv, nil
}

// UpdateDetachedInvoiceStatus updates the status of a detached invoice
func (m *NFeModel) UpdateDetachedInvoiceStatus(id int, status, protocol, sefazCode, sefazMotive string) error {
	query := `
		UPDATE detached_nfe_invoice
		SET status = $2, protocol = $3, sefaz_status_code = $4, sefaz_motive = $5,
		    authorized_at = CASE WHEN $2 = 'authorized' THEN CURRENT_TIMESTAMP ELSE authorized_at END
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, status, protocol, sefazCode, sefazMotive)
	return err
}

// UpdateDetachedInvoiceSignedXML stores the signed XML
func (m *NFeModel) UpdateDetachedInvoiceSignedXML(id int, xmlSigned string) error {
	query := `
		UPDATE detached_nfe_invoice
		SET xml_signed = $2, signed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, xmlSigned)
	return err
}

// UpdateDetachedInvoiceAuthorizedXML stores the authorized XML
func (m *NFeModel) UpdateDetachedInvoiceAuthorizedXML(id int, xmlAuthorized string) error {
	query := `UPDATE detached_nfe_invoice SET xml_authorized = $2 WHERE id = $1`
	_, err := m.pool.Exec(context.Background(), query, id, xmlAuthorized)
	return err
}

// UpdateDetachedInvoiceCancelled marks a detached invoice as cancelled
func (m *NFeModel) UpdateDetachedInvoiceCancelled(id int, reason, eventXML, sefazCode, sefazMotive string) error {
	query := `
		UPDATE detached_nfe_invoice
		SET status = 'cancelled', cancellation_reason = $2, xml_cancel_event = $3,
		    sefaz_status_code = $4, sefaz_motive = $5, cancelled_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id, reason, eventXML, sefazCode, sefazMotive)
	return err
}

// SupersedeDetachedInvoice marks an old detached invoice as superseded
func (m *NFeModel) SupersedeDetachedInvoice(oldID, newID int) error {
	query := `
		UPDATE detached_nfe_invoice
		SET status = 'superseded', contingency_parent_id = $2
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, oldID, newID)
	return err
}

// detachedInvoiceScanner is the common surface for pgx.Row and pgx.Rows
type detachedInvoiceScanner interface {
	Scan(dest ...any) error
}

// scanDetachedInvoice scans a row into a DetachedInvoice
func scanDetachedInvoice(row detachedInvoiceScanner) (*DetachedInvoice, error) {
	var inv DetachedInvoice
	var xmlSigned, xmlAuthorized, xmlCancelEvent, protocol, sefazCode, sefazMotive, rejectionReason, cancellationReason, xJust, svcEndpoint *string
	var contingencyParentID *int
	var naturezaOp *string
	var modFrete *int
	var icmsCST, pisCST, cofinsCST, ibsCST, cbsCST, cClassTrib, infCpl *string
	var icmsRate, pisRate, cofinsRate, ibsRate, cbsRate *decimal.Decimal
	var itemsJSON []byte

	err := row.Scan(
		&inv.ID, &inv.FarmID, &inv.RecipientID, &inv.AccessKey, &inv.Serie, &inv.Number, &inv.Status,
		&naturezaOp, &modFrete,
		&inv.TotalValue, &inv.IBSValue, &inv.CBSValue,
		&itemsJSON,
		&xmlSigned, &xmlAuthorized, &xmlCancelEvent,
		&protocol, &sefazCode, &sefazMotive, &rejectionReason, &cancellationReason,
		&inv.TpEmis, &inv.DhCont, &xJust, &contingencyParentID, &svcEndpoint,
		&icmsCST, &pisCST, &cofinsCST, &ibsCST, &cbsCST, &cClassTrib, &infCpl,
		&inv.RetryCount, &inv.LastRetryAt,
		&inv.CreatedAt, &inv.SignedAt, &inv.SentAt, &inv.AuthorizedAt, &inv.CancelledAt,
		&icmsRate, &pisRate, &cofinsRate, &ibsRate, &cbsRate,
	)
	if err != nil {
		return nil, err
	}

	inv.NaturezaOp = naturezaOp
	inv.ModFrete = modFrete
	inv.XMLSigned = xmlSigned
	inv.XMLAuthorized = xmlAuthorized
	inv.XMLCancelEvent = xmlCancelEvent
	inv.Protocol = protocol
	inv.SefazStatusCode = sefazCode
	inv.SefazMotive = sefazMotive
	inv.RejectionReason = rejectionReason
	inv.CancellationReason = cancellationReason
	inv.XJust = xJust
	inv.ContingencyParentID = contingencyParentID
	inv.SVCEndpointUsed = svcEndpoint
	inv.ICMSCST = icmsCST
	inv.PISCST = pisCST
	inv.COFINSCST = cofinsCST
	inv.IBSCST = ibsCST
	inv.CBSCST = cbsCST
	inv.CClassTrib = cClassTrib
	inv.InfCpl = infCpl

	// Parse items JSON
	if itemsJSON != nil {
		if err := json.Unmarshal(itemsJSON, &inv.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}
	}

	// Parse tax rates
	if icmsRate != nil || pisRate != nil || cofinsRate != nil || ibsRate != nil || cbsRate != nil {
		inv.TaxRates = &entity.TaxRates{
			ICMSRate:   icmsRate,
			PISRate:    pisRate,
			COFINSRate: cofinsRate,
			IBSRate:    ibsRate,
			CBSRate:    cbsRate,
		}
	}

	return &inv, nil
}

// AllocateDetachedNumber atomically allocates a new invoice number for a farm/serie
// Uses the same sequence as nfe_invoice to avoid number collisions
func (m *NFeModel) AllocateDetachedNumber(farmID uint32, serie int) (int, error) {
	var number int
	err := m.pool.QueryRow(context.Background(), "SELECT nfe_allocate_number($1, $2)", farmID, serie).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("failed to allocate detached number: %w", err)
	}
	return number, nil
}

// GetDetachedInvoiceTaxRates returns the tax rate overrides for a detached invoice
func (m *NFeModel) GetDetachedInvoiceTaxRates(invoiceID int) (*entity.TaxRates, error) {
	var tr entity.TaxRates
	err := m.pool.QueryRow(context.Background(),
		`SELECT icms_rate, pis_rate, cofins_rate, ibs_rate, cbs_rate
		 FROM detached_nfe_tax_rates
		 WHERE invoice_id = $1`,
		invoiceID,
	).Scan(&tr.ICMSRate, &tr.PISRate, &tr.COFINSRate, &tr.IBSRate, &tr.CBSRate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get detached invoice tax rates: %w", err)
	}
	return &tr, nil
}

// IncrementDetachedRetryCount increments the retry count for a detached invoice
func (m *NFeModel) IncrementDetachedRetryCount(id int) error {
	query := `
		UPDATE detached_nfe_invoice
		SET retry_count = retry_count + 1, last_retry_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := m.pool.Exec(context.Background(), query, id)
	return err
}

// GetPendingDetachedInvoicesForRetry returns pending detached invoices for status polling
func (m *NFeModel) GetPendingDetachedInvoicesForRetry() ([]DetachedInvoice, error) {
	query := `
		SELECT d.id, d.farm_id, d.recipient_id, d.access_key, d.serie, d.number, d.status,
			d.natureza_op, d.mod_frete,
			d.total_value, d.ibs_value, d.cbs_value,
			d.items_json,
			d.xml_signed, d.xml_authorized, d.xml_cancel_event,
			d.protocol, d.sefaz_status_code, d.sefaz_motive, d.rejection_reason, d.cancellation_reason,
			d.tp_emis, d.dh_cont, d.x_just, d.contingency_parent_id, d.svc_endpoint_used,
			d.icms_cst, d.pis_cst, d.cofins_cst, d.ibs_cst, d.cbs_cst, d.c_class_trib, d.inf_cpl,
			d.retry_count, d.last_retry_at,
			d.created_at, d.signed_at, d.sent_at, d.authorized_at, d.cancelled_at,
			t.icms_rate, t.pis_rate, t.cofins_rate, t.ibs_rate, t.cbs_rate
		FROM detached_nfe_invoice d
		LEFT JOIN detached_nfe_tax_rates t ON t.invoice_id = d.id
		WHERE d.status = 'pending'
		  AND d.retry_count < 10
		  AND (
		    d.last_retry_at IS NULL
		    OR d.last_retry_at <= CURRENT_TIMESTAMP - (INTERVAL '5 minutes' * LEAST(POWER(2, d.retry_count), 12))
		  )
		ORDER BY d.created_at ASC
		LIMIT 50
	`
	rows, err := m.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending detached invoices: %w", err)
	}
	defer rows.Close()

	var invoices []DetachedInvoice
	for rows.Next() {
		inv, scanErr := scanDetachedInvoice(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan pending detached invoice: %w", scanErr)
		}
		invoices = append(invoices, *inv)
	}

	return invoices, rows.Err()
}

// GetDraftDetachedInvoicesForRetry returns draft detached invoices eligible for auto-retry
func (m *NFeModel) GetDraftDetachedInvoicesForRetry() ([]DetachedInvoice, error) {
	query := `
		SELECT d.id, d.farm_id, d.recipient_id, d.access_key, d.serie, d.number, d.status,
			d.natureza_op, d.mod_frete,
			d.total_value, d.ibs_value, d.cbs_value,
			d.items_json,
			d.xml_signed, d.xml_authorized, d.xml_cancel_event,
			d.protocol, d.sefaz_status_code, d.sefaz_motive, d.rejection_reason, d.cancellation_reason,
			d.tp_emis, d.dh_cont, d.x_just, d.contingency_parent_id, d.svc_endpoint_used,
			d.icms_cst, d.pis_cst, d.cofins_cst, d.ibs_cst, d.cbs_cst, d.c_class_trib, d.inf_cpl,
			d.retry_count, d.last_retry_at,
			d.created_at, d.signed_at, d.sent_at, d.authorized_at, d.cancelled_at,
			t.icms_rate, t.pis_rate, t.cofins_rate, t.ibs_rate, t.cbs_rate
		FROM detached_nfe_invoice d
		LEFT JOIN detached_nfe_tax_rates t ON t.invoice_id = d.id
		WHERE d.status = 'draft'
		  AND d.retry_count < 5
		  AND d.created_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
		  AND (
		    d.last_retry_at IS NULL
		    OR d.last_retry_at <= CURRENT_TIMESTAMP - (INTERVAL '10 minutes' * LEAST(POWER(2, d.retry_count), 6))
		  )
		ORDER BY d.created_at ASC
		LIMIT 50
	`
	rows, err := m.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get draft detached invoices: %w", err)
	}
	defer rows.Close()

	var invoices []DetachedInvoice
	for rows.Next() {
		inv, scanErr := scanDetachedInvoice(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan draft detached invoice: %w", scanErr)
		}
		invoices = append(invoices, *inv)
	}

	return invoices, rows.Err()
}
