package stats

import (
	entity_public "armazenda/entity/public"
	"html/template"
	"os"
	"testing"
)

func TestBuildFieldProductSeries(t *testing.T) {
	weightMetric := func(r entity_public.FieldProductTotal) float64 { return r.TotalWeight }

	t.Run("groups rows by field with one dataset per product", func(t *testing.T) {
		rows := []entity_public.FieldProductTotal{
			{FieldName: "Talhão B", ProductName: "Soja", TotalWeight: 67467},
			{FieldName: "Talhão A", ProductName: "Milho", TotalWeight: 25000},
			{FieldName: "Talhão A", ProductName: "Soja", TotalWeight: 12000},
		}

		series := buildFieldProductSeries(rows, weightMetric)

		if len(series.Labels) != 2 || series.Labels[0] != "Talhão B" || series.Labels[1] != "Talhão A" {
			t.Errorf("unexpected labels: %v", series.Labels)
		}
		if len(series.Datasets) != 2 {
			t.Fatalf("expected 2 datasets, got %d", len(series.Datasets))
		}
		soja := series.Datasets[0]
		if soja.Product != "Soja" || len(soja.Values) != 2 || soja.Values[0] != 67467 || soja.Values[1] != 12000 {
			t.Errorf("unexpected Soja dataset: %+v", soja)
		}
		milho := series.Datasets[1]
		if milho.Product != "Milho" || len(milho.Values) != 2 || milho.Values[0] != 0 || milho.Values[1] != 25000 {
			t.Errorf("unexpected Milho dataset: %+v", milho)
		}
	})

	t.Run("empty rows produce empty series", func(t *testing.T) {
		series := buildFieldProductSeries(nil, weightMetric)
		if len(series.Labels) != 0 || len(series.Datasets) != 0 {
			t.Errorf("expected empty series, got %+v", series)
		}
	})
}

// TestAnalysisTemplatesParse guards against template syntax errors, which
// would otherwise only surface on server startup.
func TestAnalysisTemplatesParse(t *testing.T) {
	entries, err := os.ReadDir("../../templates/analysis")
	if err != nil {
		t.Fatalf("failed to read templates dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		_, err := template.ParseFiles("../../templates/analysis/" + entry.Name())
		if err != nil {
			t.Errorf("template %s failed to parse: %v", entry.Name(), err)
		}
	}
}
