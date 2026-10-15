package nfe_model_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func TestInvoice_EmissionInputs_DBRoundTrip(t *testing.T) {
	pool, ctx, cancel := dbRoundTripPool(t)
	defer cancel()
	defer pool.Close()

	var farmID, productID, cropID, vehicleID, departureID int
	if err := pool.QueryRow(ctx, `SELECT MIN(id) FROM farm`).Scan(&farmID); err != nil || farmID == 0 {
		t.Fatalf("no farm rows available for the round trip: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO product (name) VALUES (@name) RETURNING id`,
		pgx.NamedArgs{"name": fmt.Sprintf("emission-inputs-product-%d", time.Now().UnixNano())}).Scan(&productID); err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM product WHERE id = @id`, pgx.NamedArgs{"id": productID})
	})
	if err := pool.QueryRow(ctx,
		`INSERT INTO crop (name, product, startDate, farm) VALUES (@name, @product, CURRENT_DATE, @farm) RETURNING id`,
		pgx.NamedArgs{"name": fmt.Sprintf("emission-inputs-crop-%d", time.Now().UnixNano()), "product": productID, "farm": farmID}).Scan(&cropID); err != nil {
		t.Fatalf("failed to seed crop: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM crop WHERE id = @id`, pgx.NamedArgs{"id": cropID})
	})
	if err := pool.QueryRow(ctx,
		`INSERT INTO vehicle (plate, name, farm) VALUES (@plate, @name, @farm) RETURNING id`,
		pgx.NamedArgs{"plate": fmt.Sprintf("EIN%d", time.Now().UnixNano()%100000), "name": "emission inputs test", "farm": farmID}).Scan(&vehicleID); err != nil {
		t.Fatalf("failed to seed vehicle: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM vehicle WHERE id = @id`, pgx.NamedArgs{"id": vehicleID})
	})
	if err := pool.QueryRow(ctx,
		`INSERT INTO departure (departureDate, vehicle, crop, grossWeight, tare, netWeight, farm)
		 VALUES (CURRENT_TIMESTAMP, @vehicle, @crop, 110, 10, 100, @farm) RETURNING id`,
		pgx.NamedArgs{"vehicle": vehicleID, "crop": cropID, "farm": farmID}).Scan(&departureID); err != nil {
		t.Fatalf("failed to seed departure: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM departure WHERE id = @id`, pgx.NamedArgs{"id": departureID})
	})

	model, err := nfe_model.InitNFeModel(pool)
	if err != nil {
		t.Fatalf("InitNFeModel failed: %v", err)
	}
	accessKey := fmt.Sprintf("emission-inputs-%d", time.Now().UnixNano())
	want := &entity.InvoiceEmissionInputs{
		Payment:         entity.PaymentData{Detalhes: []entity.PagamentoDetalhe{{IndPag: 1, TPag: "01", VPag: decimal.RequireFromString("90.00")}}},
		IncludeCobranca: false,
		Transport: entity.TransportData{
			ModFrete:       1,
			Transportadora: &entity.TransportadoraData{Type: 2, CPF: "52998224725", XNome: "Transportador Teste", Endereco: "Rodovia MT-100, Km 10", Municipio: "Cuiaba", UF: "MT"},
			Veiculo:        &entity.VeiculoData{Placa: "ABC1D23", UF: "MT", RNTC: "12345"},
			Volumes:        []entity.VolumeData{{QVol: 2, Esp: "Granel", PesoL: decimal.NewFromInt(100), PesoB: decimal.NewFromInt(110)}},
		},
	}
	invoiceID, err := model.CreateInvoiceWithEmission(
		uint32(departureID), accessKey, 920, 1, "5101", "10059010",
		decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.NewFromInt(100),
		decimal.Zero, decimal.Zero, 1, nil, "", nil, snapshotCodes(), want,
	)
	if err != nil {
		t.Fatalf("CreateInvoiceWithEmission failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": invoiceID})
	})

	loaded, err := model.GetInvoiceByAccessKey(accessKey)
	if err != nil {
		t.Fatalf("GetInvoiceByAccessKey failed: %v", err)
	}
	if loaded == nil || loaded.EmissionInputs == nil {
		t.Fatal("expected emission inputs to load from JSONB snapshot")
	}
	got := loaded.EmissionInputs
	if got.Payment.Detalhes[0].TPag != "01" || got.Payment.Detalhes[0].IndPag != 1 {
		t.Errorf("payment detail = %+v, want tPag 01 / indPag 1", got.Payment.Detalhes[0])
	}
	if got.IncludeCobranca {
		t.Error("IncludeCobranca = true, want false")
	}
	if got.Transport.ModFrete != 1 || got.Transport.Transportadora == nil || got.Transport.Transportadora.CPF != "52998224725" {
		t.Errorf("transportadora snapshot not preserved: %+v", got.Transport.Transportadora)
	}
	if got.Transport.Veiculo == nil || got.Transport.Veiculo.RNTC != "12345" || len(got.Transport.Volumes) != 1 || got.Transport.Volumes[0].QVol != 2 {
		t.Errorf("vehicle/volume snapshot not preserved: %+v", got.Transport)
	}
	byDeparture, err := model.GetInvoiceByDeparture(uint32(departureID))
	if err != nil || byDeparture == nil || byDeparture.EmissionInputs == nil {
		t.Fatalf("GetInvoiceByDeparture must scan the emission snapshot: invoice=%+v err=%v", byDeparture, err)
	}
	listed, _, err := model.GetInvoicesByFarm(uint32(farmID), 1)
	if err != nil {
		t.Fatalf("GetInvoicesByFarm failed to scan the emission snapshot: %v", err)
	}
	foundSnapshot := false
	for _, invoice := range listed {
		if invoice.ID == invoiceID {
			foundSnapshot = invoice.EmissionInputs != nil
			break
		}
	}
	if !foundSnapshot {
		t.Error("GetInvoicesByFarm did not return the invoice emission snapshot")
	}

	legacyKey := fmt.Sprintf("emission-inputs-legacy-%d", time.Now().UnixNano())
	legacyID, err := model.CreateInvoiceWithEmission(
		uint32(departureID), legacyKey, 920, 2, "5101", "10059010",
		decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.NewFromInt(100),
		decimal.Zero, decimal.Zero, 1, nil, "", nil, snapshotCodes(), nil,
	)
	if err != nil {
		t.Fatalf("CreateInvoiceWithEmission legacy row failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM nfe_invoice WHERE id = @id`, pgx.NamedArgs{"id": legacyID})
	})
	legacy, err := model.GetInvoiceByAccessKey(legacyKey)
	if err != nil {
		t.Fatalf("GetInvoiceByAccessKey legacy row failed: %v", err)
	}
	if legacy == nil || legacy.EmissionInputs != nil {
		t.Errorf("empty legacy JSON snapshot should load as nil, got %+v", legacy)
	}
}
