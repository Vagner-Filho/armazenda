package nfe_service

import (
	"strings"
	"testing"

	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

func validTransportInput() DetachedTransportInput {
	nome := "Carreteiro Fulano"
	cpf := "529.982.247-25"
	uf := "MT"
	return DetachedTransportInput{
		Nome:     &nome,
		Document: &cpf,
		UF:       &uf,
	}
}

func TestIsValidTransportCPF(t *testing.T) {
	if !isValidCPF("52998224725") {
		t.Error("52998224725 must be a valid CPF")
	}
	if isValidCPF("52998224724") {
		t.Error("52998224724 must be an invalid CPF")
	}
	if isValidCPF("11111111111") {
		t.Error("repeated-digit CPFs must be rejected")
	}
	if isValidCPF("529.982.247-25") {
		t.Error("punctuation must not be accepted uncleaned")
	}
}

func TestIsValidTransportCNPJ(t *testing.T) {
	if !isValidCNPJ("11222333000181") {
		t.Error("11222333000181 must be a valid CNPJ")
	}
	if isValidCNPJ("11222333000180") {
		t.Error("11222333000180 must be an invalid CNPJ")
	}
	if isValidCNPJ("00000000000000") {
		t.Error("all-zero CNPJ must be rejected")
	}
}

func TestIsValidDetachedPlaca(t *testing.T) {
	cases := []struct {
		placa string
		valid bool
	}{
		{"ABC-1234", true}, // legacy XXX9999, separator stripped
		{"ABC1234", true},  // legacy XXX9999
		{"ABC123", true},   // legacy XXX999
		{"AB1234", true},   // legacy XX9999
		{"ABCD123", true},  // legacy XXXX999 (trailer)
		{"ABC1D23", true},  // MERCOSUL AAA9A99
		{"abc1d23", true},  // case-insensitive input (normalized)
		{"A1B2C3", false},  // mixed nonsense
		{"ABCD12", false},  // 4 letters + 2 digits
		{"1234567", false}, // digits only
		{"ABCDE12", false}, // 5 letters + 2 digits
		{"", false},
	}
	for _, tc := range cases {
		normalized := normalizePlaca(tc.placa)
		if got := isValidPlaca(normalized); got != tc.valid {
			t.Errorf("isValidPlaca(%q) = %v, want %v", tc.placa, got, tc.valid)
		}
	}
}

func TestValidateDetachedTransport_BlankInputIsAllowed(t *testing.T) {
	errs := validateDetachedTransport(DetachedTransportInput{}, "MT", "MT")
	if len(errs) > 0 {
		t.Errorf("blank transport must be allowed, got %v", errs)
	}
}

func TestValidateDetachedTransport_HappyPathNoErrors(t *testing.T) {
	in := validTransportInput()
	if errs := validateDetachedTransport(in, "MT", "MT"); len(errs) > 0 {
		t.Errorf("valid transport produced errors: %v", errs)
	}
}

func TestValidateDetachedTransport_DocumentRequiredWithName(t *testing.T) {
	nome := "Carreteiro"
	in := DetachedTransportInput{Nome: &nome}
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "CNPJ ou CPF do transportador") {
		t.Errorf("expected missing-document error, got %v", errs)
	}
}

func TestValidateDetachedTransport_BadCheckDigit(t *testing.T) {
	doc := "52998224724"
	in := validTransportInput()
	in.Document = &doc
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "CPF do transportador inválido") {
		t.Errorf("expected invalid CPF error, got %v", errs)
	}
}

func TestValidateDetachedTransport_WrongDocumentLength(t *testing.T) {
	doc := "1234567891011"
	in := validTransportInput()
	in.Document = &doc
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "11 dígitos (CPF) ou 14 (CNPJ)") {
		t.Errorf("expected document length error, got %v", errs)
	}
}

