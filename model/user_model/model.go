package user_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type userModel struct {
	conn *pgx.Conn
}

var userModelImpl *userModel

func InitUserModel(conn *pgx.Conn) (*userModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if userModelImpl == nil {
		userModelImpl = &userModel{
			conn: conn,
		}
	}

	return userModelImpl, nil
}

func GetUserModel() *userModel {
	if userModelImpl == nil {
		panic("\nuser model hasnt been initialized\n")
	}
	return userModelImpl
}

func (um *userModel) AuthUser(email string, psswd string) (*entity_public.User, *model_error.ModelError) {
	rows, queryErr := um.conn.Query(context.Background(),
		`SELECT FROM user WHERE email = @email AND psswd = @psswd`,
		pgx.NamedArgs{"email": email, "psswd": psswd})
	if queryErr != nil {
		model_error.Logger(um.conn, queryErr.Error())
		return nil, &model_error.ModelError{IsServerErr: true}
	}

	user, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.User])
	if collectErr != nil {
		if errors.Is(pgx.ErrNoRows, collectErr) {
			return nil, &model_error.ModelError{Message: "Email ou senha inválidos"}
		}
	}

	return &user, nil
}
