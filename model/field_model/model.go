package field_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fieldModel struct {
	conn *pgx.Conn
}

var fieldModelImpl *fieldModel

func InitFieldModel(conn *pgx.Conn) (*fieldModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if fieldModelImpl == nil {
		fieldModelImpl = &fieldModel{
			conn: conn,
		}
	}

	return fieldModelImpl, nil
}

func GetFieldModel() (*fieldModel, error) {
	if fieldModelImpl == nil {
		return nil, errors.New("field model hasnt been initialized")
	}
	return fieldModelImpl, nil
}

func (fm *fieldModel) AddField(f entity_public.Field) (entity_public.Field, *model_error.ModelError) {
	var id uint16
	var name string

	scanErr := fm.conn.QueryRow(context.Background(), "INSERT INTO field (name) VALUES (@name) RETURNING id, name", pgx.NamedArgs{"name": f.Name}).Scan(&id, &name)

	if scanErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(scanErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.Field{}, &model_error.ModelError{Message: "Já existe um talhão com este nome"}
		}
		return entity_public.Field{}, &model_error.ModelError{Message: "Falhamos ao adicionar o talhão", IsServerErr: true}
	}

	return entity_public.Field{
		Id:   id,
		Name: name,
	}, nil
}

func (fm *fieldModel) GetFields() ([]entity_public.Field, error) {
	rows, queryErr := fm.conn.Query(context.Background(), "SELECT * FROM field")
	if queryErr != nil {
		return []entity_public.Field{}, queryErr
	}

	fields, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.Field])
	if collectErr != nil {
		return []entity_public.Field{}, collectErr
	}

	return fields, nil
}
