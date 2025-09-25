package vehicle_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type vehicleModel struct {
	pool *pgxpool.Pool
}

var vehicleModelImpl *vehicleModel

func InitVehicleModel(pool *pgxpool.Pool) (*vehicleModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if vehicleModelImpl == nil {
		vehicleModelImpl = &vehicleModel{
			pool: pool,
		}
	}

	return vehicleModelImpl, nil
}

func GetVehicleModel() (*vehicleModel, error) {
	if vehicleModelImpl == nil {
		return nil, errors.New("vehicle model hasnt been initialized")
	}
	return vehicleModelImpl, nil
}

func (vm *vehicleModel) AddVehicle(v entity_public.Vehicle) (entity_public.Vehicle, *model_error.ModelError) {
	var plate string
	var name string
	var farm uint32
	var id uint16

	scanErr := vm.pool.QueryRow(context.Background(), "INSERT INTO vehicle (plate, name, farm) VALUES (@plate, @name, @farm) RETURNING id, plate, name, farm", pgx.NamedArgs{"plate": v.Plate, "name": v.Name, "farm": v.Farm}).Scan(&id, &plate, &name, &farm)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.Vehicle{}, &model_error.ModelError{Message: "Já existe um veículo com esta placa"}
		}
		return entity_public.Vehicle{}, &model_error.ModelError{Message: "Falhamos ao adicionar o veículo", IsServerErr: true}
	}

	return entity_public.Vehicle{
		Id:    id,
		Plate: plate,
		Name:  name,
		Farm:  farm,
	}, nil
}

func (vm *vehicleModel) GetVehiclesByFarm(farm uint32) ([]entity_public.Vehicle, error) {
	rows, queryErr := vm.pool.Query(context.Background(), "SELECT * FROM vehicle v WHERE v.farm = @userFarm", pgx.NamedArgs{"userFarm": farm})
	if queryErr != nil {
		return []entity_public.Vehicle{}, queryErr
	}

	vehicles, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Vehicle])
	if collectErr != nil {
		return []entity_public.Vehicle{}, collectErr
	}

	return vehicles, nil
}

func (vm *vehicleModel) GetVehicle(vehicleId uint16) (entity_public.Vehicle, *model_error.ModelError) {
	var plate string
	var name string
	scanErr := vm.pool.QueryRow(context.Background(), "SELECT * FROM vehicle v WHERE v.id = @id", pgx.NamedArgs{"id": vehicleId}).Scan(&plate, &name)

	if scanErr != nil {
		return entity_public.Vehicle{}, &model_error.ModelError{Message: scanErr.Error()}
	}

	return entity_public.Vehicle{
		Id:    vehicleId,
		Plate: plate,
		Name:  name,
	}, nil
}