func TestValidateDetachedTransport_IEWithoutUF(t *testing.T) {
	ie := "ISENTO"
	in := validTransportInput()
	in.IE = &ie
	in.UF = nil // IE informed without the transportadora UF must be rejected
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "UF do transportador") {
		t.Errorf("expected IE-without-UF error (Rej. 559), got %v", errs)
	}
}

func TestValidateDetachedTransport_XEnderOverSixty(t *testing.T) {
	street := strings.Repeat("R", 58)
	number := "123"
	in := validTransportInput()
	in.Logradouro = &street
	in.Numero = &number
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "excede 60 caracteres") {
		t.Errorf("expected length error for xEnder, got %v", errs)
	}
}

func TestValidateDetachedTransport_RNTCOverTwenty(t *testing.T) {
	rntc := "123456789012345678901" // 21 chars
	in := validTransportInput()
	in.RNTC = &rntc
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "ANTT (RNTRC)") {
		t.Errorf("expected RNTC length error, got %v", errs)
	}
}

func TestValidateDetachedTransport_VolumeQtdMustBePositive(t *testing.T) {
	qtd := 0
	in := validTransportInput()
	in.VolQtd = &qtd
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "Quantidade de volumes") {
		t.Errorf("expected volume quantity error, got %v", errs)
	}
}

func TestValidateDetachedTransport_NegativeWeightsRejected(t *testing.T) {
	peso := decimal.RequireFromString("-1")
	in := validTransportInput()
	in.VolPesoL = &peso
	errs := validateDetachedTransport(in, "MT", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "negativos") {
		t.Errorf("expected negative weight error, got %v", errs)
	}
}

func TestValidateDetachedTransport_InterstateVehicleRejected(t *testing.T) {
	placa := "ABC1234"
	in := validTransportInput()
	in.Placa = &placa
	errs := validateDetachedTransport(in, "SP", "MT")
	if !strings.Contains(strings.Join(errs, "; "), "operacão interestadual") &&
		!strings.Contains(strings.Join(errs, "; "), "interestadual") {
		t.Errorf("expected interstate guard error (Rej. 868), got %v", errs)
	}
}

func TestValidateDetachedTransport_SameStateVehicleAllowed(t *testing.T) {
	placa := "ABC1234"
	in := validTransportInput()
	in.Placa = &placa
	in.PlacaUF = func() *string { s := "MT"; return &s }()
	if errs := validateDetachedTransport(in, "MT", "MT"); len(errs) > 0 {
		t.Errorf("same-state vehicle must be allowed, got %v", errs)
	}
}

func TestValidateDetachedTransport_UnknownDestKUFAllowsVehicle(t *testing.T) {
	// Unknown destination UF cannot trigger the interstate guard.
	placa := "ABC1234"
	in := validTransportInput()
	in.Placa = &placa
	in.PlacaUF = func() *string { s := "MT"; return &s }()
	if errs := validateDetachedTransport(in, "", "MT"); len(errs) > 0 {
		t.Errorf("vehicle with unknown dest UF must be allowed, got %v", errs)
	}
}

func TestTransportadorXender_Concatenation(t *testing.T) {
	street := "Rodovia BR-163"
	number := "Km 12"
	in := DetachedTransportInput{Logradouro: &street, Numero: &number}
	if got := transportadorXender(in); got != "Rodovia BR-163, Km 12" {
		t.Errorf("expected 'Rodovia BR-163, Km 12', got %q", got)
	}

	if got := transportadorXender(DetachedTransportInput{Logradouro: &street}); got != street {
		t.Errorf("street-only must stay as-is, got %q", got)
	}
	if got := transportadorXender(DetachedTransportInput{Numero: &number}); got != number {
		t.Errorf("number-only must stay as-is, got %q", got)
	}
}

