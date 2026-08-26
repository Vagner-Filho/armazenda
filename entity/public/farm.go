package entity_public

import (
	"time"

	"github.com/shopspring/decimal"
)

type Farm struct {
	Id                    uint32  `form:"id"`
	InscricaoEstadual     string  `form:"inscricaoEstadual" binding:"required"`
	OwnerDocument         *string `form:"ownerDocument"`
	OwnerDocumentType     *int    `form:"ownerDocumentType"`
	Name                  *string `form:"name"`
	HumidityProgressionId *uint32 `form:"humidityProgressionId" json:"humidityProgressionId"`
	Address
	StorageName                   *string `form:"storageName"`
	FarmUsedHumidityProgressionId *uint32 `form:"farmUsedHumidityProgressionId" json:"farmUsedHumidityProgressionId"`
	UF                            string  `form:"uf"`
}

type FarmCND struct {
	CertificateNumber *string                 `form:"certificateNumber"`
	ExpDate           *time.Time              `form:"expDate" time_format:"2006-01-02"`
	Meta              *map[string]interface{} `form:"meta"`
}

// FarmConfig represents the NFe configuration for a farm.
type FarmConfig struct {
	ID                           uint32
	FarmID                       *uint32
	CertificatePath              *string
	CertificateData              []byte
	CertificatePasswordEncrypted string `form:"certificatePassword"`
	Environment                  int    `f̀orm:"environment"`
	Serie                        int    `form:"serie"`
	NextNumber                   int
	TaxRegime                    int `form:"taxRegime"`
	EmitterType                  int
	DocEmitter                   *string
	IEEmitter                    string
	EmitterUF                    string
	DefaultModFrete              int             `form:"defaultModFrete"`
	DefaultCFOP                  string          `form:"defaultCFOP"`
	DefaultCEST                  *string         `form:"defaultCEST"`
	DefaultUnit                  string          `form:"defaultUnit"`
	DefaultICMSCST               *string         `form:"defaultICMSCST"`
	DefaultPISCST                *string         `form:"defaultPISCST"`
	DefaultCOFINSCST             *string         `form:"defaultCOFINSCST"`
	DefaultNaturezaOp            *string         `form:"defaultNaturezaOp"`
	ICMSRate                     decimal.Decimal `form:"icmsRate"`
	PISRate                      decimal.Decimal `form:"pisRate"`
	COFINSRate                   decimal.Decimal `form:"cofinsRate"`
	// Tax reform (IBS / CBS) — Reforma Tributária, NT 2025.002-RTC.
	// IBSRate / CBSRate are decimal *rates* (e.g. 0.001 = 0.1 % IBS,
	// 0.009 = 0.9 % CBS). CST fields hold the 3-digit IBSCBS classification.
	IBSRate                     decimal.Decimal `form:"ibsRate"`
	CBSRate                     decimal.Decimal `form:"cbsRate"`
	DefaultIBSCST               *string         `form:"defaultIBSCST"`
	DefaultCBSCST               *string         `form:"defaultCBSCST"`
	DefaultCClassTrib           *string         `form:"defaultCClassTrib"`
	FarmCND
}
