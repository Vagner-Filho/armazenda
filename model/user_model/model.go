package user_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

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

func (um *userModel) AuthUser(cpf string, passwd string) (*entity_public.User, *model_error.ModelError) {
	rows, queryErr := um.conn.Query(context.Background(),
		`SELECT * FROM app_user WHERE cpf = @cpf`,
		pgx.NamedArgs{"cpf": cpf})

	if queryErr != nil {
		model_error.Logger(um.conn, queryErr.Error())
		return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
	}

	user, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.User])
	if collectErr != nil {
		model_error.Logger(um.conn, collectErr.Error())
		if errors.Is(pgx.ErrNoRows, collectErr) {
			return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
		}
	}

	failed := bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(passwd))
	if failed != nil {
		return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
	}

	return &user, nil
}

func (um *userModel) CreateUser(user entity_public.NewUser) (bool, error) {
	enc, encErr := bcrypt.GenerateFromPassword([]byte(user.Passwd), 10)

	if encErr != nil {
		return false, encErr
	}

	var farmId uint32

	createFarmErr := um.conn.QueryRow(context.Background(), `INSERT INTO farm (inscricao_estadual) VALUES (@inscricao_estadual) RETURNING id`, pgx.NamedArgs{"inscricao_estadual": user.InscricaoEstadual}).Scan(&farmId)

	if createFarmErr != nil {
		return false, createFarmErr
	}

	_, err := um.conn.Exec(context.Background(), `INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf) VALUES (@email, @name, @passwd, @inscricao_estadual, @farm, @cpf)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": string(enc), "inscricao_estadual": user.InscricaoEstadual, "farm": farmId, "cpf": user.Cpf})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (um *userModel) CreateUserApproval(user entity_public.NewUser, farmId uint32) (bool, error) {
	enc, encErr := bcrypt.GenerateFromPassword([]byte(user.Passwd), 10)

	if encErr != nil {
		return false, encErr
	}

	_, err := um.conn.Exec(context.Background(), `INSERT INTO user_approval (email, name, passwd, inscricao_estadual, farm_id, cpf) VALUES (@email, @name, @passwd, @inscricao_estadual, @farm_id, @cpf)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": string(enc), "inscricao_estadual": user.InscricaoEstadual, "farm_id": farmId, "cpf": user.Cpf})

	if err != nil {
		return false, err
	}

	return true, nil
}
