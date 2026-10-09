package nfe_model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// DetachedTransportadora is the transporta group (MOC 7.0, X03/X04–X10) of a
// detached NF-e transport document. TpDoc selects the document tag:
// 1=CNPJ (X04), 2=CPF (X05). TpDoc mirrors the builder's TransportadoraData
// Type convention.
type DetachedTransportadora struct {
	TpDoc     int    `json:"tpDoc"` // 1=CNPJ, 2=CPF
	DocNumber string `json:"docNumber"`
	XNome     string `json:"xNome"`
	IE        string `json:"ie,omitempty"`
	XEnder    string `json:"xEnder,omitempty"`
	XMun      string `json:"xMun,omitempty"`
	UF        string `json:"uf,omitempty"`
}

// DetachedVeiculo is the veicTransp group (X18–X21) of a detached NF-e
// transport document: plate, plate UF and the ANTT registry (RNTC).
type DetachedVeiculo struct {
	Placa string `json:"placa"`
	UF    string `json:"uf,omitempty"`
	RNTC  string `json:"rntc,omitempty"`
}

// DetachedVol is the single volume block (X26–X32) of a detached NF-e
// transport document.
type DetachedVol struct {
	Qtd   int             `json:"qtd"`
	Esp   string          `json:"esp,omitempty"`
	Marca string          `json:"marca,omitempty"`
	PesoL decimal.Decimal `json:"pesoL"`
	PesoB decimal.Decimal `json:"pesoB"`
}

// DetachedInvoiceTransport is the transport document (MOC 7.0, Grupo X) of
// one detached NF-e, persisted as a single JSONB document on
// detached_invoice_transportation (same pattern as items_json). Each group is
// optional; a group is present only when at least one of its fields was
// informed. The document stores the resolved values actually emitted.
type DetachedInvoiceTransport struct {
	Transportadora *DetachedTransportadora `json:"transportadora,omitempty"`
	Veiculo        *DetachedVeiculo        `json:"veiculo,omitempty"`
	Vol            *DetachedVol            `json:"vol,omitempty"`
}

// CreateDetachedInvoiceTransport stores the transport JSONB document for a
// detached invoice. The UNIQUE(invoice_id) constraint rejects duplicates;
// each invoice is persisted exactly once.
func (m *NFeModel) CreateDetachedInvoiceTransport(invoiceID int, t DetachedInvoiceTransport) error {
	payload, marshalErr := json.Marshal(t)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal detached transport: %w", marshalErr)
	}

	var id int
	queryErr := m.pool.QueryRow(context.Background(),
		`INSERT INTO detached_invoice_transportation (invoice_id, transport)
		 VALUES (@invoice_id, @transport)
		 RETURNING id`,
		pgx.NamedArgs{"invoice_id": invoiceID, "transport": payload},
	).Scan(&id)
	if queryErr != nil {
		return fmt.Errorf("failed to create detached invoice transport: %w", queryErr)
	}
	return nil
}

// GetDetachedInvoiceTransport returns the transport document stored for a
// detached invoice, or nil when the invoice carries no transport data.
func (m *NFeModel) GetDetachedInvoiceTransport(invoiceID int) (*DetachedInvoiceTransport, error) {
	var payload []byte
	err := m.pool.QueryRow(context.Background(),
		`SELECT transport
		 FROM detached_invoice_transportation
		 WHERE invoice_id = @invoice_id
		 LIMIT 1`,
		pgx.NamedArgs{"invoice_id": invoiceID},
	).Scan(&payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get detached invoice transport: %w", err)
	}

	var t DetachedInvoiceTransport
	if unmarshalErr := json.Unmarshal(payload, &t); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal detached invoice transport: %w", unmarshalErr)
	}
	return &t, nil
}
