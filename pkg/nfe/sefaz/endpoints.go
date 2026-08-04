package sefaz

import (
	"fmt"

	"armazenda/pkg/nfe/defaults"
)

// InfraType represents the SEFAZ infrastructure type for a state.
type InfraType int

const (
	InfraOwn  InfraType = iota // State runs its own endpoints (e.g., MT, SP, RS)
	InfraSVRS                  // Shared via SVRS (e.g., AC, AL, AP, RJ)
	InfraSVAN                  // Shared via SVAN legacy (e.g., MA, PA)
)

// Endpoint holds a single SEFAZ web service endpoint.
type Endpoint struct {
	Name         string
	Namespace    string // SOAP namespace for this service
	SOAPAction   string // SOAPAction header for Axis2 routing
	Production   string
	Homologation string
}

// EndpointSet groups all NF-e web service endpoints for a state or virtual environment.
type EndpointSet struct {
	Name      string
	InfraType InfraType
	Endpoints []Endpoint
}

// MSEndpointSet holds all NF-e web service endpoints for Mato Grosso do Sul.
var MSEndpointSet = &EndpointSet{
	Name:      "MS",
	InfraType: InfraOwn,
	Endpoints: []Endpoint{
		{
			Name:         "NFeAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeAutorizacao4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeAutorizacao4",
		},
		{
			Name:         "NFeRetAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeRetAutorizacao4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeRetAutorizacao4",
		},
		{
			Name:         "NFeConsultaProtocolo4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4",
		},
		{
			Name:         "NfeStatusServico4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeStatusServico4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeStatusServico4",
		},
		{
			Name:         "RecepcaoEvento4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4/nfeRecepcaoEvento",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4",
		},
		{
			Name:         "NfeInutilizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/NFeInutilizacao4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/NFeInutilizacao4",
		},
		{
			Name:         "CadConsultaCadastro4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4/consultaCadastro",
			Production:   "https://nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4",
			Homologation: "https://hom.nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4",
		},
	},
}

// MTEndpointSet holds all NF-e web service endpoints for Mato Grosso.
var MTEndpointSet = &EndpointSet{
	Name:      "MT",
	InfraType: InfraOwn,
	Endpoints: []Endpoint{
		{
			Name:         "NFeAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4",
		},
		{
			Name:         "NFeRetAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4",
		},
		{
			Name:         "NFeConsultaProtocolo4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4",
		},
		{
			Name:         "NfeStatusServico4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4",
		},
		{
			Name:         "RecepcaoEvento4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4/nfeRecepcaoEvento",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4",
		},
		{
			Name:         "NfeInutilizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4",
		},
		{
			Name:         "CadConsultaCadastro4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4/consultaCadastro",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4",
		},
	},
}

// SVCANEndpointSet holds all NF-e web service endpoints for SVC-AN (SEFAZ Virtual do Ambiente Nacional).
var SVCANEndpointSet = &EndpointSet{
	Name:      "SVC-AN",
	InfraType: InfraSVAN,
	Endpoints: []Endpoint{
		{
			Name:         "NFeAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
			Production:   "https://www.svc.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx",
			Homologation: "https://hom.svc.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx",
		},
		{
			Name:         "NFeRetAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
			Production:   "https://www.svc.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx",
			Homologation: "https://hom.svc.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx",
		},
		{
			Name:         "NFeConsultaProtocolo4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
			Production:   "https://www.svc.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx",
			Homologation: "https://hom.svc.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx",
		},
		{
			Name:         "NfeStatusServico4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
			Production:   "https://www.svc.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx",
			Homologation: "https://hom.svc.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx",
		},
		{
			Name:         "RecepcaoEvento4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4/nfeRecepcaoEvento",
			Production:   "https://www.svc.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
			Homologation: "https://hom.svc.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
		},
	},
}

