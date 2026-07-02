package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// code128CPatterns maps each CODE-128C symbol value (0-105) to its
// bar/space pattern [B, S, B, S, B, S] per MOC 7.0 Anexo II, Anexo III.01.
// Each pattern sums to 11 modules. Index 105 is the Start C character.
var code128CPatterns = [106][6]int{
	{2, 1, 2, 2, 2, 2}, // 00
	{2, 2, 2, 1, 2, 2}, // 01
	{2, 2, 2, 2, 2, 1}, // 02
	{1, 2, 1, 2, 2, 3}, // 03
	{1, 2, 1, 3, 2, 2}, // 04
	{1, 3, 1, 2, 2, 2}, // 05
	{1, 2, 2, 2, 1, 3}, // 06
	{1, 2, 2, 3, 1, 2}, // 07
	{1, 3, 2, 2, 1, 2}, // 08
	{2, 2, 1, 2, 1, 3}, // 09
	{2, 2, 1, 3, 1, 2}, // 10
	{2, 3, 1, 2, 1, 2}, // 11
	{1, 1, 2, 2, 3, 2}, // 12
	{1, 2, 2, 1, 3, 2}, // 13
	{1, 2, 2, 2, 3, 1}, // 14
	{1, 1, 3, 2, 2, 2}, // 15
	{1, 2, 3, 1, 2, 2}, // 16
	{1, 2, 3, 2, 2, 1}, // 17
	{2, 2, 3, 2, 1, 1}, // 18
	{2, 2, 1, 1, 3, 2}, // 19
	{2, 2, 1, 2, 3, 1}, // 20
	{2, 1, 3, 2, 1, 2}, // 21
	{2, 2, 3, 1, 1, 2}, // 22
	{3, 1, 2, 1, 3, 1}, // 23
	{3, 1, 1, 2, 2, 2}, // 24
	{3, 2, 1, 1, 2, 2}, // 25
	{3, 2, 1, 2, 2, 1}, // 26
	{3, 1, 2, 2, 1, 2}, // 27
	{3, 2, 2, 1, 1, 2}, // 28
	{3, 2, 2, 2, 1, 1}, // 29
	{2, 1, 2, 1, 2, 3}, // 30
	{2, 1, 2, 3, 2, 1}, // 31
	{2, 3, 2, 1, 2, 1}, // 32
	{1, 1, 1, 3, 2, 3}, // 33
	{1, 3, 1, 1, 2, 3}, // 34
	{1, 3, 1, 3, 2, 1}, // 35
	{1, 1, 2, 3, 1, 3}, // 36
	{1, 3, 2, 1, 1, 3}, // 37
	{1, 3, 2, 3, 1, 1}, // 38
	{2, 1, 1, 3, 1, 3}, // 39
	{2, 3, 1, 1, 1, 3}, // 40
	{2, 3, 1, 3, 1, 1}, // 41
	{1, 1, 2, 1, 3, 3}, // 42
	{1, 1, 2, 3, 3, 1}, // 43
	{1, 3, 2, 1, 3, 1}, // 44
	{1, 1, 3, 1, 2, 3}, // 45
	{1, 1, 3, 3, 2, 1}, // 46
	{1, 3, 3, 1, 2, 1}, // 47
	{3, 1, 3, 1, 2, 1}, // 48
	{2, 1, 1, 3, 3, 1}, // 49
	{2, 3, 1, 1, 3, 1}, // 50
	{2, 1, 3, 1, 1, 3}, // 51
	{2, 1, 3, 3, 1, 1}, // 52
	{2, 1, 3, 1, 3, 1}, // 53
	{3, 1, 1, 1, 2, 3}, // 54
	{3, 1, 1, 3, 2, 1}, // 55
	{3, 3, 1, 1, 2, 1}, // 56
	{3, 1, 2, 1, 1, 3}, // 57
	{3, 1, 2, 3, 1, 1}, // 58
	{3, 3, 2, 1, 1, 1}, // 59
	{3, 1, 4, 1, 1, 1}, // 60
	{2, 2, 1, 4, 1, 1}, // 61
	{4, 3, 1, 1, 1, 1}, // 62
	{1, 1, 1, 2, 2, 4}, // 63
	{1, 1, 1, 4, 2, 2}, // 64
	{1, 2, 1, 1, 2, 4}, // 65
	{1, 2, 1, 4, 2, 1}, // 66
	{1, 4, 1, 1, 2, 2}, // 67
	{1, 4, 1, 2, 2, 1}, // 68
	{1, 1, 2, 2, 1, 4}, // 69
	{1, 1, 2, 4, 1, 2}, // 70
	{1, 2, 2, 1, 1, 4}, // 71
	{1, 2, 2, 4, 1, 1}, // 72
	{1, 4, 2, 1, 1, 2}, // 73
	{1, 4, 2, 2, 1, 1}, // 74
	{2, 4, 1, 2, 1, 1}, // 75
	{2, 2, 1, 1, 1, 4}, // 76
	{4, 1, 3, 1, 1, 1}, // 77
	{2, 4, 1, 1, 1, 2}, // 78
	{1, 3, 4, 1, 1, 1}, // 79
	{1, 1, 1, 2, 4, 2}, // 80
	{1, 2, 1, 1, 4, 2}, // 81
	{1, 2, 1, 2, 4, 1}, // 82
	{1, 1, 4, 2, 1, 2}, // 83
	{1, 2, 4, 1, 1, 2}, // 84
	{1, 2, 4, 2, 1, 1}, // 85
	{4, 1, 1, 2, 1, 2}, // 86
	{4, 2, 1, 1, 1, 2}, // 87
	{4, 2, 1, 2, 1, 1}, // 88
	{2, 1, 2, 1, 4, 1}, // 89
	{2, 1, 4, 1, 2, 1}, // 90
	{4, 1, 2, 1, 2, 1}, // 91
	{1, 1, 1, 1, 4, 3}, // 92
	{1, 1, 1, 3, 4, 1}, // 93
	{1, 3, 1, 1, 4, 1}, // 94
	{1, 1, 4, 1, 1, 3}, // 95
	{1, 1, 4, 3, 1, 1}, // 96
	{4, 1, 1, 1, 1, 3}, // 97
	{4, 1, 1, 3, 1, 1}, // 98
	{1, 1, 3, 1, 4, 1}, // 99
	{1, 1, 4, 1, 3, 1}, // 100
	{3, 1, 1, 1, 4, 1}, // 101
	{4, 1, 1, 1, 3, 1}, // 102
	{2, 1, 1, 4, 1, 2}, // 103
	{2, 1, 1, 2, 1, 4}, // 104
	{2, 1, 1, 2, 3, 2}, // 105 (Start C)
}

