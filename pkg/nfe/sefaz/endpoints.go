package sefaz

import (
	"fmt"
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
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4",
			SOAPAction:   "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4/nfeRecepcaoEvento",
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

// StateRegistry maps a state UF to its EndpointSet.
// To add a new state, register its EndpointSet here.
// Multiple states can point to the same EndpointSet (e.g., SVRS states).
var StateRegistry = map[string]*EndpointSet{
	"MT": MTEndpointSet,
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
	set, ok := StateRegistry[uf]
	if !ok {
		return "", "", "", fmt.Errorf("state %s not registered in endpoint registry", uf)
	}
	for _, ep := range set.Endpoints {
		if ep.Name == serviceName {
			if production {
				return ep.Production, ep.Namespace, ep.SOAPAction, nil
			}
			return ep.Homologation, ep.Namespace, ep.SOAPAction, nil
		}
	}
	return "", "", "", fmt.Errorf("service %s not found for state %s", serviceName, uf)
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
