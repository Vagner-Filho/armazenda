package nfe_service

import (
	"strings"
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

func boolRef(v bool) *bool { return &v }

func intRef(v int) *int { return &v }

func decimalRef(v string) *decimal.Decimal {
	d := decimal.RequireFromString(v)
	return &d
}

func TestDeparturePaymentAndBillingDefaults(t *testing.T) {
	total := decimal.RequireFromString("25000.00")
	defaultsPayment := departurePayment(nil, total)
	if len(defaultsPayment.Detalhes) != 1 {
		t.Fatalf("default payment details = %d, want one", len(defaultsPayment.Detalhes))
	}
	if got := defaultsPayment.Detalhes[0]; got.TPag != "90" || got.IndPag != 0 || !got.VPag.Equal(total) {
		t.Errorf("legacy payment default = %+v, want tPag 90 / indPag 0 / total value", got)
	}
	if !departureCobrancaEnabled(nil) {
		t.Error("legacy nil billing option must keep cobr/fat enabled")
	}

	options := &entity.DepartureInvoiceOverrides{
		TPag:            "01",
		IndPag:          1,
		IncludeCobranca: boolRef(false),
	}
	selected := departurePayment(options, total)
	if got := selected.Detalhes[0]; got.TPag != "01" || got.IndPag != 1 || !got.VPag.Equal(total) {
		t.Errorf("selected payment = %+v, want tPag 01 / indPag 1 / total value", got)
	}
	if departureCobrancaEnabled(options) {
		t.Error("explicitly disabled billing must stay disabled")
	}
	if errs := validateDeparturePayment(options); len(errs) != 0 {
		t.Errorf("valid payment options returned errors: %v", errs)
	}
	invalid := &entity.DepartureInvoiceOverrides{TPag: "03", IndPag: 3}
	if errs := validateDeparturePayment(invalid); len(errs) != 2 {
		t.Errorf("invalid payment options errors = %v, want tPag and indPag errors", errs)
	}
}

func TestApplyInvoiceEmissionInputsReusesResolvedSnapshot(t *testing.T) {
	input := entity.InvoiceInput{
		Payment:  defaultPayment("90", decimal.NewFromInt(200)),
		Cobranca: &entity.CobrancaData{Enabled: true},
		Transport: entity.TransportData{
			ModFrete: 9,
			Veiculo:  &entity.VeiculoData{Placa: "CURRENT"},
			Volumes:  []entity.VolumeData{{QVol: 1, Esp: "Current"}},
		},
	}
	snapshot := &entity.InvoiceEmissionInputs{
		Payment:         entity.PaymentData{Detalhes: []entity.PagamentoDetalhe{{IndPag: 1, TPag: "01", VPag: decimal.NewFromInt(200)}}},
		IncludeCobranca: false,
		Transport: entity.TransportData{
			ModFrete: 1,
			Veiculo:  &entity.VeiculoData{Placa: "ORIGINAL"},
			Volumes:  nil,
		},
	}

	applyInvoiceEmissionInputs(&input, snapshot)
	if input.Payment.Detalhes[0].TPag != "01" || input.Payment.Detalhes[0].IndPag != 1 {
		t.Errorf("payment was not restored from snapshot: %+v", input.Payment)
	}
	if input.Cobranca == nil || input.Cobranca.Enabled {
		t.Errorf("billing toggle was not restored: %+v", input.Cobranca)
	}
	if input.Transport.ModFrete != 1 || input.Transport.Veiculo == nil || input.Transport.Veiculo.Placa != "ORIGINAL" || len(input.Transport.Volumes) != 0 {
		t.Errorf("transport was re-derived instead of restored: %+v", input.Transport)
	}
}

func TestDepartureTaxCodeOptionsAndValidation(t *testing.T) {
	if options := DepartureICMSOptions(defaults.TaxRegimeLucroReal); len(options) != 7 || options[5].Value != "51" {
		t.Errorf("normal-regime CST options = %+v", options)
	}
	if options := DepartureICMSOptions(defaults.TaxRegimeSimplesNacional); len(options) != 10 || options[0].Value != "101" {
		t.Errorf("Simples CSOSN options = %+v", options)
	}
	if options := DeparturePISCOFINSOptions(); len(options) != 8 || options[6].Value != "08" {
		t.Errorf("PIS/COFINS CST options = %+v", options)
	}

	validNormal := entity.InvoiceOverrides{ICMSCST: stringRef("51"), PISCST: stringRef("08"), COFINSCST: stringRef("08")}
	if errs := validateDepartureTaxCodes(validNormal, defaults.TaxRegimeLucroReal); len(errs) != 0 {
		t.Errorf("valid normal-regime codes returned errors: %v", errs)
	}
	validSimple := entity.InvoiceOverrides{ICMSCST: stringRef("102"), PISCST: stringRef("08"), COFINSCST: stringRef("08")}
	if errs := validateDepartureTaxCodes(validSimple, defaults.TaxRegimeSimplesNacional); len(errs) != 0 {
		t.Errorf("valid Simples codes returned errors: %v", errs)
	}
	invalid := entity.InvoiceOverrides{ICMSCST: stringRef("03"), PISCST: stringRef("03"), COFINSCST: stringRef("10")}
	if errs := validateDepartureTaxCodes(invalid, defaults.TaxRegimeLucroReal); len(errs) != 3 {
		t.Errorf("invalid tax codes errors = %v, want three errors", errs)
	}
}

func TestMergeDepartureTransportDefaultsAndExplicitOmission(t *testing.T) {
	departure := entity_public.Departure{
		CargoWeight: entity_public.CargoWeight{
			NetWeight:   decimal.NewFromInt(100),
			GrossWeight: decimal.NewFromInt(110),
		},
	}
	base := entity.TransportData{
		ModFrete: 1,
		Veiculo:  &entity.VeiculoData{Placa: "ABC1234", UF: "MT"},
		Volumes:  []entity.VolumeData{{QVol: 1, Esp: "Granel", PesoL: departure.NetWeight, PesoB: departure.GrossWeight}},
	}

	legacy, errs := mergeDepartureTransport(base, nil, departure, "MT", "MT")
	if len(errs) != 0 || legacy.Veiculo == nil || len(legacy.Volumes) != 1 {
		t.Fatalf("legacy transport defaults changed: transport=%+v errors=%v", legacy, errs)
	}

	document := "529.982.247-25"
	form := &entity.DepartureTransportInput{
		TransporterName:     "Transportador de Teste",
		TransporterDocument: document,
		TransporterStreet:   "Rodovia MT-100",
		TransporterNumber:   "Km 10",
		TransporterCity:     "Cuiabá",
		TransporterUF:       "MT",
		IncludeVehicle:      boolRef(false),
		IncludeVolumes:      boolRef(false),
	}
	merged, errs := mergeDepartureTransport(base, form, departure, "MT", "MT")
	if len(errs) != 0 {
		t.Fatalf("valid transporter returned errors: %v", errs)
	}
	if merged.Transportadora == nil || merged.Transportadora.Type != 2 || merged.Transportadora.CPF != "52998224725" {
		t.Errorf("transporter CPF mapping = %+v", merged.Transportadora)
	}
	if merged.Transportadora.Endereco != "Rodovia MT-100, Km 10" {
		t.Errorf("transporter xEnder = %q, want combined street/number", merged.Transportadora.Endereco)
	}
	if merged.Veiculo != nil || len(merged.Volumes) != 0 {
		t.Errorf("explicitly omitted groups were restored: vehicle=%+v volumes=%+v", merged.Veiculo, merged.Volumes)
	}
}

func TestMergeDepartureTransportUserOverridesAndRejectsInterstatePlate(t *testing.T) {
	departure := entity_public.Departure{CargoWeight: entity_public.CargoWeight{NetWeight: decimal.NewFromInt(100), GrossWeight: decimal.NewFromInt(110)}}
	base := entity.TransportData{
		ModFrete: 1,
		Veiculo:  &entity.VeiculoData{Placa: "ABC1234", UF: "MT"},
		Volumes:  []entity.VolumeData{{QVol: 1, Esp: "Granel", PesoL: departure.NetWeight, PesoB: departure.GrossWeight}},
	}
	form := &entity.DepartureTransportInput{
		VehiclePlate:   "abc-1d23",
		VehicleUF:      "MT",
		VehicleRNTC:    "12345",
		IncludeVehicle: boolRef(true),
		IncludeVolumes: boolRef(true),
		VolumeQuantity: intRef(2),
		VolumeSpecies:  "Sacas",
		VolumeBrand:    "Armazenda",
		VolumeNet:      decimalRef("95.5"),
		VolumeGross:    decimalRef("100.25"),
	}
	merged, errs := mergeDepartureTransport(base, form, departure, "MT", "MT")
	if len(errs) != 0 {
		t.Fatalf("valid override returned errors: %v", errs)
	}
	if merged.Veiculo == nil || merged.Veiculo.Placa != "ABC1D23" || merged.Veiculo.RNTC != "12345" {
		t.Errorf("vehicle override = %+v", merged.Veiculo)
	}
	if len(merged.Volumes) != 1 || merged.Volumes[0].QVol != 2 || merged.Volumes[0].Esp != "Sacas" || !merged.Volumes[0].PesoL.Equal(decimal.RequireFromString("95.5")) {
		t.Errorf("volume override = %+v", merged.Volumes)
	}

	_, errs = mergeDepartureTransport(base, form, departure, "MT", "GO")
	if len(errs) == 0 || !strings.Contains(strings.Join(errs, ";"), "interestadual") {
		t.Errorf("interstate vehicle should be rejected, got %v", errs)
	}
}

func stringRef(v string) *string { return &v }
