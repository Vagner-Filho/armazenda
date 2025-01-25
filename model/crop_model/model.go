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
)

type cropModel struct {
	conn *pgx.Conn
}

var cropModelImpl *cropModel

func InitCropModel(conn *pgx.Conn) (*cropModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if cropModelImpl == nil {
		cropModelImpl = &cropModel{
			conn: conn,
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
	scanErr := cm.conn.QueryRow(context.Background(), "INSERT INTO crop (name, startDate) VALUES (@name, @startDate) RETURNING id, name, startDate", pgx.NamedArgs{"name": c.Name, "startDate": c.StartDate}).Scan(&id, &name, &startDateAsTime)

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
	}, nil
}

func (cm *cropModel) GetCrops() ([]entity_public.Crop, error) {
	rows, queryErr := cm.conn.Query(context.Background(), "SELECT * FROM crop")
	if queryErr != nil {
		return []entity_public.Crop{}, queryErr
	}

	crops, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Crop])
	if collectErr != nil {
		return []entity_public.Crop{}, collectErr
	}

	return crops, nil
}

var crops = []entity_public.Crop{
	{
		Id:        0,
		Name:      "Safra 2024",
		StartDate: time.Now().AddDate(0, -9, -3),
	},
	{
		Id:        1,
		Name:      "Safra 2023",
		StartDate: time.Now().AddDate(-1, -5, -7),
	},
	{
		Id:        2,
		Name:      "Safra 2022",
		StartDate: time.Now().AddDate(-2, 0, -7),
	},
	{
		Id:        3,
		Name:      "Entressafra 2024",
		StartDate: time.Now(),
	},
}
