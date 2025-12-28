package user_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type userModel struct {
	pool *pgxpool.Pool
}

var userModelImpl *userModel

func InitUserModel(pool *pgxpool.Pool) (*userModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if userModelImpl == nil {
		userModelImpl = &userModel{
			pool: pool,
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
	rows, queryErr := um.pool.Query(context.Background(),
		`SELECT u.id, u.email, u.name, u.passwd, u.inscricao_estadual, u.farm, u.cpf, u.role FROM app_user u
		 LEFT JOIN inactive_user iu ON u.id = iu.user_id
		 WHERE u.cpf = @cpf AND iu.user_id IS NULL`,
		pgx.NamedArgs{"cpf": cpf})

	if queryErr != nil {
		return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
	}

	user, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.User])
	if collectErr != nil {
		if errors.Is(pgx.ErrNoRows, collectErr) {
			return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
		}
		return nil, &model_error.ModelError{Message: "Cpf ou senha inválidos"}
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

	createFarmErr := um.pool.QueryRow(context.Background(), `INSERT INTO farm (inscricao_estadual) VALUES (@inscricao_estadual) RETURNING id`, pgx.NamedArgs{"inscricao_estadual": user.InscricaoEstadual}).Scan(&farmId)

	if createFarmErr != nil {
		return false, createFarmErr
	}

	var userCount int
	countErr := um.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM app_user WHERE farm = @farmId`, pgx.NamedArgs{"farmId": farmId}).Scan(&userCount)

	if countErr != nil {
		return false, countErr
	}

	role := "user"
	if userCount == 0 {
		role = "admin"
	}

	_, err := um.pool.Exec(context.Background(), `INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf, role) VALUES (@email, @name, @passwd, @inscricao_estadual, @farm, @cpf, @role)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": string(enc), "inscricao_estadual": user.InscricaoEstadual, "farm": farmId, "cpf": user.Cpf, "role": role})

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

	_, err := um.pool.Exec(context.Background(), `INSERT INTO user_approval (email, name, passwd, inscricao_estadual, farm_id, cpf, role) VALUES (@email, @name, @passwd, @inscricao_estadual, @farm_id, @cpf, @role)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": string(enc), "inscricao_estadual": user.InscricaoEstadual, "farm_id": farmId, "cpf": user.Cpf, "role": "user"})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (um *userModel) GetUserById(userId uint32) (*entity_public.User, error) {
	rows, queryErr := um.pool.Query(context.Background(),
		`SELECT u.id, u.email, u.name, u.passwd, u.inscricao_estadual, u.farm, u.cpf, u.role FROM app_user u
		 LEFT JOIN inactive_user iu ON u.id = iu.user_id
		 WHERE u.id = @userId AND iu.user_id IS NULL`,
		pgx.NamedArgs{"userId": userId})

	if queryErr != nil {
		return nil, queryErr
	}

	user, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.User])
	if collectErr != nil {
		return nil, collectErr
	}

	return &user, nil
}

func (um *userModel) MakeAdmin(userId uint32) error {
	_, err := um.pool.Exec(context.Background(),
		`UPDATE app_user SET role = 'admin' WHERE id = @userId`,
		pgx.NamedArgs{"userId": userId})

	return err
}

func (um *userModel) RemoveAdmin(userId uint32) (*entity_public.User, error) {
	var user entity_public.User
	err := um.pool.QueryRow(context.Background(),
		`UPDATE app_user SET role = 'user' WHERE id = @userId 
         RETURNING id, email, name, passwd, inscricao_estadual, farm, cpf, role`,
		pgx.NamedArgs{"userId": userId}).Scan(
		&user.Id, &user.Email, &user.Name, &user.Passwd,
		&user.InscricaoEstadual, &user.Farm, &user.Cpf, &user.Role)
	return &user, err
}

func (um *userModel) RemoveUser(userId uint32) error {
	_, err := um.pool.Exec(context.Background(),
		`INSERT INTO inactive_user (user_id) VALUES (@userId)`,
		pgx.NamedArgs{"userId": userId})

	return err
}

func (um *userModel) IsAdmin(userId uint32) (bool, error) {
	var role string
	err := um.pool.QueryRow(context.Background(),
		`SELECT role FROM app_user WHERE id = @userId`,
		pgx.NamedArgs{"userId": userId}).Scan(&role)

	if err != nil {
		return false, err
	}

	return role == "admin", nil
}

func (um *userModel) GetAdminCount(farmId uint32) (int, error) {
	var count int
	err := um.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM app_user u
		 LEFT JOIN inactive_user iu ON u.id = iu.user_id
		 WHERE u.farm = @farmId AND u.role = 'admin' AND iu.user_id IS NULL`,
		pgx.NamedArgs{"farmId": farmId}).Scan(&count)

	return count, err
}

func (um *userModel) CanRemoveAdmin(userId uint32) (bool, error) {
	user, err := um.GetUserById(userId)
	if err != nil {
		return false, err
	}

	if user.Role != "admin" {
		return false, nil
	}

	adminCount, err := um.GetAdminCount(user.Farm)
	if err != nil {
		return false, err
	}

	return adminCount > 1, nil
}

func (um *userModel) GetUsersByFarm(farmId uint32, isAdmin bool) ([]entity_public.PendingUser, error) {
	var rows pgx.Rows
	var err error

	if isAdmin == false {
		rows, err = um.pool.Query(context.Background(),
			`SELECT u.id, u.name, u.email, u.cpf, u.role, TRUE AS is_active FROM app_user u
			LEFT JOIN inactive_user iu ON u.id = iu.user_id
			WHERE u.farm = @farmId AND iu.user_id IS NULL`,
			pgx.NamedArgs{"farmId": farmId})
	} else {
		rows, err = um.pool.Query(context.Background(),
			`SELECT u.id, u.name, u.email, u.cpf, u.role, iu.user_id IS NULL AS is_active FROM app_user u
			LEFT JOIN inactive_user iu ON u.id = iu.user_id
			WHERE u.farm = @farmId`,
			pgx.NamedArgs{"farmId": farmId})
	}

	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity_public.PendingUser, error) {
		var user entity_public.PendingUser
		err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Cpf, &user.Role, &user.IsActive)
		return user, err
	})

	return users, err
}

func (um *userModel) ActivateUser(userId uint32) error {
	_, err := um.pool.Exec(context.Background(),
		`DELETE FROM inactive_user WHERE user_id = @userId`,
		pgx.NamedArgs{"userId": userId})

	return err
}
