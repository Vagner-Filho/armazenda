package vehicle_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type vehicleModel struct {
	conn *pgx.Conn
}

var vehicleModelImpl *vehicleModel

func InitVehicleModel(conn *pgx.Conn) (*vehicleModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if vehicleModelImpl == nil {
		vehicleModelImpl = &vehicleModel{
			conn: conn,
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

	scanErr := vm.conn.QueryRow(context.Background(), "INSERT INTO vehicle (plate, name) VALUES (@plate, @name) RETURNING plate, name", pgx.NamedArgs{"plate": v.Plate, "name": v.Name}).Scan(&plate, &name)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.Vehicle{}, &model_error.ModelError{Message: "Já existe um veículo com esta placa"}
		}
		return entity_public.Vehicle{}, &model_error.ModelError{Message: "Falhamos ao adicionar o veículo", IsServerErr: true}
	}

	return entity_public.Vehicle{
		Plate: plate,
		Name:  name,
	}, nil
}

func (vm *vehicleModel) GetVehicles() ([]entity_public.Vehicle, error) {
	rows, queryErr := vm.conn.Query(context.Background(), "SELECT * FROM vehicle")
	if queryErr != nil {
		return []entity_public.Vehicle{}, queryErr
	}

	vehicles, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Vehicle])
	if collectErr != nil {
		return []entity_public.Vehicle{}, collectErr
	}

	return vehicles, nil
}

func (vm *vehicleModel) GetVehicle(plateParam string) (entity_public.Vehicle, *model_error.ModelError) {
	var plate string
	var name string
	scanErr := vm.conn.QueryRow(context.Background(), "SELECT * FROM vehicle WHERE vehicle.plate = @plate", pgx.NamedArgs{"plate": plateParam}).Scan(&plate, &name)

	if scanErr != nil {
		return entity_public.Vehicle{}, &model_error.ModelError{Message: scanErr.Error()}
	}

	return entity_public.Vehicle{
		Plate: plate,
		Name:  name,
	}, nil
}
