package service_test

import (
	"testing"

	"armazenda/pkg/nfe/service"
)

func TestCode128CCheckDigit_WorkedExample(t *testing.T) {
	// MOC 7.0 Anexo II §2.1: payload "09758364" → DV = 48
	dv, err := service.Code128CCheckDigit("09758364")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dv != 48 {
		t.Errorf("DV = %d, want 48 (per MOC §2.1 worked example)", dv)
	}
}

func TestCode128CCheckDigit_OddLength(t *testing.T) {
	_, err := service.Code128CCheckDigit("123")
	if err == nil {
		t.Fatal("expected error for odd-length payload, got nil")
	}
}

func TestCode128CCheckDigit_NonNumeric(t *testing.T) {
	_, err := service.Code128CCheckDigit("12AB")
	if err == nil {
		t.Fatal("expected error for non-numeric payload, got nil")
	}
}

func TestCode128CModuleCount_44DigitKey(t *testing.T) {
	// 44-digit access key → 22 data pairs.
	// Total = quiet(20) + start(11) + 22×11(242) + DV(11) + stop(13) = 297
	count, err := service.Code128CModuleCount("51250312345678000190550010000001231234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 297 {
		t.Errorf("module count = %d, want 297 (per MOC §2)", count)
	}
}

func TestCode128CModuleCount_36DigitDadosNFe(t *testing.T) {
	// 36-digit "Dados da NF-e" → 18 data pairs.
	// Total = quiet(20) + start(11) + 18×11(198) + DV(11) + stop(13) = 253
	count, err := service.Code128CModuleCount("123456789012345678901234567890123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 253 {
		t.Errorf("module count = %d, want 253", count)
	}
}

func TestCode128CBarcode_ValidPNG(t *testing.T) {
	pngBytes, err := service.Code128CBarcode("51250312345678000190550010000001231234567890", 600, 80)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("expected non-empty PNG bytes")
	}
	// PNG magic bytes: 0x89 0x50 0x4E 0x47
	if pngBytes[0] != 0x89 || pngBytes[1] != 0x50 || pngBytes[2] != 0x4E || pngBytes[3] != 0x47 {
		t.Fatalf("output is not a PNG, first bytes: %x", pngBytes[:4])
	}
}

func TestCode128CBarcode_WidthTooSmall(t *testing.T) {
	_, err := service.Code128CBarcode("12345678", 10, 80)
	if err == nil {
		t.Fatal("expected error for width too small, got nil")
	}
}

func TestCode128CBarcode_OddLength(t *testing.T) {
	_, err := service.Code128CBarcode("123", 200, 80)
	if err == nil {
		t.Fatal("expected error for odd-length payload, got nil")
	}
}

func TestCode128CSymbols_ContainsStartAndDV(t *testing.T) {
	symbols, err := service.Code128CSymbols("09758364")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expected: [105(Start), 09, 75, 83, 64, 48(DV)]
	if len(symbols) != 6 {
		t.Fatalf("symbols len = %d, want 6", len(symbols))
	}
	if symbols[0] != 105 {
		t.Errorf("symbols[0] = %d, want 105 (Start C)", symbols[0])
	}
	if symbols[1] != 9 {
		t.Errorf("symbols[1] = %d, want 9", symbols[1])
	}
	if symbols[5] != 48 {
		t.Errorf("symbols[5] (DV) = %d, want 48", symbols[5])
	}
}
