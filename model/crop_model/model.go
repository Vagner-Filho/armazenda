package crop_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type cropModel struct {
	pool *pgxpool.Pool
}

var cropModelImpl *cropModel

func InitCropModel(pool *pgxpool.Pool) (*cropModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if cropModelImpl == nil {
		cropModelImpl = &cropModel{
			pool: pool,
		}
	}

	return cropModelImpl, nil
}

func GetCropModel() (*cropModel, error) {
	if cropModelImpl == nil {
		return nil, errors.New("crop model hasnt been initialized")
	}
	return cropModelImpl, nil
}

func (cm *cropModel) AddCrop(c entity_public.Crop) (entity_public.Crop, *model_error.ModelError) {
	var id uint8
	var name string
	var startDateAsTime time.Time
	var product uint8
	var farm uint32
	scanErr := cm.pool.QueryRow(context.Background(), `
		INSERT INTO crop (name, startDate, product, farm)
		VALUES (@name, @startDate, @product, @farm) RETURNING id, name, startDate, product, farm
		`, pgx.NamedArgs{"name": c.Name, "startDate": c.StartDate, "product": c.Product, "farm": c.Farm}).Scan(&id, &name, &startDateAsTime, &product, &farm)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.Crop{}, &model_error.ModelError{Message: "Já existe uma safra com este nome"}
		}

		return entity_public.Crop{}, &model_error.ModelError{Message: "Falhamos ao adicionar a safra", IsServerErr: true}
	}

	return entity_public.Crop{
		Id:        id,
		Name:      name,
		StartDate: startDateAsTime,
		Product:   product,
		Farm:      farm,
	}, nil
}

func (cm *cropModel) GetCropsByFarm(farm uint32) ([]entity_public.Crop, error) {
	rows, queryErr := cm.pool.Query(context.Background(), "SELECT * FROM crop c WHERE c.farm = @userFarm", pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		return []entity_public.Crop{}, queryErr
	}

	crops, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Crop])
	if collectErr != nil {
		return []entity_public.Crop{}, collectErr
	}

	return crops, nil
}