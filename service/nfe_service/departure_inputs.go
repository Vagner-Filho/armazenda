package nfe_service

import (
	"fmt"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// TaxCodeOption is a server-rendered finite tax-code choice for the departure
// emission modal.
type TaxCodeOption struct {
	Value string
	Label string
}

var departureCSTOptions = []TaxCodeOption{
	{Value: "00", Label: "00 - Tributada integralmente"},
	{Value: "20", Label: "20 - Com redução da base"},
	{Value: "40", Label: "40 - Isenta"},
	{Value: "41", Label: "41 - Não tributada"},
	{Value: "50", Label: "50 - Suspensão"},
	{Value: "51", Label: "51 - Diferimento"},
	{Value: "90", Label: "90 - Outras"},
}

var departureCSOSNOptions = []TaxCodeOption{
	{Value: "101", Label: "101 - Tributada com permissão de crédito"},
	{Value: "102", Label: "102 - Tributada sem permissão de crédito"},
	{Value: "103", Label: "103 - Isenção para faixa de receita bruta"},
	{Value: "201", Label: "201 - Tributada com permissão de crédito e ICMS-ST"},
	{Value: "202", Label: "202 - Tributada sem permissão de crédito e com ICMS-ST"},
	{Value: "203", Label: "203 - Isenção e ICMS-ST"},
	{Value: "300", Label: "300 - Imune"},
	{Value: "400", Label: "400 - Não tributada"},
	{Value: "500", Label: "500 - ICMS cobrado anteriormente por substituição"},
	{Value: "900", Label: "900 - Outros"},
}

var departurePISCOFINSOptions = []TaxCodeOption{
	{Value: "01", Label: "01 - Tributável cumulativo"},
	{Value: "02", Label: "02 - Tributável não cumulativo"},
	{Value: "04", Label: "04 - Monofásico (alíquota zero)"},
	{Value: "05", Label: "05 - Suspenso por litígio"},
	{Value: "06", Label: "06 - Alíquota zero"},
	{Value: "07", Label: "07 - Isenta"},
	{Value: "08", Label: "08 - Sem incidência"},
	{Value: "09", Label: "09 - Suspensão"},
}

// DepartureICMSOptions returns CST choices for regime normal and CSOSN choices
// for Simples Nacional. The values track the ICMS groups supported by the XML
// builder and prevent the CST/CSOSN field from accepting arbitrary text.
func DepartureICMSOptions(regime defaults.TaxRegime) []TaxCodeOption {
	options := departureCSTOptions
	if regime == defaults.TaxRegimeSimplesNacional {
		options = departureCSOSNOptions
	}
	return append([]TaxCodeOption(nil), options...)
}

// DeparturePISCOFINSOptions returns the CST choices shared by PIS and COFINS.
func DeparturePISCOFINSOptions() []TaxCodeOption {
	return append([]TaxCodeOption(nil), departurePISCOFINSOptions...)
}

func validateDepartureTaxCodes(codes entity.InvoiceOverrides, regime defaults.TaxRegime) []string {
	var errs []string
	if codes.ICMSCST != nil && *codes.ICMSCST != "" && !containsTaxCode(DepartureICMSOptions(regime), *codes.ICMSCST) {
		label := "CST ICMS"
		if regime == defaults.TaxRegimeSimplesNacional {
			label = "CSOSN ICMS"
		}
		errs = append(errs, fmt.Sprintf("%s inválido: %s", label, *codes.ICMSCST))
	}
	if codes.PISCST != nil && *codes.PISCST != "" && !containsTaxCode(departurePISCOFINSOptions, *codes.PISCST) {
		errs = append(errs, fmt.Sprintf("CST PIS inválido: %s", *codes.PISCST))
	}
	if codes.COFINSCST != nil && *codes.COFINSCST != "" && !containsTaxCode(departurePISCOFINSOptions, *codes.COFINSCST) {
		errs = append(errs, fmt.Sprintf("CST COFINS inválido: %s", *codes.COFINSCST))
	}
	return errs
}

func containsTaxCode(options []TaxCodeOption, code string) bool {
	for _, option := range options {
		if option.Value == code {
			return true
		}
	}
	return false
}

func validateDeparturePayment(options *entity.DepartureInvoiceOverrides) []string {
	if options == nil {
		return nil
	}
	var errs []string
	if options.TPag != "" && options.TPag != "01" && options.TPag != "90" && options.TPag != "99" {
		errs = append(errs, "Forma de pagamento inválida; selecione 01, 90 ou 99")
	}
	return errs
}

func validDepartureModFrete(value int) bool {
	switch value {
	case 0, 1, 2, 3, 4, 9:
		return true
	default:
		return false
	}
}

func departurePayment(options *entity.DepartureInvoiceOverrides, total decimal.Decimal) entity.PaymentData {
	tpag := "90"
	if options != nil && options.TPag != "" {
		tpag = options.TPag
	}
	payment := defaultPayment(tpag, total)
	if len(payment.Detalhes) > 0 {
		payment.Detalhes[0].IndPag = 0
	}
	return payment
}

func departureCobrancaEnabled(options *entity.DepartureInvoiceOverrides) bool {
	return true
}

func emissionInputsFromInvoiceInput(input entity.InvoiceInput) *entity.InvoiceEmissionInputs {
	includeCobranca := true
	if input.Cobranca != nil {
		includeCobranca = input.Cobranca.Enabled
	}
	return &entity.InvoiceEmissionInputs{
		Payment:         input.Payment,
		IncludeCobranca: includeCobranca,
		Transport:       input.Transport,
	}
}

func applyInvoiceEmissionInputs(input *entity.InvoiceInput, snapshot *entity.InvoiceEmissionInputs) {
	if input == nil || snapshot == nil {
		return
	}
	input.Payment = snapshot.Payment
	input.Cobranca = &entity.CobrancaData{Enabled: snapshot.IncludeCobranca}
	input.Transport = snapshot.Transport
}

// departureTransportInput adapts the departure form's typed fields to the
// shared Grupo X validator/mapper used by Detached NF-e.
func departureTransportInput(in entity.DepartureTransportInput) DetachedTransportInput {
	stringRef := func(value string) *string {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil
		}
		return &value
	}
	return DetachedTransportInput{
		Nome:       stringRef(in.TransporterName),
		Document:   stringRef(in.TransporterDocument),
		IE:         stringRef(in.TransporterIE),
		Logradouro: stringRef(in.TransporterStreet),
		Numero:     stringRef(in.TransporterNumber),
		Municipio:  stringRef(in.TransporterCity),
		UF:         stringRef(in.TransporterUF),
		Placa:      stringRef(in.VehiclePlate),
		PlacaUF:    stringRef(in.VehicleUF),
		RNTC:       stringRef(in.VehicleRNTC),
		VolQtd:     in.VolumeQuantity,
		VolEsp:     stringRef(in.VolumeSpecies),
		VolMarca:   stringRef(in.VolumeBrand),
		VolPesoL:   in.VolumeNet,
		VolPesoB:   in.VolumeGross,
	}
}

