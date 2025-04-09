package user_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
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

func (um *userModel) AuthUser(email string, passwd string) (*entity_public.User, *model_error.ModelError) {
	rows, queryErr := um.conn.Query(context.Background(),
		`SELECT * FROM app_user WHERE email = @email`,
		pgx.NamedArgs{"email": email})

	if queryErr != nil {
		model_error.Logger(um.conn, queryErr.Error())
		return nil, &model_error.ModelError{Message: "Email ou senha inválidos"}
	}

	user, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.User])
	fmt.Printf("\ncollectErr: %v", collectErr)
	if collectErr != nil {
		model_error.Logger(um.conn, collectErr.Error())
		if errors.Is(pgx.ErrNoRows, collectErr) {
			return nil, &model_error.ModelError{Message: "Email ou senha inválidos"}
		}
	}

	failed := bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(passwd))
	if failed != nil {
		return nil, &model_error.ModelError{Message: "Email ou senha inválidos"}
	}

	return &user, nil
}

func (um *userModel) CreateUser(user entity_public.NewUser) (bool, error) {
	enc, encErr := bcrypt.GenerateFromPassword([]byte(user.Passwd), 10)

	if encErr != nil {
		return false, encErr
	}

	_, err := um.conn.Exec(context.Background(), `INSERT INTO app_user (email, name, passwd, inscricao_estadual) VALUES (@email, @name, @passwd, @inscricao_estadual)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": string(enc), "inscricao_estadual": user.InscricaoEstadual})

	if err != nil {
		return false, err
	}

	return true, nil
}
