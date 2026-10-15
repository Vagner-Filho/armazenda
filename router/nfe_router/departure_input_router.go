package nfe_router

import (
	"fmt"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/pkg/nfe/entity"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// parseDepartureFormOverrides combines the existing scalar overrides with the
// payment, billing, and transport controls added to the departure modal. Old
// clients that do not send the new markers retain legacy service defaults.
func parseDepartureFormOverrides(c *gin.Context) (*entity.InvoiceOverrides, entity_public.Toast) {
	overrides := parseInvoiceOverrides(c)
	departure, toast := parseDepartureInvoiceOverrides(c)
	if toast.Type == entity_public.WarningToast || toast.Type == entity_public.ErrorToast {
		return nil, toast
	}
	if departure != nil {
		if overrides == nil {
			overrides = &entity.InvoiceOverrides{}
		}
		overrides.Departure = departure
	}
	return overrides, entity_public.Toast{}
}

func parseDepartureInvoiceOverrides(c *gin.Context) (*entity.DepartureInvoiceOverrides, entity_public.Toast) {
	if raw, exists := c.GetPostForm("modFrete"); exists && strings.TrimSpace(raw) != "" {
		modFrete, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || !validModFreteFormValue(modFrete) {
			return nil, entity_public.GetWarningToast("Modalidade de frete inválida", "Selecione uma modalidade de frete válida")
		}
	}
	hasPayment := hasPostField(c, "tpag")
	hasTransport := hasPostField(c, "transportFieldsSpecified") || hasAnyTransportValue(c)
	if !hasPayment && !hasTransport {
		return nil, entity_public.Toast{}
	}

	options := &entity.DepartureInvoiceOverrides{}
	if hasPayment {
		tpag := strings.TrimSpace(c.PostForm("tpag"))
		if tpag == "" {
			tpag = "90"
		}
		options.TPag = tpag
	}

	if hasTransport {
		transport := entity.DepartureTransportInput{
			TransporterName:     safePostValue(c, "transportNome"),
			TransporterDocument: safePostValue(c, "transportDocumento"),
			TransporterIE:       safePostValue(c, "transportIE"),
			TransporterStreet:   safePostValue(c, "transportLogradouro"),
			TransporterNumber:   safePostValue(c, "transportNumero"),
			TransporterCity:     safePostValue(c, "transportMunicipio"),
			TransporterUF:       strings.ToUpper(safePostValue(c, "transportUF")),
			VehiclePlate:        safePostValue(c, "transportPlaca"),
			VehicleUF:           strings.ToUpper(safePostValue(c, "transportPlacaUF")),
			VehicleRNTC:         safePostValue(c, "transportRNTC"),
			VolumeSpecies:       safePostValue(c, "volEsp"),
			VolumeBrand:         safePostValue(c, "volMarca"),
		}

		if hasPostField(c, "includeVehicleSpecified") {
			include, err := parseMarkedBool(c, "includeVehicleSpecified", "includeVehicle")
			if err != nil {
				return nil, entity_public.GetWarningToast("Dados do veículo inválidos", err.Error())
			}
			transport.IncludeVehicle = &include
		}
		if hasPostField(c, "includeVolumesSpecified") {
			include, err := parseMarkedBool(c, "includeVolumesSpecified", "includeVolumes")
			if err != nil {
				return nil, entity_public.GetWarningToast("Dados dos volumes inválidos", err.Error())
			}
			transport.IncludeVolumes = &include
		}

		volumesEnabled := transport.IncludeVolumes == nil || *transport.IncludeVolumes
		if volumesEnabled {
			if value := strings.TrimSpace(c.PostForm("volQtd")); value != "" {
				parsed, err := strconv.Atoi(value)
				if err != nil {
					return nil, entity_public.GetWarningToast("Dados dos volumes inválidos", "Informe uma quantidade inteira de volumes")
				}
				transport.VolumeQuantity = &parsed
			}
			var parseToast *entity_public.Toast
			transport.VolumeNet, parseToast = parseDepartureDecimal(c, "volPesoL", "peso líquido dos volumes")
			if parseToast != nil {
				return nil, *parseToast
			}
			transport.VolumeGross, parseToast = parseDepartureDecimal(c, "volPesoB", "peso bruto dos volumes")
			if parseToast != nil {
				return nil, *parseToast
			}
		}

		options.Transport = transport
	}

	return options, entity_public.Toast{}
}

func parseDepartureDecimal(c *gin.Context, key, label string) (*decimal.Decimal, *entity_public.Toast) {
	value := strings.TrimSpace(c.PostForm(key))
	if value == "" {
		return nil, nil
	}
	// Accept the Brazilian decimal comma used by the form as well as a dot.
	if strings.Contains(value, ",") && strings.Contains(value, ".") {
		toast := entity_public.GetWarningToast("Valor inválido", fmt.Sprintf("Informe um valor válido para %s", label))
		return nil, &toast
	}
	value = strings.ReplaceAll(value, ",", ".")
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		toast := entity_public.GetWarningToast("Valor inválido", fmt.Sprintf("Informe um valor válido para %s", label))
		return nil, &toast
	}
	return &parsed, nil
}

func parseMarkedBool(c *gin.Context, marker, field string) (bool, error) {
	if _, exists := c.GetPostForm(marker); !exists {
		return false, fmt.Errorf("marcador ausente para %s", field)
	}
	value := strings.ToLower(strings.TrimSpace(c.PostForm(field)))
	switch value {
	case "1", "true", "on":
		return true, nil
	case "", "0", "false", "off":
		return false, nil
	default:
		return false, fmt.Errorf("valor inválido para %s", field)
	}
}

func hasPostField(c *gin.Context, key string) bool {
	_, exists := c.GetPostForm(key)
	return exists
}

func safePostValue(c *gin.Context, key string) string {
	return strings.TrimSpace(c.PostForm(key))
}

func hasAnyTransportValue(c *gin.Context) bool {
	for _, key := range []string{
		"transportNome", "transportDocumento", "transportIE", "transportLogradouro",
		"transportNumero", "transportMunicipio", "transportUF", "transportPlaca",
		"transportPlacaUF", "transportRNTC", "volQtd", "volEsp", "volMarca",
		"volPesoL", "volPesoB",
	} {
		if hasPostField(c, key) {
			return true
		}
	}
	return false
}

func validModFreteFormValue(value int) bool {
	switch value {
	case 0, 1, 2, 3, 4, 9:
		return true
	default:
		return false
	}
}
