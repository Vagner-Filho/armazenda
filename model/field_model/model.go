package field_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type fieldModel struct {
	pool *pgxpool.Pool
}

var fieldModelImpl *fieldModel

func InitFieldModel(pool *pgxpool.Pool) (*fieldModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if fieldModelImpl == nil {
		fieldModelImpl = &fieldModel{
			pool: pool,
		}
	}

	return fieldModelImpl, nil
}

func GetFieldModel() *fieldModel {
	if fieldModelImpl == nil {
		panic("field model hasnt been initialized")
	}
	return fieldModelImpl
}

func (fm *fieldModel) AddField(f entity_public.Field) (entity_public.Field, *model_error.ModelError) {
	var id uint16
	var name string
	var farm uint32
	var ha decimal.Decimal

	scanErr := fm.pool.QueryRow(context.Background(), "INSERT INTO field (name, farm, hectares) VALUES (@name, @farm, @hectares) RETURNING id, name", pgx.NamedArgs{"name": f.Name, "farm": f.Farm, "hectares": f.Hectares}).Scan(&id, &name)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.Field{}, &model_error.ModelError{Message: "Já existe um talhão com este nome"}
		}
		return entity_public.Field{}, &model_error.ModelError{Message: "Falhamos ao adicionar o talhão", IsServerErr: true}
	}

	return entity_public.Field{
		Id:       id,
		Name:     name,
		Farm:     farm,
		Hectares: ha,
	}, nil
}

func (fm *fieldModel) GetFieldsByFarm(farm uint32) ([]entity_public.Field, error) {
	rows, queryErr := fm.pool.Query(context.Background(), "SELECT * FROM field f WHERE f.farm = @userFarm", pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		return []entity_public.Field{}, queryErr
	}

	fields, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Field])
	if collectErr != nil {
		return []entity_public.Field{}, collectErr
	}

	return fields, nil
}