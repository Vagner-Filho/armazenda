package user_approval_model

import (
	"armazenda/entity/public"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userApprovalModel struct {
	pool *pgxpool.Pool
}

var userApprovalModelImpl *userApprovalModel

func InitUserApprovalModel(pool *pgxpool.Pool) (*userApprovalModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if userApprovalModelImpl == nil {
		userApprovalModelImpl = &userApprovalModel{
			pool: pool,
		}
	}

	return userApprovalModelImpl, nil
}

func GetUserApprovalModel() *userApprovalModel {
	if userApprovalModelImpl == nil {
		panic("\nuser approval model hasnt been initialized\n")
	}
	return userApprovalModelImpl
}

func (uam *userApprovalModel) GetPendingUsersByFarm(farmId uint32) ([]entity_public.PendingUser, error) {
	rows, err := uam.pool.Query(context.Background(), `SELECT id, name, email, cpf, role, TRUE FROM user_approval WHERE farm_id = @farm_id AND status = 'pending'`, pgx.NamedArgs{"farm_id": farmId})
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.PendingUser])
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (uam *userApprovalModel) ApproveUser(userId uint32) error {
	// Get user from approval table
	rows, err := uam.pool.Query(context.Background(), `SELECT id, email, name, passwd, inscricao_estadual, farm_id, cpf, role, status FROM user_approval WHERE id = @id`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return err
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.UserApproval])
	if err != nil {
		return err
	}

	// Insert user into app_user table
	_, err = uam.pool.Exec(context.Background(), `INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf, role) VALUES (@email, @name, @passwd, @inscricao_estadual, @farm, @cpf, @role)`, pgx.NamedArgs{"email": user.Email, "name": user.Name, "passwd": user.Passwd, "inscricao_estadual": user.InscricaoEstadual, "farm": user.Farm, "cpf": user.Cpf, "role": user.Role})
	if err != nil {
		return err
	}

	// Update status in user_approval table
	_, err = uam.pool.Exec(context.Background(), `UPDATE user_approval SET status = 'approved' WHERE id = @id`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return err
	}

	return nil
}

func (uam *userApprovalModel) DeclineUser(userId uint32) error {
	_, err := uam.pool.Exec(context.Background(), `UPDATE user_approval SET status = 'declined' WHERE id = @id`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return err
	}

	return nil
}
