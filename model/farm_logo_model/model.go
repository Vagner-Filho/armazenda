package farm_logo_model

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type farmLogoModel struct {
	pool *pgxpool.Pool
}

var farmLogoModelImpl *farmLogoModel

func InitFarmLogoModel(pool *pgxpool.Pool) (*farmLogoModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}
	if farmLogoModelImpl == nil {
		farmLogoModelImpl = &farmLogoModel{pool: pool}
	}
	return farmLogoModelImpl, nil
}

func GetFarmLogoModel() *farmLogoModel {
	if farmLogoModelImpl == nil {
		panic("farm logo model hasnt been initialized")
	}
	return farmLogoModelImpl
}

func (m *farmLogoModel) GetFarmLogo(farmID uint32) ([]byte, string, error) {
	var data []byte
	var path *string
	err := m.pool.QueryRow(context.Background(), `SELECT logo_data, logo_path FROM farm_logo WHERE farm_id = @farm_id`, pgx.NamedArgs{"farm_id": farmID}).Scan(&data, &path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("failed to get farm logo: %w", err)
	}
	p := ""
	if path != nil {
		p = *path
	}
	return data, p, nil
}

func (m *farmLogoModel) UpsertFarmLogo(farmID uint32, path string, data []byte) error {
	_, err := m.pool.Exec(context.Background(), `
		INSERT INTO farm_logo (farm_id, logo_path, logo_data, updated_at)
		VALUES (@farm_id, @logo_path, @logo_data, NOW())
		ON CONFLICT (farm_id) DO UPDATE SET
			logo_path = EXCLUDED.logo_path,
			logo_data = EXCLUDED.logo_data,
			updated_at = NOW()
	`, pgx.NamedArgs{
		"farm_id":   farmID,
		"logo_path": path,
		"logo_data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert farm logo: %w", err)
	}
	return nil
}
