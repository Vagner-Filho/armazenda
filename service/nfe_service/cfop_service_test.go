package nfe_service

import (
	"testing"

	"armazenda/model/cfop_model"
	model_error "armazenda/model/error"
)

// recordingCfopIncrementer is a hand-written mock of the model slice consumed
// by the emission counting hook (pattern of service/entry_service/test/mocks.go).
type recordingCfopIncrementer struct {
	calls    []incrementCall
	failWith *model_error.ModelError
}

type incrementCall struct {
	farmID uint32
	codes  []string
}

func (m *recordingCfopIncrementer) IncrementFarmCfopUse(farmID uint32, codes []string) *model_error.ModelError {
	recorded := make([]string, len(codes))
	copy(recorded, codes)
	m.calls = append(m.calls, incrementCall{farmID: farmID, codes: recorded})
	return m.failWith
}

// The real model must satisfy the interface the hook consumes.
var _ farmCfopUseIncrementer = (*cfop_model.CfopModel)(nil)

func TestValidSelectorCFOP(t *testing.T) {
	valid := []string{"1101", "2202", "3350", "5101", "6102", "7101", "7999"}
	for _, code := range valid {
		if !ValidSelectorCFOP(code) {
			t.Errorf("expected %q to be a valid selector CFOP", code)
		}
	}

	invalid := []string{"", "510", "51011", "51A1", "0101", "4101", "8101", "9101", " 5101"}
	for _, code := range invalid {
		if ValidSelectorCFOP(code) {
			t.Errorf("expected %q to be rejected", code)
		}
	}
}

func TestOriginDestinationForCFOP(t *testing.T) {
	cases := map[string]string{
		"1101": CfopOriginSameState,
		"5102": CfopOriginSameState,
		"2202": CfopOriginOtherState,
		"6202": CfopOriginOtherState,
		"3301": CfopOriginForeign,
		"7101": CfopOriginForeign,
		"":     "",
		"4101": "",
	}
	for code, want := range cases {
		if got := OriginDestinationForCFOP(code); got != want {
			t.Errorf("OriginDestinationForCFOP(%q) = %q, want %q", code, got, want)
		}
	}
}

func TestDistinctCFOPs(t *testing.T) {
	got := distinctCFOPs([]string{"5101", "5101", "", " 6102 ", "5101", "6102"})
	want := []string{"5101", "6102"}
	if len(got) != len(want) {
		t.Fatalf("distinctCFOPs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("distinctCFOPs = %v, want %v", got, want)
		}
	}
}

func TestDetachedItemCFOPs(t *testing.T) {
	items := []DetachedItemInput{{CFOP: "5101"}, {CFOP: "5101"}, {CFOP: "6102"}}
	got := detachedItemCFOPs(items)
	if len(got) != 3 || got[0] != "5101" || got[1] != "5101" || got[2] != "6102" {
		t.Fatalf("detachedItemCFOPs = %v", got)
	}
}

// One emission action with ["5101","5101","6102"] bumps 5101 and 6102 once
// each (dedup) in a single model call.
func TestCountDistinctFarmCfopUse_DedupsAndCallsOnce(t *testing.T) {
	mock := &recordingCfopIncrementer{}

	countDistinctFarmCfopUse(mock, 7, []string{"5101", "5101", "6102"})

	if len(mock.calls) != 1 {
		t.Fatalf("expected exactly 1 increment call, got %d", len(mock.calls))
	}
	call := mock.calls[0]
	if call.farmID != 7 {
		t.Errorf("farmID = %d, want 7", call.farmID)
	}
	if len(call.codes) != 2 || call.codes[0] != "5101" || call.codes[1] != "6102" {
		t.Errorf("codes = %v, want [5101 6102]", call.codes)
	}
}

func TestCountDistinctFarmCfopUse_EmptyInputSkips(t *testing.T) {
	mock := &recordingCfopIncrementer{}
	countDistinctFarmCfopUse(mock, 1, []string{"", "  "})
	if len(mock.calls) != 0 {
		t.Fatalf("expected no increment call for empty input, got %d", len(mock.calls))
	}
}

// Increment failures are logged only: the helper never panics and the caller
// (BuildDetachedInvoice) never toasts or blocks on them.
func TestCountDistinctFarmCfopUse_ErrorOnlyLogs(t *testing.T) {
	mock := &recordingCfopIncrementer{failWith: &model_error.ModelError{Message: "boom", IsServerErr: true}}
	countDistinctFarmCfopUse(mock, 1, []string{"5101"})
	if len(mock.calls) != 1 {
		t.Fatalf("expected the increment to be attempted, got %d calls", len(mock.calls))
	}
}
