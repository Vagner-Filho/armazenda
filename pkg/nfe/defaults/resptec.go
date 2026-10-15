package defaults

// resptec.go holds the system-scoped identificação do responsável técnico
// (infRespTec, NT 2022.002 / E28a). The block identifies the software vendor
// that generated the NF-e — Armazenda itself — so it is SYSTEM-scoped, not
// farm-scoped: the same values apply to every tenant and to both
// environments (production and homologação), and no farm-level
// configuration, migration or UI is involved.
//
// Provider data supplied by the human. All four blocks are mandatory in the
// 4.00 XSD (CNPJ, xContato, email, fone — each 1-1); RespTecConfigured()
// gates on the complete set so an incomplete vendor identity omits the whole
// block instead of emitting an invalid one (SEFAZ 215).
//
// idCSRT / CSRT (the SEFAZ admission pair issued to a registered vendor)
// is a later follow-up; it is not carried here.

// SystemRespTecData mirrors the E28a fields in layout order. Local struct —
// pkg/nfe/entity imports this package, so the service maps it into
// entity.RespTecData for the builder input.
type SystemRespTecData struct {
	CNPJ     string
	XContato string
	Email    string
	Fone     string
}

// RespTec is the vendor (system) responsible data for every NF-e emitted
// through Armazenda.
//
// xContato carries the full vendor razão social (72 chars); the TString
// element caps at 60 chars (SEFAZ cvc-type.3.1.3 on the first homolog
// attempt) — the builder truncates the emitted <xContato> at 60 runes
// without warning, so this constant may hold the long form safely.
var RespTec = SystemRespTecData{
	CNPJ:     "47962538000140",
	XContato: "VAGNER SIMONE MARTINS FILHO CONSULTORIA EM TECNOLOGIA DA INFORMACAO LTDA",
	Email:    "vagner.filhopro@gmail.com",
	Fone:     "66996615930",
}

// RespTecConfigured reports whether the system-level responsável técnico
// block is emitted: the gate is on while all four 1-1 fields (CNPJ,
// xContato, email, fone) are present, so an incomplete vendor identity
// keeps the whole block omitted instead of emitting one SEFAZ rejects.
// When false, both emission paths leave RespTec nil and the builder omits
// <infRespTec> entirely.
func RespTecConfigured() bool {
	return RespTec.CNPJ != "" && RespTec.XContato != "" && RespTec.Email != "" && RespTec.Fone != ""
}