// code128CStop is the Stop pattern (7 elements, 13 modules).
var code128CStop = [7]int{2, 3, 3, 1, 1, 1, 2}

const (
	code128CStartValue = 105
	code128CModulo     = 103
	code128CQuietZone  = 10 // modules per quiet zone (left and right)
)

// Code128CCheckDigit computes the modulo-103 check digit for a strictly
// numeric payload using CODE-128C rules per MOC 7.0 Anexo II §2.1.
//
// The payload must contain an even number of digits. The check digit is:
//   DV = (105 + Σ pairValue_i × (i+1)) mod 103
// where i is 0-indexed over data pairs.
func Code128CCheckDigit(payload string) (int, error) {
	if len(payload)%2 != 0 {
		return 0, fmt.Errorf("code128c: payload must have even length, got %d", len(payload))
	}
	pairs, err := code128CPairs(payload)
	if err != nil {
		return 0, err
	}
	sum := code128CStartValue // Start C = 105, weight 1
	for i, v := range pairs {
		sum += v * (i + 1)
	}
	return sum % code128CModulo, nil
}

// code128CPairs converts a numeric string into a slice of pair values (0-99).
func code128CPairs(payload string) ([]int, error) {
	pairs := make([]int, 0, len(payload)/2)
	for i := 0; i < len(payload); i += 2 {
		v := int(payload[i]-'0')*10 + int(payload[i+1]-'0')
		if v < 0 || v > 99 {
			return nil, fmt.Errorf("code128c: non-digit character in payload at position %d", i)
		}
		pairs = append(pairs, v)
	}
	return pairs, nil
}

