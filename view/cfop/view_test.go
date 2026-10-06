package cfop_view_test

import (
	"testing"

	entity_public "armazenda/entity/public"
	cfop_view "armazenda/view/cfop"
)

func codes(options []entity_public.CfopOption) []string {
	out := make([]string, len(options))
	for i, option := range options {
		out[i] = option.Code
	}
	return out
}

func TestBuildCfopOptions_GroupsAndOrders(t *testing.T) {
	options := []entity_public.CfopOption{
		{Code: "6102"},
		{Code: "5101", UseCount: 4},
		{Code: "5202", UseCount: 1, IsFarm: true},
		{Code: "1102", UseCount: 1},
	}

	view := cfop_view.BuildCfopOptions(options, "5101")

	mostUsed := codes(view.MostUsed)
	if len(mostUsed) != 3 || mostUsed[0] != "5101" || mostUsed[1] != "1102" || mostUsed[2] != "5202" {
		t.Errorf("most-used order = %v, want [5101 1102 5202]", mostUsed)
	}
	all := codes(view.All)
	if len(all) != 4 || all[0] != "1102" || all[1] != "5101" || all[2] != "5202" || all[3] != "6102" {
		t.Errorf("all order = %v, want [1102 5101 5202 6102]", all)
	}
}

func TestBuildCfopOptions_EmptyUseBucketOmitted(t *testing.T) {
	view := cfop_view.BuildCfopOptions([]entity_public.CfopOption{{Code: "5101"}}, "5101")
	if len(view.MostUsed) != 0 {
		t.Errorf("expected no most-used group, got %v", codes(view.MostUsed))
	}
}

func TestBuildCfopOptions_PrependsMissingDefault(t *testing.T) {
	view := cfop_view.BuildCfopOptions([]entity_public.CfopOption{{Code: "6102"}}, "5101")
	all := codes(view.All)
	if len(all) != 2 || all[0] != "5101" || all[1] != "6102" {
		t.Fatalf("all order = %v, want [5101 6102]", all)
	}
	if view.All[0].Description == "" {
		t.Error("prepended default must carry a derived description")
	}
}

func TestBuildCfopOptions_BlankDefaultFallsBackToSystemDefault(t *testing.T) {
	view := cfop_view.BuildCfopOptions([]entity_public.CfopOption{{Code: "6102"}}, "")
	all := codes(view.All)
	if len(all) != 2 || all[0] != "5101" {
		t.Fatalf("all order = %v, want a prepended 5101", all)
	}
}

func TestBuildCfopOptions_ExistingDefaultNotDuplicated(t *testing.T) {
	view := cfop_view.BuildCfopOptions([]entity_public.CfopOption{
		{Code: "6102"},
		{Code: "5101", UseCount: 2},
	}, "5101")
	all := codes(view.All)
	if len(all) != 2 {
		t.Fatalf("default must not be duplicated: %v", all)
	}
}
