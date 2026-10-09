package nfe_service

import (
	"fmt"
	"strings"

	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/entity"

	"github.com/shopspring/decimal"
)

// DetachedTransportInput is the transport section (MOC 7.0, Grupo X) of the
// detached NF-e form. Every field is optional; undefined pointers mean "not
// informed" and fall back to the existing auto-derivation. Nil pointers are
// converted from empty form strings by the router.
type DetachedTransportInput struct {
	// Transportadora (transporta, X03-X10)
	Nome       *string // xNome (X06)
	Document   *string // CPF (11 digits) or CNPJ (14 digits), auto-detected
	IE         *string // ISENTO literal allowed; UF becomes required
	Logradouro *string // folded into xEnder (no dedicated schema tag)
	Numero     *string // folded into xEnder ("Logradouro, Número")
	Municipio  *string // xMun (X09)
	UF         *string // transportadora UF (X10)

	// Veículo (veicTransp, X18-X21)
	Placa   *string
	PlacaUF *string
	RNTC    *string // ANTT (Registro Nacional de Transportador de Carga)

	// Volumes (vol, X26-X32). Blank fields fall back per-field: qVol -> 1,
	// esp -> "Granel", pesoL -> sum of item quantities, pesoB -> sum of item
	// gross weights (the pre-transport auto-derivation).
	VolQtd   *int
	VolEsp   *string
	VolMarca *string
	VolPesoL *decimal.Decimal
	VolPesoB *decimal.Decimal
}

// strPtr returns the string value of p, or "" when nil.
func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// decPtr returns p, guarding against a nil pointer.
func decPtr(p *decimal.Decimal) decimal.Decimal {
	if p == nil {
		return decimal.Zero
	}
	return *p
}

// hasTransportadora reports whether any transporta group field was informed.
func hasTransportadora(in DetachedTransportInput) bool {
	return strPtr(in.Nome) != "" || strPtr(in.Document) != "" || strPtr(in.IE) != "" ||
		strPtr(in.Logradouro) != "" || strPtr(in.Numero) != "" ||
		strPtr(in.Municipio) != "" || strPtr(in.UF) != ""
}

// hasVolValue reports whether any volume field was informed by the user.
func hasVolValue(in DetachedTransportInput) bool {
	return in.VolQtd != nil || strPtr(in.VolEsp) != "" || strPtr(in.VolMarca) != "" ||
		in.VolPesoL != nil || in.VolPesoB != nil
}

// transportadorXender concatenates logradouro and número into the single
// xEnder field (X08, 1-60 chars). The NF-e schema has no dedicated number tag.
func transportadorXender(in DetachedTransportInput) string {
	street := strings.TrimSpace(strPtr(in.Logradouro))
	number := strings.TrimSpace(strPtr(in.Numero))
	switch {
	case street != "" && number != "":
		return fmt.Sprintf("%s, %s", street, number)
	case street != "":
		return street
	default:
		return number
	}
}

// cleanDigits strips every non-digit rune (punctuation, separators, spaces).
func cleanDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// allIdenticalDigits reports whether the digit string repeats one digit
// (SEFAZ rejects CPF documents like "111...", "222...", zeros).
func allIdenticalDigits(digits string) bool {
	if digits == "" {
		return false
	}
	first := digits[0]
	for i := 1; i < len(digits); i++ {
		if digits[i] != first {
			return false
		}
	}
	return true
}

