package entity_test

import (
	"testing"

	"armazenda/pkg/nfe/entity"
)

func TestGenerateAccessKey_DifferentTpEmis(t *testing.T) {
	base := entity.AccessKeyData{
		CUF:      "51",
		AAMM:     "2507",
		Document: "12345678000195",
		Mod:      "55",
		Serie:    1,
		NNF:      42,
		TpEmis:   "1",
		CNF:      "12345678",
	}

	keyNormal := entity.GenerateAccessKey(base)
	if len(keyNormal) != 44 {
		t.Fatalf("expected 44-digit key, got %d digits: %s", len(keyNormal), keyNormal)
	}

	// Same data but with SVC-AN tpEmis (6) — must produce a different key
	baseSvcAN := base
	baseSvcAN.TpEmis = "6"
	keySvcAN := entity.GenerateAccessKey(baseSvcAN)
	if keySvcAN == keyNormal {
		t.Errorf("SVC-AN key should differ from normal key: normal=%s svc-an=%s", keyNormal, keySvcAN)
	}

	// Same data but with SVC-RS tpEmis (7) — must produce a different key
	baseSvcRS := base
	baseSvcRS.TpEmis = "7"
	keySvcRS := entity.GenerateAccessKey(baseSvcRS)
	if keySvcRS == keyNormal {
		t.Errorf("SVC-RS key should differ from normal key: normal=%s svc-rs=%s", keyNormal, keySvcRS)
	}

	// SVC-AN and SVC-RS must also differ from each other
	if keySvcAN == keySvcRS {
		t.Errorf("SVC-AN and SVC-RS keys should differ: svc-an=%s svc-rs=%s", keySvcAN, keySvcRS)
	}
}

func TestGenerateAccessKey_LengthAndDV(t *testing.T) {
	data := entity.AccessKeyData{
		CUF:      "51",
		AAMM:     "2507",
		Document: "12345678000195",
		Mod:      "55",
		Serie:    1,
		NNF:      42,
		TpEmis:   "1",
		CNF:      "12345678",
	}
	key := entity.GenerateAccessKey(data)
	if len(key) != 44 {
		t.Fatalf("expected 44-digit key, got %d", len(key))
	}
	// tpEmis should be at position 35 (0-indexed) — the 34th character in the 43-char prefix
	if key[34:35] != "1" {
		t.Errorf("expected tpEmis '1' at position 34, got '%s'", key[34:35])
	}
}
