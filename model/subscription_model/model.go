package subscription_model

import (
	entity_public "armazenda/entity/public"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type subscriptionModel struct {
	pool *pgxpool.Pool
}

var subscriptionModelImpl *subscriptionModel

func InitSubscriptionModel(pool *pgxpool.Pool) (*subscriptionModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if subscriptionModelImpl == nil {
		subscriptionModelImpl = &subscriptionModel{
			pool: pool,
		}
	}

	return subscriptionModelImpl, nil
}

func GetSubscriptionModel() *subscriptionModel {
	if subscriptionModelImpl == nil {
		panic("\nsubscription model hasnt been initialized\n")
	}
	return subscriptionModelImpl
}

func (sm *subscriptionModel) CreatePendingRegistration(user entity_public.NewUser, stripeSessionID string) (uint32, error) {
	enc, encErr := bcrypt.GenerateFromPassword([]byte(user.Passwd), 10)
	if encErr != nil {
		return 0, encErr
	}

	var id uint32
	err := sm.pool.QueryRow(context.Background(),
		`INSERT INTO pending_registration (email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at)
		 VALUES (@email, @name, @passwd, @cpf, @inscricao_estadual, @role, @stripe_checkout_session_id, @created_at)
		 RETURNING id`,
		pgx.NamedArgs{
			"email":                      user.Email,
			"name":                       user.Name,
			"passwd":                     string(enc),
			"cpf":                        user.Cpf,
			"inscricao_estadual":         user.InscricaoEstadual,
			"role":                       user.Role,
			"stripe_checkout_session_id": stripeSessionID,
			"created_at":                 time.Now(),
		}).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sm *subscriptionModel) GetPendingRegistrationBySessionID(sessionID string) (*entity_public.PendingRegistration, error) {
	rows, err := sm.pool.Query(context.Background(),
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at
		 FROM pending_registration
		 WHERE stripe_checkout_session_id = @session_id`,
		pgx.NamedArgs{"session_id": sessionID})

	if err != nil {
		return nil, err
	}

	pending, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.PendingRegistration])
	if collectErr != nil {
		return nil, collectErr
	}

	return &pending, nil
}

func (sm *subscriptionModel) GetPendingRegistrationByID(id uint32) (*entity_public.PendingRegistration, error) {
	rows, err := sm.pool.Query(context.Background(),
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at
		 FROM pending_registration
		 WHERE id = @id`,
		pgx.NamedArgs{"id": id})

	if err != nil {
		return nil, err
	}

	pending, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.PendingRegistration])
	if collectErr != nil {
		return nil, collectErr
	}

	return &pending, nil
}

func (sm *subscriptionModel) UpdatePendingRegistrationSessionID(id uint32, sessionID string) error {
	_, err := sm.pool.Exec(context.Background(),
		`UPDATE pending_registration SET stripe_checkout_session_id = @session_id WHERE id = @id`,
		pgx.NamedArgs{"session_id": sessionID, "id": id})
	return err
}

func (sm *subscriptionModel) DeletePendingRegistration(id uint32) error {
	_, err := sm.pool.Exec(context.Background(),
		`DELETE FROM pending_registration WHERE id = @id`,
		pgx.NamedArgs{"id": id})
	return err
}

func (sm *subscriptionModel) CreateFarmAndUserFromPending(pending entity_public.PendingRegistration) (uint32, error) {
	ctx := context.Background()
	tx, err := sm.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var farmId uint32
	createFarmErr := tx.QueryRow(ctx,
		`INSERT INTO farm (inscricao_estadual) VALUES (@inscricao_estadual) RETURNING id`,
		pgx.NamedArgs{"inscricao_estadual": pending.InscricaoEstadual}).Scan(&farmId)

	if createFarmErr != nil {
		return 0, createFarmErr
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf, role)
		 VALUES (@email, @name, @passwd, @inscricao_estadual, @farm, @cpf, @role)`,
		pgx.NamedArgs{
			"email":              pending.Email,
			"name":               pending.Name,
			"passwd":             pending.Passwd,
			"inscricao_estadual": pending.InscricaoEstadual,
			"farm":               farmId,
			"cpf":                pending.Cpf,
			"role":               pending.Role,
		})

	if err != nil {
		return 0, err
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return 0, commitErr
	}

	return farmId, nil
}

func (sm *subscriptionModel) SetFarmStripeIDs(farmId uint32, customerID, subscriptionID string) error {
	_, err := sm.pool.Exec(context.Background(),
		`UPDATE farm SET
			stripe_customer_id = @customer_id,
			stripe_subscription_id = @subscription_id
		 WHERE id = @farm_id`,
		pgx.NamedArgs{
			"customer_id":     customerID,
			"subscription_id": subscriptionID,
			"farm_id":         farmId,
		})
	return err
}

func (sm *subscriptionModel) UpdateFarmSubscription(farmId uint32, customerID, subscriptionID, status string, periodEnd time.Time) error {
	_, err := sm.pool.Exec(context.Background(),
		`UPDATE farm SET
			stripe_customer_id = @customer_id,
			stripe_subscription_id = @subscription_id,
			subscription_status = @status,
			subscription_current_period_end = @period_end
		 WHERE id = @farm_id`,
		pgx.NamedArgs{
			"customer_id":     customerID,
			"subscription_id": subscriptionID,
			"status":          status,
			"period_end":      periodEnd,
			"farm_id":         farmId,
		})
	return err
}

func (sm *subscriptionModel) GetFarmSubscriptionStatus(farmId uint32) (string, error) {
	var status string
	err := sm.pool.QueryRow(context.Background(),
		`SELECT COALESCE(subscription_status, 'active') FROM farm WHERE id = @farm_id`,
		pgx.NamedArgs{"farm_id": farmId}).Scan(&status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "active", nil
		}
		return "", err
	}

	return status, nil
}

func (sm *subscriptionModel) GetFarmByStripeSubscriptionID(subscriptionID string) (uint32, error) {
	var farmId uint32
	err := sm.pool.QueryRow(context.Background(),
		`SELECT id FROM farm WHERE stripe_subscription_id = @subscription_id`,
		pgx.NamedArgs{"subscription_id": subscriptionID}).Scan(&farmId)

	if err != nil {
		return 0, err
	}

	return farmId, nil
}