// isValidCPF validates the two check digits of an 11-character digit string.
func isValidCPF(digits string) bool {
	if len(digits) != 11 || allIdenticalDigits(digits) {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	dv1 := 11 - (sum % 11)
	if dv1 >= 10 {
		dv1 = 0
	}
	if dv1 != int(digits[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	dv2 := 11 - (sum % 11)
	if dv2 >= 10 {
		dv2 = 0
	}
	return dv2 == int(digits[10]-'0')
}

// isValidCNPJ validates the two check digits of a 14-character digit string.
func isValidCNPJ(digits string) bool {
	if len(digits) != 14 || allIdenticalDigits(digits) {
		return false
	}
	// Canonical MOC weight sequences, left-to-right.
	dv1Weights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	dv2Weights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	dv1 := cnpjCheckDigit(digits[:12], dv1Weights)
	if dv1 != int(digits[12]-'0') {
		return false
	}
	return cnpjCheckDigit(digits[:13], dv2Weights) == int(digits[13]-'0')
}

// cnpjCheckDigit computes one check digit from digits (12 or 13 long) using
// the provided weight sequence, ordered left-to-right.
func cnpjCheckDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	dv := 11 - (sum % 11)
	if dv >= 10 {
		dv = 0
	}
	return dv
}

// normalizePlaca strips separators and uppercases a plate candidate.
func normalizePlaca(placa string) string {
	var b strings.Builder
	for _, r := range placa {
		switch {
		case r == ' ' || r == '-':
			continue
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - 32)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isAlphaRune(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigitRune(r rune) bool {
	return r >= '0' && r <= '9'
}

// isValidPlaca accepts the MOC formats (XXX9999, XXX999, XX9999, XXXX999) and
// the MERCOSUL format AAA9A99, all without separators.
func isValidPlaca(placa string) bool {
	if placa == "" {
		return false
	}
	runes := []rune(placa)

	switch len(runes) {
	case 7:
		legacyStandard := allRunes(runes[0:3], isAlphaRune) && allRunes(runes[3:7], isDigitRune)
		legacyTrailer := allRunes(runes[0:4], isAlphaRune) && allRunes(runes[4:7], isDigitRune)
		mercosul := allRunes(runes[0:3], isAlphaRune) &&
			isDigitRune(runes[3]) &&
			(isAlphaRune(runes[4]) || isDigitRune(runes[4])) &&
			allRunes(runes[5:7], isDigitRune)
		return legacyStandard || legacyTrailer || mercosul
	case 6:
		return (allRunes(runes[0:3], isAlphaRune) && allRunes(runes[3:6], isDigitRune)) ||
			(allRunes(runes[0:2], isAlphaRune) && allRunes(runes[2:6], isDigitRune))
	default:
		return false
	}
}

func allRunes(runes []rune, pred func(rune) bool) bool {
	for _, r := range runes {
		if !pred(r) {
			return false
		}
	}
	return true
}

// validateDetachedTransport validates the optional transport section (MOC
// Grupo X) of the detached flow. destUF is the recipient UF and emitterUF the
// emitter UF ("" when unknown); the interstate vehicle guard implements MOC
// rule X02-20 (Rejeição 868: no veicTransp when idDest=2). Messages are
// pt-br, user-facing.
func validateDetachedTransport(in DetachedTransportInput, destUF, emitterUF string) []string {
	var errs []string

	if hasTransportadora(in) {
		doc := cleanDigits(strPtr(in.Document))
		if doc == "" {
			errs = append(errs, "Informe o CNPJ ou CPF do transportador")
		} else {
			switch len(doc) {
			case 11:
				if !isValidCPF(doc) {
					errs = append(errs, "CPF do transportador inválido")
				}
			case 14:
				if !isValidCNPJ(doc) {
					errs = append(errs, "CNPJ do transportador inválido")
				}
			default:
				errs = append(errs, "Documento do transportador deve ter 11 dígitos (CPF) ou 14 (CNPJ)")
			}
		}

		if strPtr(in.Nome) == "" {
			errs = append(errs, "Informe o nome do transportador")
		} else if len(strPtr(in.Nome)) > 60 {
			errs = append(errs, "Nome do transportador excede 60 caracteres")
		}

		if ie := strPtr(in.IE); ie != "" {
			if strPtr(in.UF) == "" {
				errs = append(errs, "Informe a UF do transportador quando a inscrição estadual for informada")
			}
			if len(ie) > 14 {
				errs = append(errs, "Inscrição estadual do transportador excede 14 caracteres")
			}
		}

		if xEnder := transportadorXender(in); len(xEnder) > 60 {
			errs = append(errs, "Endereço do transportador (logradouro e número) excede 60 caracteres")
		}

		if mun := strPtr(in.Municipio); len(mun) > 60 {
			errs = append(errs, "Município do transportador excede 60 caracteres")
		}
	}

	if strPtr(in.Placa) != "" {
		placa := normalizePlaca(strPtr(in.Placa))
		if !isValidPlaca(placa) {
			errs = append(errs, "Placa do veículo inválida: use o formato AAA9999, AAA999, AA9999, AAAA999 ou AAA9A99 (Mercosul)")
		}
		// MOC rule X02-20: vehicle groups must not be informed in interstate
		// operations (Rejeição 868).
		if destUF != "" && emitterUF != "" && strings.ToUpper(destUF) != strings.ToUpper(emitterUF) {
			errs = append(errs, "Dados do veículo não podem ser informados em operação interestadual; remova a placa")
		}
	}

	if rntc := strPtr(in.RNTC); len(rntc) > 20 {
		errs = append(errs, "ANTT (RNTRC) excede 20 caracteres")
	}

	if in.VolQtd != nil && *in.VolQtd < 1 {
		errs = append(errs, "Quantidade de volumes deve ser maior que zero")
	}
	if esp := strPtr(in.VolEsp); len(esp) > 60 {
		errs = append(errs, "Espécie dos volumes excede 60 caracteres")
	}
	if marca := strPtr(in.VolMarca); len(marca) > 60 {
		errs = append(errs, "Marca dos volumes excede 60 caracteres")
	}
	if (in.VolPesoL != nil && decPtr(in.VolPesoL).IsNegative()) ||
		(in.VolPesoB != nil && decPtr(in.VolPesoB).IsNegative()) {
		errs = append(errs, "Pesos dos volumes não podem ser negativos")
	}

	return errs
}

// mapDetachedTransport maps the parsed transport input into the XML
// entity.TransportData. modFrete is the already-resolved modality; weights
// are the item sums used by the per-field fallback. When the user provided no
// volume value the block reproduces the pre-transport auto-derivation.
func mapDetachedTransport(in DetachedTransportInput, modFrete int, emitterUF string, totalNetWeight, totalGrossWeight decimal.Decimal) entity.TransportData {
	transport := entity.TransportData{ModFrete: modFrete}

	// Transportadora (X03): mapped when the group is informed. Validation has
	// already ensured document + name presence and the document digit count.
	if hasTransportadora(in) {
		doc := cleanDigits(strPtr(in.Document))
		transportadora := entity.TransportadoraData{
			XNome:     strPtr(in.Nome),
			IE:        strPtr(in.IE),
			Endereco:  transportadorXender(in),
			Municipio: strPtr(in.Municipio),
			UF:        strPtr(in.UF),
		}
		if len(doc) == 11 {
			transportadora.Type = 2
			transportadora.CPF = doc
		} else {
			transportadora.Type = 1
			transportadora.CNPJ = doc
		}
		transport.Transportadora = &transportadora
	}

	// Veículo (X18): only mapped with an informed plate; the plate UF falls
	// back to the emitter UF (the farm's own vehicle).
	if placa := normalizePlaca(strPtr(in.Placa)); placa != "" {
		veicUF := strPtr(in.PlacaUF)
		if veicUF == "" {
			veicUF = emitterUF
		}
		transport.Veiculo = &entity.VeiculoData{
			Placa: placa,
			UF:    veicUF,
			RNTC:  strPtr(in.RNTC),
		}
	}

	// Volumes (X26): any user value switches to the user block with per-field
	// fallback; all-blank keeps today's auto-derivation.
	if hasVolValue(in) {
		block := entity.VolumeData{QVol: 1, Esp: "Granel"}
		if in.VolQtd != nil {
			block.QVol = *in.VolQtd
		}
		if esp := strPtr(in.VolEsp); esp != "" {
			block.Esp = esp
		}
		if marca := strPtr(in.VolMarca); marca != "" {
			block.Marca = marca
		}
		if in.VolPesoL != nil {
			block.PesoL = *in.VolPesoL
		} else {
			block.PesoL = totalNetWeight
		}
		if in.VolPesoB != nil {
			block.PesoB = *in.VolPesoB
		} else {
			block.PesoB = totalGrossWeight
		}
		transport.Volumes = []entity.VolumeData{block}
	} else if !totalNetWeight.IsZero() || !totalGrossWeight.IsZero() {
		transport.Volumes = []entity.VolumeData{{
			QVol:  1,
			Esp:   "Granel",
			PesoL: totalNetWeight,
			PesoB: totalGrossWeight,
		}}
	}

	return transport
}

// detachedTransportRecordFromEntity converts the mapped (validated)
// entity.TransportData into the persisted JSONB document. ok is false when no
// transport field was informed; the persistence row is skipped then.
func detachedTransportRecordFromEntity(t entity.TransportData) (nfe_model.DetachedInvoiceTransport, bool) {
	var record nfe_model.DetachedInvoiceTransport
	any := false

	if t.Transportadora != nil {
		docNumber := t.Transportadora.CNPJ
		if t.Transportadora.Type == 2 {
			docNumber = t.Transportadora.CPF
		}
		if docNumber != "" {
			record.Transportadora = &nfe_model.DetachedTransportadora{
				TpDoc:     t.Transportadora.Type,
				DocNumber: docNumber,
				XNome:     t.Transportadora.XNome,
				IE:        t.Transportadora.IE,
				XEnder:    t.Transportadora.Endereco,
				XMun:      t.Transportadora.Municipio,
				UF:        t.Transportadora.UF,
			}
			any = true
		}
	}

	if t.Veiculo != nil && t.Veiculo.Placa != "" {
		record.Veiculo = &nfe_model.DetachedVeiculo{
			Placa: t.Veiculo.Placa,
			UF:    t.Veiculo.UF,
			RNTC:  t.Veiculo.RNTC,
		}
		any = true
	}

	if len(t.Volumes) > 0 {
		v := t.Volumes[0]
		record.Vol = &nfe_model.DetachedVol{
			Qtd:   v.QVol,
			Esp:   v.Esp,
			Marca: v.Marca,
			PesoL: v.PesoL,
			PesoB: v.PesoB,
		}
		any = true
	}

	return record, any
}