// SVCRSEndpointSet holds all NF-e web service endpoints for SVC-RS (SEFAZ Virtual do Rio Grande do Sul).
var SVCRSEndpointSet = &EndpointSet{
	Name:      "SVC-RS",
	InfraType: InfraSVRS,
	Endpoints: []Endpoint{
		{
			Name:         "NFeAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
			Production:   "https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
			Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
		},
		{
			Name:         "NFeRetAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
			Production:   "https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
			Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
		},
		{
			Name:         "NFeConsultaProtocolo4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
			Production:   "https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NFeConsultaProtocolo4.asmx",
			Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NFeConsultaProtocolo4.asmx",
		},
		{
			Name:         "NfeStatusServico4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
			Production:   "https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
			Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
		},
		{
			Name:         "RecepcaoEvento4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4/nfeRecepcaoEvento",
			Production:   "https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
			Homologation: "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
		},
	},
}

// SVCRegistry maps a SVC emission type to its EndpointSet.
var SVCRegistry = map[defaults.TpEmis]*EndpointSet{
	defaults.SVCAN: SVCANEndpointSet,
	defaults.SVCRS: SVCRSEndpointSet,
}

// StateRegistry maps a state UF to its EndpointSet.
// To add a new state, register its EndpointSet here.
// Multiple states can point to the same EndpointSet (e.g., SVRS states).
var StateRegistry = map[string]*EndpointSet{
	"MT": MTEndpointSet,
	"MS": MSEndpointSet,
}

// GetEndpoint returns the URL for a specific service in a given state.
func GetEndpoint(uf, serviceName string, production bool) (string, error) {
	set, ok := StateRegistry[uf]
	if !ok {
		return "", fmt.Errorf("state %s not registered in endpoint registry", uf)
	}
	for _, ep := range set.Endpoints {
		if ep.Name == serviceName {
			if production {
				return ep.Production, nil
			}
			return ep.Homologation, nil
		}
	}
	return "", fmt.Errorf("service %s not found for state %s", serviceName, uf)
}

// GetEndpointWithNamespace returns the URL and SOAP namespace for a service.
func GetEndpointWithNamespace(uf, serviceName string, production bool) (string, string, error) {
	url, ns, _, err := GetEndpointWithSOAPAction(uf, serviceName, production)
	return url, ns, err
}

// GetEndpointWithSOAPAction returns the URL, SOAP namespace, and SOAPAction for a service.
func GetEndpointWithSOAPAction(uf, serviceName string, production bool) (string, string, string, error) {
	return GetEndpointWithSOAPActionAndEmission(uf, serviceName, production, defaults.EmissaoNormal)
}

// GetEndpointWithSOAPActionAndEmission returns the URL, SOAP namespace, and SOAPAction for a service,
// routing to SVC endpoints when tpEmis indicates a contingency mode.
func GetEndpointWithSOAPActionAndEmission(uf, serviceName string, production bool, tpEmis defaults.TpEmis) (string, string, string, error) {
	var set *EndpointSet

	switch tpEmis {
	case defaults.SVCAN, defaults.SVCRS:
		var ok bool
		set, ok = SVCRegistry[tpEmis]
		if !ok {
			return "", "", "", fmt.Errorf("SVC endpoint set not found for tpEmis %s", tpEmis)
		}
	default:
		var ok bool
		set, ok = StateRegistry[uf]
		if !ok {
			return "", "", "", fmt.Errorf("state %s not registered in endpoint registry", uf)
		}
	}

	for _, ep := range set.Endpoints {
		if ep.Name == serviceName {
			if production {
				return ep.Production, ep.Namespace, ep.SOAPAction, nil
			}
			return ep.Homologation, ep.Namespace, ep.SOAPAction, nil
		}
	}
	return "", "", "", fmt.Errorf("service %s not found for %s", serviceName, set.Name)
}

// GetInfraType returns the infrastructure type for a state.
func GetInfraType(uf string) (InfraType, error) {
	set, ok := StateRegistry[uf]
	if !ok {
		return InfraOwn, fmt.Errorf("state %s not registered in endpoint registry", uf)
	}
	return set.InfraType, nil
}

// IsStateRegistered checks if a state has endpoints registered.
func IsStateRegistered(uf string) bool {
	_, ok := StateRegistry[uf]
	return ok
}
