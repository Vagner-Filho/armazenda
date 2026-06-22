package entity

import "github.com/shopspring/decimal"

// DANFEData holds the data needed to generate a DANFE.
type DANFEData struct {
	AccessKey      string
	EmitterName    string
	EmitterCNPJ    string
	EmitterAddress string
	EmitterCity    string
	EmitterUF      string
	DestName       string
	DestCNPJ       string
	DestAddress    string
	DestCity       string
	DestUF         string
	NaturezaOp     string
	Numero         int
	Serie          int
	EmissionDate   string
	Products       []DANFEProduct
	TotalValue     decimal.Decimal
	ICMSValue      decimal.Decimal
	Protocol       string
	ProtocolDate   string
}

// DANFEProduct holds a single product line for the DANFE.
type DANFEProduct struct {
	Code      string
	Desc      string
	NCM       string
	CFOP      string
	Unit      string
	Quantity  decimal.Decimal
	UnitPrice decimal.Decimal
	Total     decimal.Decimal
}
