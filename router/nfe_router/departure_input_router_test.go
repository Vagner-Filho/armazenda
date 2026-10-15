package nfe_router

import (
	"net/url"
	"testing"
)

func TestParseDepartureFormOverrides_PaymentBillingAndTransport(t *testing.T) {
	values := url.Values{
		"modFrete":                 {"1"},
		"tpag":                     {"01"},
		"indPag":                   {"1"},
		"includeCobrancaSpecified": {"1"},
		"includeCobranca":          {"0"},
		"transportFieldsSpecified": {"1"},
		"transportNome":            {"Transportador Teste"},
		"transportDocumento":       {"529.982.247-25"},
		"transportLogradouro":      {"Rodovia MT-100"},
		"transportNumero":          {"Km 10"},
		"transportMunicipio":       {"Cuiabá"},
		"transportUF":              {"mt"},
		"includeVehicleSpecified":  {"1"},
		"includeVehicle":           {"0"},
		"transportPlaca":           {"ABC1D23"},
		"includeVolumesSpecified":  {"1"},
		"includeVolumes":           {"1"},
		"volQtd":                   {"2"},
		"volEsp":                   {"Sacas"},
		"volPesoL":                 {"95,500"},
		"volPesoB":                 {"100.250"},
	}
	overrides, toast := parseDepartureFormOverrides(newFormContext(t, values))
	if toast.Type != 0 {
		t.Fatalf("unexpected parser toast: %+v", toast)
	}
	if overrides == nil || overrides.Departure == nil {
		t.Fatal("expected departure-specific overrides")
	}
	departure := overrides.Departure
	if departure.TPag != "01" || departure.IndPag != 1 {
		t.Errorf("payment options = tPag %q / indPag %d", departure.TPag, departure.IndPag)
	}
	if departure.IncludeCobranca == nil || *departure.IncludeCobranca {
		t.Errorf("IncludeCobranca = %v, want explicit false", departure.IncludeCobranca)
	}
	transport := departure.Transport
	if transport.TransporterDocument != "529.982.247-25" || transport.TransporterUF != "MT" {
		t.Errorf("transportadora values not normalized/preserved: %+v", transport)
	}
	if transport.IncludeVehicle == nil || *transport.IncludeVehicle {
		t.Errorf("IncludeVehicle = %v, want explicit false", transport.IncludeVehicle)
	}
	if transport.IncludeVolumes == nil || !*transport.IncludeVolumes {
		t.Errorf("IncludeVolumes = %v, want true", transport.IncludeVolumes)
	}
	if transport.VolumeQuantity == nil || *transport.VolumeQuantity != 2 {
		t.Errorf("VolumeQuantity = %v, want 2", transport.VolumeQuantity)
	}
	if transport.VolumeNet == nil || transport.VolumeNet.String() != "95.5" {
		t.Errorf("VolumeNet = %v, want 95.5", transport.VolumeNet)
	}
	if transport.VolumeGross == nil || transport.VolumeGross.String() != "100.25" {
		t.Errorf("VolumeGross = %v, want 100.25", transport.VolumeGross)
	}
	if overrides.ModFrete == nil || *overrides.ModFrete != 1 {
		t.Errorf("ModFrete = %v, want 1", overrides.ModFrete)
	}
}

func TestParseDepartureFormOverrides_InvalidValuesAndLegacyRequests(t *testing.T) {
	legacy, toast := parseDepartureFormOverrides(newFormContext(t, url.Values{"modFrete": {"1"}}))
	if toast.Type != 0 || legacy == nil || legacy.Departure != nil || legacy.ModFrete == nil || *legacy.ModFrete != 1 {
		t.Errorf("legacy request should retain only the existing modFrete override; got %+v / %+v", legacy, toast)
	}

	for name, value := range map[string]string{
		"bad volume integer": "1.5",
	} {
		values := url.Values{
			"transportFieldsSpecified": {"1"},
			"includeVolumesSpecified":  {"1"},
			"includeVolumes":           {"1"},
			"volQtd":                   {value},
		}
		if _, toast := parseDepartureFormOverrides(newFormContext(t, values)); toast.Type == 0 {
			t.Errorf("%s should be rejected", name)
		}
	}

	if _, toast := parseDepartureFormOverrides(newFormContext(t, url.Values{"modFrete": {"8"}})); toast.Type == 0 {
		t.Error("unsupported modFrete should be rejected")
	}

}
