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

type ProductModel struct {
	pool *pgxpool.Pool
}

var productModelImpl *ProductModel

func InitProductModel(pool *pgxpool.Pool) (*ProductModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if productModelImpl == nil {
		productModelImpl = &ProductModel{
			pool: pool,
		}
	}

	return productModelImpl, nil
}

func GetProductModel() *ProductModel {
	if productModelImpl == nil {
		panic("\nproduct model hasnt been initialized\n")
	}
	return productModelImpl
}

func (pm *ProductModel) GetProducts() ([]entity_public.Product, error) {
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

func (pm *ProductModel) GetProductById(id uint8) (entity_public.Product, error) {
	rows, err := pm.pool.Query(context.Background(), "SELECT * FROM product WHERE id = @id", pgx.NamedArgs{"id": id})
	if err != nil {
		return entity_public.Product{}, &model_error.ModelError{Message: err.Error()}
	}

	product, collectErr := pgx.CollectOneRow(rows, pgx.RowToStructByPos[entity_public.Product])
	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return entity_public.Product{}, &model_error.ModelError{Message: "Produto não encontrado"}
		}
		return entity_public.Product{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return product, nil
}
