package nfe_model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// DetachedProfileItem is one ordered item default stored in a Rascunho de
// NF-e. It carries the product fields collected by the detached form plus the
// per-item CFOP and an exact unit price. Quantity and gross weight are
// deliberately not stored: they are inputs for each emission.
type DetachedProfileItem struct {
	FarmProductID *uint16         `json:"farm_product_id,omitempty"`
	ProductName   string          `json:"product_name"`
	NCM           string          `json:"ncm"`
	CEST          *string         `json:"cest,omitempty"`
	CFOP          string          `json:"cfop"`
	Unit          string          `json:"unit"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
}

// DetachedProfile is a named, farm-scoped reusable starting point for a
// detached NF-e emission. It is called "Rascunho de NF-e" in the UI. The
// internal name avoids the invoice-status word "draft" (which already drives
// GetDraftDetachedInvoicesForRetry/processDetachedDraftInvoice).
type DetachedProfile struct {
	ID          int
	FarmID      uint32
	Name        string
	RecipientID *uint32
	NaturezaOp  *string
	ModFrete    *int
	InfCpl      *string
	Items       []DetachedProfileItem
	ICMSCST     *string
	PISCST      *string
	COFINSCST   *string
	IBSCST      *string
	CBSCST      *string
	CClassTrib  *string
	TaxRates    *entity.TaxRates
	CreatedAt   interface{}
	UpdatedAt   interface{}

	// RecipientName is display-only, resolved by the list/get queries.
	RecipientName *string
}

// ErrDetachedProfileNameTaken is returned when the farm already has a
// rascunho with the same name (case-insensitive).
var ErrDetachedProfileNameTaken = errors.New("detached profile name already taken")

// ErrDetachedProfileNotFound is returned when a rascunho does not exist for
// the requested farm.
var ErrDetachedProfileNotFound = errors.New("detached profile not found")

// isUniqueViolation reports whether err is a PostgreSQL unique constraint
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

// CreateDetachedProfile inserts a new rascunho. The unique index on
// (farm_id, LOWER(name)) rejects duplicate names case-insensitively and is
// translated into ErrDetachedProfileNameTaken.
func (m *NFeModel) CreateDetachedProfile(profile DetachedProfile) (int, error) {
	itemsJSON, err := json.Marshal(profile.Items)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal profile items: %w", err)
	}

	query := `
		INSERT INTO detached_nfe_profile (
			farm_id, name, recipient_id,
			natureza_op, mod_frete, inf_cpl,
			items_json,
			icms_cst, pis_cst, cofins_cst, ibs_cst, cbs_cst, c_class_trib,
			icms_rate, pis_rate, cofins_rate, ibs_rate, cbs_rate
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18)
		RETURNING id
	`
	var id int
	err = m.pool.QueryRow(context.Background(), query,
		profile.FarmID, strings.TrimSpace(profile.Name), profile.RecipientID,
		profile.NaturezaOp, profile.ModFrete, profile.InfCpl,
		itemsJSON,
		profile.ICMSCST, profile.PISCST, profile.COFINSCST,
		profile.IBSCST, profile.CBSCST, profile.CClassTrib,
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.ICMSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.PISRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.COFINSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.IBSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.CBSRate }),
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrDetachedProfileNameTaken
		}
		return 0, fmt.Errorf("failed to create detached profile: %w", err)
	}
	return id, nil
}

// UpdateDetachedProfile replaces every editable field of a farm-scoped
// rascunho. It returns pgx.ErrNoRows when the rascunho does not exist or
// belongs to another farm, so callers can surface "Rascunho não encontrado".
func (m *NFeModel) UpdateDetachedProfile(profile DetachedProfile) error {
	itemsJSON, err := json.Marshal(profile.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal profile items: %w", err)
	}

	query := `
		UPDATE detached_nfe_profile
		SET name = $3,
		    recipient_id = $4,
		    natureza_op = $5,
		    mod_frete = $6,
		    inf_cpl = $7,
		    items_json = $8,
		    icms_cst = $9, pis_cst = $10, cofins_cst = $11,
		    ibs_cst = $12, cbs_cst = $13, c_class_trib = $14,
		    icms_rate = $15, pis_rate = $16, cofins_rate = $17,
		    ibs_rate = $18, cbs_rate = $19,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND farm_id = $2
	`
	tag, err := m.pool.Exec(context.Background(), query,
		profile.ID, profile.FarmID,
		strings.TrimSpace(profile.Name), profile.RecipientID,
		profile.NaturezaOp, profile.ModFrete, profile.InfCpl,
		itemsJSON,
		profile.ICMSCST, profile.PISCST, profile.COFINSCST,
		profile.IBSCST, profile.CBSCST, profile.CClassTrib,
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.ICMSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.PISRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.COFINSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.IBSRate }),
		rateOrNil(profile.TaxRates, func(t *entity.TaxRates) *decimal.Decimal { return t.CBSRate }),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDetachedProfileNameTaken
		}
		return fmt.Errorf("failed to update detached profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDetachedProfileNotFound
	}
	return nil
}

// DeleteDetachedProfile hard-deletes a farm-scoped rascunho. It reports
// whether a row was removed; false means the id does not exist for that farm.
func (m *NFeModel) DeleteDetachedProfile(farmID uint32, id int) (bool, error) {
	tag, err := m.pool.Exec(context.Background(),
		`DELETE FROM detached_nfe_profile WHERE id = $1 AND farm_id = $2`, id, farmID)
	if err != nil {
		return false, fmt.Errorf("failed to delete detached profile: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

const detachedProfileSelect = `
	SELECT p.id, p.farm_id, p.name, p.recipient_id,
		p.natureza_op, p.mod_frete, p.inf_cpl,
		p.items_json,
		p.icms_cst, p.pis_cst, p.cofins_cst, p.ibs_cst, p.cbs_cst, p.c_class_trib,
		p.icms_rate, p.pis_rate, p.cofins_rate, p.ibs_rate, p.cbs_rate,
		p.created_at, p.updated_at,
		r.recipient_name
	FROM detached_nfe_profile p
	LEFT JOIN LATERAL (
		SELECT COALESCE(lp.companyname, np.name) AS recipient_name
		FROM person pe
		LEFT JOIN legal_person lp ON lp.personid = pe.id
		LEFT JOIN natural_person np ON np.personid = pe.id
		WHERE pe.id = p.recipient_id
	) r ON TRUE
`

// GetDetachedProfile returns a single farm-scoped rascunho. Returns (nil, nil)
// when the id does not exist for that farm.
func (m *NFeModel) GetDetachedProfile(farmID uint32, id int) (*DetachedProfile, error) {
	query := detachedProfileSelect + ` WHERE p.id = $1 AND p.farm_id = $2`
	row := m.pool.QueryRow(context.Background(), query, id, farmID)
	profile, err := scanDetachedProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get detached profile: %w", err)
	}
	return profile, nil
}

// GetDetachedProfilesByFarm lists every rascunho of a farm ordered by name.
func (m *NFeModel) GetDetachedProfilesByFarm(farmID uint32) ([]DetachedProfile, error) {
	query := detachedProfileSelect + ` WHERE p.farm_id = $1 ORDER BY LOWER(p.name) ASC, p.id ASC`
	rows, err := m.pool.Query(context.Background(), query, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to list detached profiles: %w", err)
	}
	defer rows.Close()

	var profiles []DetachedProfile
	for rows.Next() {
		profile, scanErr := scanDetachedProfile(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan detached profile: %w", scanErr)
		}
		profiles = append(profiles, *profile)
	}
	return profiles, rows.Err()
}

// detachedProfileScanner is the common surface for pgx.Row and pgx.Rows.
type detachedProfileScanner interface {
	Scan(dest ...any) error
}

// scanDetachedProfile scans a detachedProfileSelect row.
func scanDetachedProfile(row detachedProfileScanner) (*DetachedProfile, error) {
	var profile DetachedProfile
	var naturezaOp, infCpl, icmsCST, pisCST, cofinsCST, ibsCST, cbsCST, cClassTrib *string
	var modFrete *int
	var recipientID *uint32
	var itemsJSON []byte
	var icmsRate, pisRate, cofinsRate, ibsRate, cbsRate *decimal.Decimal
	var recipientName *string

	err := row.Scan(
		&profile.ID, &profile.FarmID, &profile.Name, &recipientID,
		&naturezaOp, &modFrete, &infCpl,
		&itemsJSON,
		&icmsCST, &pisCST, &cofinsCST, &ibsCST, &cbsCST, &cClassTrib,
		&icmsRate, &pisRate, &cofinsRate, &ibsRate, &cbsRate,
		&profile.CreatedAt, &profile.UpdatedAt,
		&recipientName,
	)
	if err != nil {
		return nil, err
	}

	profile.RecipientID = recipientID
	profile.NaturezaOp = naturezaOp
	profile.ModFrete = modFrete
	profile.InfCpl = infCpl
	profile.ICMSCST = icmsCST
	profile.PISCST = pisCST
	profile.COFINSCST = cofinsCST
	profile.IBSCST = ibsCST
	profile.CBSCST = cbsCST
	profile.CClassTrib = cClassTrib
	profile.RecipientName = recipientName

	if itemsJSON != nil {
		if err := json.Unmarshal(itemsJSON, &profile.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile items: %w", err)
		}
	}

	if icmsRate != nil || pisRate != nil || cofinsRate != nil || ibsRate != nil || cbsRate != nil {
		profile.TaxRates = &entity.TaxRates{
			ICMSRate:   icmsRate,
			PISRate:    pisRate,
			COFINSRate: cofinsRate,
			IBSRate:    ibsRate,
			CBSRate:    cbsRate,
		}
	}

	return &profile, nil
}

// ValidateDetachedProfileReferences checks that the optional recipient and
// every optional farm-product reference belong to the given farm. It is used
// server-side so a rascunho can never persist a reference from another farm.
func (m *NFeModel) ValidateDetachedProfileReferences(farmID uint32, recipientID *uint32, farmProductIDs []uint16) error {
	if recipientID != nil {
		var count int
		err := m.pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM person WHERE id = $1 AND farm = $2`,
			*recipientID, farmID,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to validate profile recipient: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("recipient %d does not belong to farm %d", *recipientID, farmID)
		}
	}

	if len(farmProductIDs) == 0 {
		return nil
	}

	// int32 keeps the comparison unambiguous regardless of the smallint
	// column type.
	ids := make([]int32, 0, len(farmProductIDs))
	for _, id := range uniqueUint16(farmProductIDs) {
		ids = append(ids, int32(id))
	}

	var count int
	err := m.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM farm_product WHERE farm_id = $1 AND id = ANY($2)`,
		farmID, ids,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate profile farm products: %w", err)
	}
	if count != len(ids) {
		return fmt.Errorf("one or more profile farm products do not belong to farm %d", farmID)
	}
	return nil
}

// uniqueUint16 returns the distinct values of ids.
func uniqueUint16(ids []uint16) []uint16 {
	seen := make(map[uint16]struct{}, len(ids))
	out := make([]uint16, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// rateOrNil returns the decimal rate from rates via getter, or nil when rates
// itself is nil, preserving "unset" (NULL) versus explicit zero.
func rateOrNil(rates *entity.TaxRates, getter func(*entity.TaxRates) *decimal.Decimal) *decimal.Decimal {
	if rates == nil {
		return nil
	}
	return getter(rates)
}

// DetachedProfileRecipient displays the recipient name, falling back to nil.
func (p *DetachedProfile) RecipientLabel() string {
	if p.RecipientName != nil && *p.RecipientName != "" {
		return *p.RecipientName
	}
	return ""
}

// FarmProductIDsFromProfileItems collects the distinct farm-product references
// of a profile's items for the farm-scope validation.
func FarmProductIDsFromProfileItems(items []DetachedProfileItem) []uint16 {
	ids := make([]uint16, 0, len(items))
	for _, item := range items {
		if item.FarmProductID != nil {
			ids = append(ids, *item.FarmProductID)
		}
	}
	return uniqueUint16(ids)
}
