package analysis_model

import (
	entity_public "armazenda/entity/public"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type analysisModel struct {
	pool *pgxpool.Pool
}

var analysisModelImpl *analysisModel

func InitAnalysisModel(pool *pgxpool.Pool) (*analysisModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if analysisModelImpl == nil {
		analysisModelImpl = &analysisModel{
			pool: pool,
		}
	}

	return analysisModelImpl, nil
}

func GetAnalysisModel() *analysisModel {
	if analysisModelImpl == nil {
		panic("analysis model hasnt been initialized")
	}
	return analysisModelImpl
}

func (am *analysisModel) GetNominalMostProductiveField(farmID uint32) (*entity_public.ProductiveField, error) {
	query := `
		SELECT
			f.name,
			SUM(e.netweight) as total_weight
		FROM
			entry e
		JOIN
			field f ON e.field = f.id
		LEFT JOIN
			inactive_entry ie ON e.id = ie.entry_id
		WHERE
			e.farm = $1 AND ie.entry_id IS NULL
		GROUP BY
			f.name
		ORDER BY
			total_weight DESC
		LIMIT 1;`

	row := am.pool.QueryRow(context.Background(), query, farmID)

	field := entity_public.ProductiveField{}

	if err := row.Scan(&field.Name, &field.Productivity); err != nil {
		return nil, err
	}

	return &field, nil
}

func (am *analysisModel) GetRelativeMostProductiveField(farmID uint32) (*entity_public.ProductiveField, error) {
	query := `
		SELECT
			f.name,
			SUM(e.netweight) / f.hectares as productivity
		FROM
			entry e
		JOIN
			field f ON e.field = f.id
		LEFT JOIN
			inactive_entry ie ON e.id = ie.entry_id
		WHERE
			e.farm = $1 AND ie.entry_id IS NULL
		GROUP BY
			f.name, f.hectares
		ORDER BY
			productivity DESC
		LIMIT 1;`

	row := am.pool.QueryRow(context.Background(), query, farmID)

	field := entity_public.ProductiveField{}

	if err := row.Scan(&field.Name, &field.Productivity); err != nil {
		return nil, err
	}

	return &field, nil
}
