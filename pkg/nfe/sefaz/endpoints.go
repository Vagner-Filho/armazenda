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
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4?wsdl",
		},
		{
			Name:         "NFeRetAutorizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4?wsdl",
		},
		{
			Name:         "NfeConsulta4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsulta4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4?wsdl",
		},
		{
			Name:         "NfeStatusServico4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4?wsdl",
		},
		{
			Name:         "RecepcaoEvento4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/RecepcaoEvento4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4?wsdl",
		},
		{
			Name:         "NfeInutilizacao4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4?wsdl",
		},
		{
			Name:         "CadConsultaCadastro4",
			Namespace:    "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4",
			Production:   "https://nfe.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4?wsdl",
			Homologation: "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4?wsdl",
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
	set, ok := StateRegistry[uf]
	if !ok {
		return "", "", fmt.Errorf("state %s not registered in endpoint registry", uf)
	}
	for _, ep := range set.Endpoints {
		if ep.Name == serviceName {
			if production {
				return ep.Production, ep.Namespace, nil
			}
			return ep.Homologation, ep.Namespace, nil
		}
	}
	return "", "", fmt.Errorf("service %s not found for state %s", serviceName, uf)
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