func TestMapDetachedTransport_BlankKeepsAutoDerivation(t *testing.T) {
	totalNet := decimal.RequireFromString("800")
	totalGross := decimal.RequireFromString("820")

	transport := mapDetachedTransport(DetachedTransportInput{}, 1, "MT", totalNet, totalGross)

	if transport.ModFrete != 1 {
		t.Errorf("expected modFrete 1, got %d", transport.ModFrete)
	}
	if transport.Transportadora != nil {
		t.Error("blank input must not create a transportadora")
	}
	if transport.Veiculo != nil {
		t.Error("blank input must not create a vehicle")
	}
	if len(transport.Volumes) != 1 {
		t.Fatalf("expected the auto volume block, got %v", transport.Volumes)
	}
	vol := transport.Volumes[0]
	if vol.QVol != 1 || vol.Esp != "Granel" {
		t.Errorf("auto block must be qVol=1/Granel, got %+v", vol)
	}
	if !vol.PesoL.Equal(totalNet) || !vol.PesoB.Equal(totalGross) {
		t.Errorf("auto block weights must match item sums, got %s/%s", vol.PesoL, vol.PesoB)
	}
}

func TestMapDetachedTransport_TransportadoraCPFBranch(t *testing.T) {
	in := validTransportInput()
	street := "Rodovia BR-163"
	number := "Km 12"
	in.Logradouro = &street
	in.Numero = &number

	transport := mapDetachedTransport(in, 0, "MT", decimal.Zero, decimal.Zero)
	tra := transport.Transportadora
	if tra == nil {
		t.Fatal("expected a transportadora group")
	}
	if tra.Type != 2 || tra.CPF != "52998224725" || tra.CNPJ != "" {
		t.Errorf("expected CPF branch (Type 2), got %+v", tra)
	}
	if tra.Endereco != "Rodovia BR-163, Km 12" {
		t.Errorf("expected concatenated xEnder, got %q", tra.Endereco)
	}
	if tra.XNome != "Carreteiro Fulano" || tra.UF != "MT" {
		t.Errorf("transportadora fields mismatch: %+v", tra)
	}
}

func TestMapDetachedTransport_TransportadoraCNPJBranch(t *testing.T) {
	in := validTransportInput()
	doc := "11222333000181"
	in.Document = &doc

	transport := mapDetachedTransport(in, 4, "MT", decimal.Zero, decimal.Zero)
	tra := transport.Transportadora
	if tra == nil || tra.Type != 1 || tra.CNPJ != "11222333000181" || tra.CPF != "" {
		t.Errorf("expected CNPJ branch (Type 1), got %+v", tra)
	}
}

func TestMapDetachedTransport_VehiclePlateNormalizationAndUFFallback(t *testing.T) {
	in := validTransportInput()
	placa := "abc-1d23"
	in.Placa = &placa
	// PlacaUF intentionally blank: must fall back to the emitter UF.

	transport := mapDetachedTransport(in, 0, "MT", decimal.Zero, decimal.Zero)
	veic := transport.Veiculo
	if veic == nil {
		t.Fatal("expected a vehicle group")
	}
	if veic.Placa != "ABC1D23" {
		t.Errorf("expected normalized plate ABC1D23, got %q", veic.Placa)
	}
	if veic.UF != "MT" {
		t.Errorf("plate UF must fall back to emitter UF, got %q", veic.UF)
	}
}

func TestMapDetachedTransport_VolumesUserBlockWithFallbacks(t *testing.T) {
	marca := "BigFarm"
	qtd := 20
	pesoL := decimal.RequireFromString("15000")
	in := validTransportInput()
	in.VolQtd = &qtd
	in.VolMarca = &marca
	in.VolPesoL = &pesoL

	totalNet := decimal.RequireFromString("800")
	totalGross := decimal.RequireFromString("820")

	transport := mapDetachedTransport(in, 0, "MT", totalNet, totalGross)
	if len(transport.Volumes) != 1 {
		t.Fatalf("expected one volume block, got %v", transport.Volumes)
	}
	vol := transport.Volumes[0]
	if vol.QVol != 20 {
		t.Errorf("expected user qVol 20, got %d", vol.QVol)
	}
	// Blank esp falls back to Granel even inside the user block.
	if vol.Esp != "Granel" {
		t.Errorf("blank esp must fall back to Granel, got %q", vol.Esp)
	}
	if vol.Marca != "BigFarm" {
		t.Errorf("expected marca BigFarm, got %q", vol.Marca)
	}
	if !vol.PesoL.Equal(pesoL) {
		t.Errorf("expected user pesoL, got %s", vol.PesoL)
	}
	// Blank pesoB falls back to the item sums.
	if !vol.PesoB.Equal(totalGross) {
		t.Errorf("blank pesoB must fall back to the item sum, got %s", vol.PesoB)
	}
}

