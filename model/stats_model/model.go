package stats_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsModel struct {
	pool *pgxpool.Pool
}

var statsModelImpl *StatsModel

func InitStatsModel(pool *pgxpool.Pool) (*StatsModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if statsModelImpl == nil {
		statsModelImpl = &StatsModel{
			pool: pool,
		}
	}

	return statsModelImpl, nil
}

func GetStatsModel() *StatsModel {
	if statsModelImpl == nil {
		panic("\nstats model hasnt been initialized\n")
	}
	return statsModelImpl
}

func (sm *StatsModel) GetTopSupplier(farmId uint32) (entity_public.StatCard, *model_error.ModelError) {
	var personName string
	var totalWeight float64

	stmt := `
		SELECT
			COALESCE(np.name, lp.companyname) as name,
			SUM(e.netweight) as total_weight
		FROM entry e
		JOIN entry_origin eo ON e.id = eo.entry_id
		JOIN person p ON eo.person_id = p.id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		WHERE e.farm = @farmId
		GROUP BY COALESCE(np.name, lp.companyname)
		ORDER BY total_weight DESC
		LIMIT 1;
	`
	err := sm.pool.QueryRow(context.Background(), stmt, pgx.NamedArgs{"farmId": farmId}).Scan(&personName, &totalWeight)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity_public.StatCard{
					Title:      "Maior Fornecedor",
					Value:      "N/A",
					PersonName: "Nenhum registro",
				},
				nil
		}
		return entity_public.StatCard{}, &model_error.ModelError{Message: "Error fetching top supplier", IsServerErr: true}
	}

	return entity_public.StatCard{
		Title:      "Maior Fornecedor (Kg)",
		Value:      fmt.Sprintf("%.2f", totalWeight),
		PersonName: personName,
		IsWeight:   true,
		Type:       "top_supplier",
	}, nil
}

func (sm *StatsModel) GetTopBuyer(farmId uint32) (entity_public.StatCard, *model_error.ModelError) {
	var personName string
	var totalWeight float64

	stmt := `
		SELECT
			COALESCE(np.name, lp.companyname) as name,
			SUM(d.netweight) as total_weight
		FROM departure d
		JOIN departure_recipient dr ON d.id = dr.departure_id
		JOIN person p ON dr.person_id = p.id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		WHERE d.farm = @farmId
		GROUP BY COALESCE(np.name, lp.companyname)
		ORDER BY total_weight DESC
		LIMIT 1;
	`
	err := sm.pool.QueryRow(context.Background(), stmt, pgx.NamedArgs{"farmId": farmId}).Scan(&personName, &totalWeight)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity_public.StatCard{
					Title:      "Maior Comprador",
					Value:      "N/A",
					PersonName: "Nenhum registro",
				},
				nil
		}
		return entity_public.StatCard{}, &model_error.ModelError{Message: "Error fetching top buyer", IsServerErr: true}
	}

	return entity_public.StatCard{
		Title:      "Maior Comprador (Kg)",
		Value:      fmt.Sprintf("%.2f", totalWeight),
		PersonName: personName,
		IsWeight:   true,
		Type:       "top_buyer",
	}, nil
}

func (sm *StatsModel) GetMostFrequentSupplier(farmId uint32) (entity_public.StatCard, *model_error.ModelError) {
	var personName string
	var deliveryCount int

	stmt := `
		SELECT
			COALESCE(np.name, lp.companyname) as name,
			COUNT(e.id) as delivery_count
		FROM entry e
		JOIN entry_origin eo ON e.id = eo.entry_id
		JOIN person p ON eo.person_id = p.id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		WHERE e.farm = @farmId
		GROUP BY COALESCE(np.name, lp.companyname)
		ORDER BY delivery_count DESC
		LIMIT 1;
	`
	err := sm.pool.QueryRow(context.Background(), stmt, pgx.NamedArgs{"farmId": farmId}).Scan(&personName, &deliveryCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity_public.StatCard{
					Title:      "Fornecedor Mais Frequente",
					Value:      "N/A",
					PersonName: "Nenhum registro",
				},
				nil
		}
		return entity_public.StatCard{}, &model_error.ModelError{Message: "Error fetching most frequent supplier", IsServerErr: true}
	}

	return entity_public.StatCard{
		Title:      "Fornecedor Mais Frequente",
		Value:      fmt.Sprintf("%d entregas", deliveryCount),
		PersonName: personName,
		Type:       "most_frequent_supplier",
	}, nil
}

func (sm *StatsModel) GetBestQualitySupplier(farmId uint32) (entity_public.StatCard, *model_error.ModelError) {
	var personName string
	var avgHumidity float64

	stmt := `
		SELECT
			COALESCE(np.name, lp.companyname) as name,
			AVG(ea.humidity) as avg_humidity
		FROM entry_analysis ea
		JOIN entry e ON ea.entryid = e.id
		JOIN entry_origin eo ON e.id = eo.entry_id
		JOIN person p ON eo.person_id = p.id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		WHERE e.farm = @farmId AND ea.humidity IS NOT NULL
		GROUP BY COALESCE(np.name, lp.companyname)
		ORDER BY avg_humidity ASC
		LIMIT 1;
	`
	err := sm.pool.QueryRow(context.Background(), stmt, pgx.NamedArgs{"farmId": farmId}).Scan(&personName, &avgHumidity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity_public.StatCard{
					Title:      "Grão de Melhor Qualidade (Umidade)",
					Value:      "N/A",
					PersonName: "Nenhum registro",
				},
				nil
		}
		return entity_public.StatCard{}, &model_error.ModelError{Message: "Error fetching best quality supplier", IsServerErr: true}
	}

	return entity_public.StatCard{
		Title:      "Grão de Melhor Qualidade (Umidade)",
		Value:      fmt.Sprintf("%.2f%%", avgHumidity),
		PersonName: personName,
		Type:       "best_quality_supplier",
	}, nil
}

func (sm *StatsModel) GetWorstQualitySupplier(farmId uint32) (entity_public.StatCard, *model_error.ModelError) {
	var personName string
	var totalAverage float64

	stmt := `
		SELECT
			COALESCE(np.name, lp.companyname) as name,
			(AVG(ea.humidity) + AVG(ea.impurity) + AVG(ea.damage)) as total_average
		FROM entry_analysis ea
		JOIN entry e ON ea.entryid = e.id
		JOIN entry_origin eo ON e.id = eo.entry_id
		JOIN person p ON eo.person_id = p.id
		LEFT JOIN natural_person np ON p.id = np.personid
		LEFT JOIN legal_person lp ON p.id = lp.personid
		WHERE e.farm = @farmId AND ea.humidity IS NOT NULL AND ea.impurity IS NOT NULL AND ea.damage IS NOT NULL
		GROUP BY COALESCE(np.name, lp.companyname)
		ORDER BY total_average DESC
		LIMIT 1;
	`
	err := sm.pool.QueryRow(context.Background(), stmt, pgx.NamedArgs{"farmId": farmId}).Scan(&personName, &totalAverage)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity_public.StatCard{
					Title:      "Pior Qualidade (Umidade+Impureza+Avaria)",
					Value:      "N/A",
					PersonName: "Nenhum registro",
				},
				nil
		}
		return entity_public.StatCard{}, &model_error.ModelError{Message: "Error fetching worst quality supplier", IsServerErr: true}
	}

	return entity_public.StatCard{
		Title:      "Pior Qualidade (Umidade+Impureza+Avaria)",
		Value:      fmt.Sprintf("%.2f%%", totalAverage),
		PersonName: personName,
		Type:       "worst_quality_supplier",
	}, nil
}