// mergeDepartureTransport applies user values to the legacy romaneio-derived
// transport block. Explicit false inclusion flags suppress a derived group;
// otherwise manual data replaces the corresponding defaults.
func mergeDepartureTransport(
	base entity.TransportData,
	form *entity.DepartureTransportInput,
	departure entity_public.Departure,
	emitterUF, destinationUF string,
) (entity.TransportData, []string) {
	if form == nil {
		return base, nil
	}
	in := departureTransportInput(*form)
	if form.IncludeVehicle != nil && !*form.IncludeVehicle {
		in.Placa = nil
		in.PlacaUF = nil
		in.RNTC = nil
	}
	if form.IncludeVolumes != nil && !*form.IncludeVolumes {
		in.VolQtd = nil
		in.VolEsp = nil
		in.VolMarca = nil
		in.VolPesoL = nil
		in.VolPesoB = nil
	}

	if errs := validateDetachedTransport(in, destinationUF, emitterUF); len(errs) > 0 {
		return base, errs
	}
	mapped := mapDetachedTransport(in, base.ModFrete, emitterUF, departure.NetWeight, departure.GrossWeight)
	if mapped.Transportadora != nil {
		base.Transportadora = mapped.Transportadora
	}
	if form.IncludeVehicle != nil && !*form.IncludeVehicle {
		base.Veiculo = nil
	} else if mapped.Veiculo != nil {
		base.Veiculo = mapped.Veiculo
	}
	if form.IncludeVolumes != nil && !*form.IncludeVolumes {
		base.Volumes = nil
	} else if form.IncludeVolumes != nil && *form.IncludeVolumes {
		base.Volumes = mapped.Volumes
	} else if hasVolValue(in) {
		base.Volumes = mapped.Volumes
	}
	return base, nil
}