func TestDetachedTransportRecordFromEntity(t *testing.T) {
	input := entity.TransportData{
		ModFrete: 1,
		Transportadora: &entity.TransportadoraData{
			Type:      1,
			CNPJ:      "11222333000181",
			XNome:     "Transportadora XYZ",
			Endereco:  "Rua X, 10",
			Municipio: "Sorriso",
			UF:        "MT",
		},
		Veiculo: &entity.VeiculoData{Placa: "ABC1234", UF: "MT", RNTC: "12345"},
		Volumes: []entity.VolumeData{{
			QVol:  3,
			Esp:   "Big Bag",
			Marca: "MarcaZ",
			PesoL: decimal.RequireFromString("1000"),
			PesoB: decimal.RequireFromString("1010"),
		}},
	}

	record, ok := detachedTransportRecordFromEntity(input)
	if !ok {
		t.Fatal("expected the record to be created")
	}
	if record.Transportadora == nil || record.Transportadora.TpDoc != 1 ||
		record.Transportadora.DocNumber != "11222333000181" || record.Transportadora.XNome != "Transportadora XYZ" {
		t.Errorf("transportadora record mismatch: %+v", record.Transportadora)
	}
	if record.Veiculo == nil || record.Veiculo.Placa != "ABC1234" || record.Veiculo.RNTC != "12345" {
		t.Errorf("veiculo record mismatch: %+v", record.Veiculo)
	}
	if record.Vol == nil || record.Vol.Qtd != 3 || record.Vol.Esp != "Big Bag" ||
		!record.Vol.PesoL.Equal(decimal.RequireFromString("1000")) {
		t.Errorf("vol record mismatch: %+v", record.Vol)
	}

	if _, ok := detachedTransportRecordFromEntity(entity.TransportData{}); ok {
		t.Error("empty transport must not produce a persistence record")
	}
}

// TestBuildDetachedDANFEData_VehicleFields verifies the DANFE mapping carries
// the vehicle group (plate, UF and RNTC) from the transport data.
func TestBuildDetachedDANFEData_VehicleFields(t *testing.T) {
	svc := NewNFeService()

	input := entity.InvoiceInput{
		Serie:       1,
		Environment: 1,
		TpEmis:      defaults.EmissaoNormal,
		Transport: entity.TransportData{
			ModFrete: 1,
			Veiculo: &entity.VeiculoData{
				Placa: "ABC1D23",
				UF:    "MT",
				RNTC:  "12345678901234567890",
			},
			Volumes: []entity.VolumeData{{
				QVol:  2,
				Esp:   "Granel",
				PesoL: decimal.RequireFromString("1000.500"),
				PesoB: decimal.RequireFromString("1010.750"),
			}},
		},
	}

	data := svc.buildDetachedDANFEData(input)

	if data.VeicPlate != "ABC1D23" {
		t.Errorf("expected plate ABC1D23, got %q", data.VeicPlate)
	}
	if data.VeicUF != "MT" {
		t.Errorf("expected vehicle UF MT, got %q", data.VeicUF)
	}
	if data.VeicRNTC != "12345678901234567890" {
		t.Errorf("expected RNTC mapped to VeicRNTC, got %q", data.VeicRNTC)
	}
	if data.QVol != "2" || !data.PesoL.Equal(decimal.RequireFromString("1000.500")) {
		t.Errorf("volume mapping mismatch: qVol %q pesoL %s", data.QVol, data.PesoL)
	}
}
