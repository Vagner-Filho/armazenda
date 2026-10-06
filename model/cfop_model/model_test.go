package cfop_model_test

import (
	"testing"

	entity_public "armazenda/entity/public"
	"armazenda/model/cfop_model"
)

func TestGroupCfopOptions_MostUsedFirstTiesByCode(t *testing.T) {
	options := []entity_public.CfopOption{
		{Code: "6102", UseCount: 0},
		{Code: "5101", UseCount: 3},
		{Code: "5202", UseCount: 2, IsFarm: true},
		{Code: "1102", UseCount: 3},
		{Code: "5901", UseCount: 2},
	}

	mostUsed, all := cfop_model.GroupCfopOptions(options)

	wantMostUsedCodes := []string{"1102", "5101", "5202", "5901"} // 3,3 then 2,2 (code ASC)
	if len(mostUsed) != len(wantMostUsedCodes) {
		t.Fatalf("most-used len = %d, want %d (%+v)", len(mostUsed), len(wantMostUsedCodes), mostUsed)
	}
	for i, code := range wantMostUsedCodes {
		if mostUsed[i].Code != code {
			t.Errorf("most-used[%d] = %s, want %s", i, mostUsed[i].Code, code)
		}
	}

	wantAllCodes := []string{"1102", "5101", "5202", "5901", "6102"}
	if len(all) != len(wantAllCodes) {
		t.Fatalf("all len = %d, want %d (%+v)", len(all), len(wantAllCodes), all)
	}
	for i, code := range wantAllCodes {
		if all[i].Code != code {
			t.Errorf("all[%d] = %s, want %s", i, all[i].Code, code)
		}
	}
}

func TestGroupCfopOptions_NoUsageOmitsMostUsedGroup(t *testing.T) {
	mostUsed, all := cfop_model.GroupCfopOptions([]entity_public.CfopOption{
		{Code: "5102"},
		{Code: "5101"},
	})

	if len(mostUsed) != 0 {
		t.Errorf("expected an empty most-used group, got %+v", mostUsed)
	}
	if len(all) != 2 || all[0].Code != "5101" || all[1].Code != "5102" {
		t.Errorf("expected all options ordered by code, got %+v", all)
	}
}

func TestGroupCfopOptions_EmptyInput(t *testing.T) {
	mostUsed, all := cfop_model.GroupCfopOptions(nil)
	if len(mostUsed) != 0 || len(all) != 0 {
		t.Errorf("expected empty groups, got mostUsed=%+v all=%+v", mostUsed, all)
	}
}

func TestGroupCfopOptions_FarmFlagPreserved(t *testing.T) {
	mostUsed, all := cfop_model.GroupCfopOptions([]entity_public.CfopOption{
		{Code: "9999", IsFarm: true, UseCount: 1},
	})
	if !mostUsed[0].IsFarm || !all[0].IsFarm {
		t.Error("expected the farm flag to survive grouping")
	}
}
