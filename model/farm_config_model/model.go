package farm_config_model

import (
	"armazenda/entity/public"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type farmConfigModel struct {
	conn *pgx.Conn
}

var farmConfigModelImpl *farmConfigModel

func InitFarmConfigModel(conn *pgx.Conn) (*farmConfigModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if farmConfigModelImpl == nil {
		farmConfigModelImpl = &farmConfigModel{
			conn: conn,
		}
	}

	return farmConfigModelImpl, nil
}

func GetFarmConfigModel() *farmConfigModel {
	if farmConfigModelImpl == nil {
		panic("farm config model hasnt been initialized")
	}
	return farmConfigModelImpl
}

func (fcm *farmConfigModel) UpsertFarmConfig(config *entity_public.Farm) error {
	query := `
		INSERT INTO farm_config (farm_id, name, street, city, state, cep, humidity_discount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (farm_id) DO UPDATE SET
			name = EXCLUDED.name,
			street = EXCLUDED.street,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			cep = EXCLUDED.cep,
			humidity_discount = EXCLUDED.humidity_discount;
	`
	_, err := fcm.conn.Exec(context.Background(), query, config.Id, config.Name, config.Street, config.City, config.State, config.Cep, config.HumidityDiscount)
	return err
}

func (fcm *farmConfigModel) GetFarmConfig(farmID uint32) (*entity_public.Farm, error) {
	query := `
		SELECT fc.farm_id, fc.name, street, city, state, cep, humidity_discount
		FROM farm_config
		WHERE farm_id = $1;
	`
	rows, _ := fcm.conn.Query(context.Background(), query, farmID)
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[entity_public.Farm])
	if err != nil {
		return nil, fmt.Errorf("failed to get farm config: %w", err)
	}
	result.Id = farmID
	return &result, nil
}