// Code128CSymbols returns the full symbol sequence for a CODE-128C encoding:
// [Start C, data pairs..., DV, Stop]. The Stop is represented as value -1
// (caller should use code128CStop for its pattern).
func Code128CSymbols(payload string) ([]int, error) {
	if len(payload)%2 != 0 {
		return nil, fmt.Errorf("code128c: payload must have even length, got %d", len(payload))
	}
	for _, c := range payload {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("code128c: payload must be numeric, got %q", payload)
		}
	}
	pairs, err := code128CPairs(payload)
	if err != nil {
		return nil, err
	}
	dv, err := Code128CCheckDigit(payload)
	if err != nil {
		return nil, err
	}
	symbols := make([]int, 0, len(pairs)+3)
	symbols = append(symbols, code128CStartValue)
	symbols = append(symbols, pairs...)
	symbols = append(symbols, dv)
	return symbols, nil
}

// Code128CModuleCount returns the total number of modules (including quiet
// zones) for a CODE-128C encoding of the given payload.
func Code128CModuleCount(payload string) (int, error) {
	symbols, err := Code128CSymbols(payload)
	if err != nil {
		return 0, err
	}
	total := code128CQuietZone * 2 // left + right quiet zones
	total += 11 * len(symbols) // start + data + DV (each 11 modules)
	total += 13                 // stop is 13 modules
	return total, nil
}

// Code128CBarcode renders a strictly-numeric payload as a CODE-128C barcode
// PNG image scaled to widthPx × heightPx pixels. The image uses 1-bit black
// (bar) and white (space) pixels.
//
// Per MOC 7.0 Anexo II §2, for a 44-digit access key the minimum total width
// is 6 cm for laser/inkjet printers and the minimum bar height is 0.8 cm.
// Callers are responsible for choosing widthPx/heightPx to satisfy these
// minimums at their target DPI.
func Code128CBarcode(payload string, widthPx, heightPx int) ([]byte, error) {
	symbols, err := Code128CSymbols(payload)
	if err != nil {
		return nil, err
	}

	totalModules := code128CQuietZone * 2
	for range symbols {
		totalModules += 11
	}
	totalModules += 13 // stop

	if widthPx < totalModules {
		return nil, fmt.Errorf("code128c: width %dpx too small for %d modules", widthPx, totalModules)
	}
	if heightPx < 1 {
		return nil, fmt.Errorf("code128c: height must be >= 1px, got %d", heightPx)
	}

	// Build the module map: true = bar (black), false = space (white).
	modules := make([]bool, 0, totalModules)
	// Left quiet zone
	for i := 0; i < code128CQuietZone; i++ {
		modules = append(modules, false)
	}
	// Start C
	modules = appendModules(modules, code128CPatterns[code128CStartValue][:])
	// Data + DV
	for _, s := range symbols[1:] { // skip Start (already added)
		modules = appendModules(modules, code128CPatterns[s][:])
	}
	// Stop
	modules = appendModules(modules, code128CStop[:])
	// Right quiet zone
	for i := 0; i < code128CQuietZone; i++ {
		modules = append(modules, false)
	}

	if len(modules) != totalModules {
		return nil, fmt.Errorf("code128c: internal error, module count %d != expected %d", len(modules), totalModules)
	}

	// Render to image. Each module gets ceil(widthPx/totalModules) pixels,
	// with the last module absorbing the remainder to fill widthPx exactly.
	img := image.NewRGBA(image.Rect(0, 0, widthPx, heightPx))
	// Fill white
	for y := 0; y < heightPx; y++ {
		for x := 0; x < widthPx; x++ {
			img.Set(x, y, color.White)
		}
	}
	// Draw bars
	modulePx := widthPx / totalModules
	remainder := widthPx - modulePx*totalModules
	x := 0
	for m := 0; m < totalModules; m++ {
		w := modulePx
		if m == totalModules-1 {
			w += remainder
		}
		if modules[m] {
			for dx := 0; dx < w; dx++ {
				for dy := 0; dy < heightPx; dy++ {
					img.Set(x+dx, dy, color.Black)
				}
			}
		}
		x += w
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("code128c: failed to encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// appendModules appends bar/space modules from a pattern. Odd-indexed elements
// are spaces (false), even-indexed are bars (true).
func appendModules(dst []bool, pattern []int) []bool {
	for i, w := range pattern {
		isBar := i%2 == 0
		for j := 0; j < w; j++ {
			dst = append(dst, isBar)
		}
	}
	return dst
}
