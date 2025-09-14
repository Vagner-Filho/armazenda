package product_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productModel struct {
	pool *pgxpool.Pool
}

var productModelImpl *productModel

func InitProductModel(pool *pgxpool.Pool) (*productModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if productModelImpl == nil {
		productModelImpl = &productModel{
			pool: pool,
		}
	}

	return productModelImpl, nil
}

func GetProductModel() (*productModel, error) {
	if productModelImpl == nil {
		return nil, errors.New("product model hasnt been initialized")
	}
	return productModelImpl, nil
}

func (pm *productModel) GetProducts() ([]entity_public.Product, error) {
	rows, err := pm.pool.Query(context.Background(), "SELECT * FROM product")
	if err != nil {
		return []entity_public.Product{}, &model_error.ModelError{Message: err.Error()}
	}

	products, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Product])
	if collectErr != nil {
		fmt.Printf("\ncollectErr: %v\n", collectErr.Error())
		return []entity_public.Product{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return products, nil
}