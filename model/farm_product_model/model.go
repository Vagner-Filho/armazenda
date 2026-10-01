package farm_product_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FarmProductModel struct {
	pool *pgxpool.Pool
}

var farmProductModelImpl *FarmProductModel

func InitFarmProductModel(pool *pgxpool.Pool) (*FarmProductModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if farmProductModelImpl == nil {
		farmProductModelImpl = &FarmProductModel{
			pool: pool,
		}
	}

	return farmProductModelImpl, nil
}

func GetFarmProductModel() *FarmProductModel {
	if farmProductModelImpl == nil {
		panic("\nfarm product model hasnt been initialized\n")
	}
	return farmProductModelImpl
}

func (fm *FarmProductModel) GetFarmProductsByFarm(farmId uint32) ([]entity_public.FarmProduct, error) {
	rows, err := fm.pool.Query(context.Background(), `
		SELECT id, name, ncm, farm_id FROM farm_product
		WHERE farm_id = @farmId
		ORDER BY name
	`, pgx.NamedArgs{"farmId": farmId})
	if err != nil {
		model_error.GetLoggerModel().Log(err.Error())
		return []entity_public.FarmProduct{}, &model_error.ModelError{Message: err.Error()}
	}

	farmProducts, collectErr := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity_public.FarmProduct, error) {
		var product entity_public.Product
		var farmId uint32
		scanErr := row.Scan(&product.Id, &product.Name, &product.NCM, &farmId)
		if scanErr != nil {
			return entity_public.FarmProduct{}, scanErr
		}
		return entity_public.FarmProduct{
			Product: product,
			FarmId:  farmId,
		}, nil
	})
	if collectErr != nil {
		fmt.Printf("\ncollectErr: %v\n", collectErr.Error())
		return []entity_public.FarmProduct{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return farmProducts, nil
}

func (fm *FarmProductModel) GetFarmProductById(id uint8, farmId uint32) (entity_public.FarmProduct, error) {
	rows, err := fm.pool.Query(context.Background(), `
		SELECT id, name, ncm, farm_id FROM farm_product
		WHERE id = @id AND farm_id = @farmId
	`, pgx.NamedArgs{"id": id, "farmId": farmId})
	if err != nil {
		model_error.GetLoggerModel().Log(err.Error())
		return entity_public.FarmProduct{}, &model_error.ModelError{Message: err.Error()}
	}

	farmProducts, collectErr := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (entity_public.FarmProduct, error) {
		var product entity_public.Product
		var farmId uint32
		scanErr := row.Scan(&product.Id, &product.Name, &product.NCM, &farmId)
		if scanErr != nil {
			return entity_public.FarmProduct{}, scanErr
		}
		return entity_public.FarmProduct{
			Product: product,
			FarmId:  farmId,
		}, nil
	})
	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Produto não encontrado"}
		}
		model_error.GetLoggerModel().Log(collectErr.Error())
		return entity_public.FarmProduct{}, &model_error.ModelError{Message: collectErr.Error()}
	}

	return farmProducts, nil
}

// AddFarmProduct inserts a farm-created product. The per-farm name
// uniqueness and the anti-shadow rules against the global catalog are
// enforced by the unique_product_name_in_farm index and the
// forbid_farm_product_global_overlap trigger.
func (fm *FarmProductModel) AddFarmProduct(fp entity_public.FarmProduct) (entity_public.FarmProduct, *model_error.ModelError) {
	var id uint8
	var name string
	var ncm string
	var farmId uint32

	scanErr := fm.pool.QueryRow(context.Background(), `
		INSERT INTO farm_product (name, ncm, farm_id)
		VALUES (@name, @ncm, @farmId)
		RETURNING id, name, ncm, farm_id
	`, pgx.NamedArgs{"name": fp.Name, "ncm": fp.NCM, "farmId": fp.FarmId}).
		Scan(&id, &name, &ncm, &farmId)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Já existe um produto com este nome nesta fazenda"}
			case pgerrcode.RaiseException:
				return entity_public.FarmProduct{}, &model_error.ModelError{Message: pgErr.Message}
			}
		}

		model_error.GetLoggerModel().Log(scanErr.Error())
		return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Falhamos ao adicionar o produto", IsServerErr: true}
	}

	return entity_public.FarmProduct{
		Product: entity_public.Product{Id: id, Name: name, NCM: ncm},
		FarmId:  farmId,
	}, nil
}

// UpdateFarmProduct renames/re-NCMs a farm-created product, scoped to the farm.
func (fm *FarmProductModel) UpdateFarmProduct(fp entity_public.FarmProduct) (entity_public.FarmProduct, *model_error.ModelError) {
	var id uint8
	var name string
	var ncm string
	var farmId uint32

	scanErr := fm.pool.QueryRow(context.Background(), `
		UPDATE farm_product
		SET name = @name, ncm = @ncm
		WHERE id = @id AND farm_id = @farmId
		RETURNING id, name, ncm, farm_id
	`, pgx.NamedArgs{"name": fp.Name, "ncm": fp.NCM, "id": fp.Id, "farmId": fp.FarmId}).
		Scan(&id, &name, &ncm, &farmId)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Já existe um produto com este nome nesta fazenda"}
			case pgerrcode.RaiseException:
				return entity_public.FarmProduct{}, &model_error.ModelError{Message: pgErr.Message}
			}
		}

		if errors.Is(scanErr, pgx.ErrNoRows) {
			return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Produto não encontrado"}
		}

		model_error.GetLoggerModel().Log(scanErr.Error())
		return entity_public.FarmProduct{}, &model_error.ModelError{Message: "Falhamos ao atualizar o produto", IsServerErr: true}
	}

	return entity_public.FarmProduct{
		Product: entity_public.Product{Id: id, Name: name, NCM: ncm},
		FarmId:  farmId,
	}, nil
}
