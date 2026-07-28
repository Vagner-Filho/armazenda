package subscription_model

import (
	entity_public "armazenda/entity/public"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
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
		`INSERT INTO pending_registration (email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at, owner_document, owner_document_type, additional_ies, uf)
		 VALUES (@email, @name, @passwd, @cpf, @inscricao_estadual, @role, @stripe_checkout_session_id, @created_at, @owner_document, @owner_document_type, @additional_ies, @uf)
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
			"owner_document":             user.OwnerDocument,
			"owner_document_type":        user.OwnerDocumentType,
			"additional_ies":             user.AdditionalIEs,
			"stripe_price_id":            user.PriceID,
			"uf":                         user.UF,
		}).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sm *subscriptionModel) GetPendingRegistrationBySessionID(sessionID string) (*entity_public.PendingRegistration, error) {
	rows, err := sm.pool.Query(context.Background(),
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at, owner_document, owner_document_type, additional_ies, stripe_price_id, uf
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
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at, owner_document, owner_document_type, additional_ies, stripe_price_id, uf
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

func (sm *subscriptionModel) GetPendingRegistrationByEmail(email string) (*entity_public.PendingRegistration, error) {
	rows, err := sm.pool.Query(context.Background(),
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at, owner_document, owner_document_type, additional_ies, stripe_price_id, uf
		 FROM pending_registration
		 WHERE email = @email`,
		pgx.NamedArgs{"email": email})

	if err != nil {
		return nil, err
	}

	pending, collectErr := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.PendingRegistration])
	if collectErr != nil {
		return nil, collectErr
	}

	return &pending, nil
}

func (sm *subscriptionModel) GetPendingRegistrationByCpf(cpf string) (*entity_public.PendingRegistration, error) {
	rows, err := sm.pool.Query(context.Background(),
		`SELECT id, email, name, passwd, cpf, inscricao_estadual, role, stripe_checkout_session_id, created_at, owner_document, owner_document_type, additional_ies, stripe_price_id, uf
		 FROM pending_registration
		 WHERE cpf = @cpf`,
		pgx.NamedArgs{"cpf": cpf})

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

func (sm *subscriptionModel) CreateFarmAndUserFromPending(pending entity_public.PendingRegistration) (uint32, []uint32, error) {
	ctx := context.Background()
	tx, err := sm.pool.Begin(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback(ctx)

	// Collect all IEs to create
	allIEs := []string{pending.InscricaoEstadual}
	if pending.AdditionalIEs != nil {
		allIEs = append(allIEs, pending.AdditionalIEs...)
	}

	farmIDs := make([]uint32, 0, len(allIEs))
	for _, ie := range allIEs {
		var fid uint32
		createFarmErr := tx.QueryRow(ctx,
			`INSERT INTO farm (inscricao_estadual)
			 VALUES (@inscricao_estadual)
			 RETURNING id`,
			pgx.NamedArgs{
				"inscricao_estadual": ie,
			}).Scan(&fid)
		if createFarmErr != nil {
			return 0, nil, createFarmErr
		}
		farmIDs = append(farmIDs, fid)
	}

	primaryFarmID := farmIDs[0]

	// Insert the user attached to the primary farm
	var userID uint32
	insertUserErr := tx.QueryRow(ctx,
		`INSERT INTO app_user (email, name, passwd, inscricao_estadual, farm, cpf, role)
		 VALUES (@email, @name, @passwd, @inscricao_estadual, @farm, @cpf, @role)
		 RETURNING id`,
		pgx.NamedArgs{
			"email":              pending.Email,
			"name":               pending.Name,
			"passwd":             pending.Passwd,
			"inscricao_estadual": pending.InscricaoEstadual,
			"farm":               primaryFarmID,
			"cpf":                pending.Cpf,
			"role":               pending.Role,
		}).Scan(&userID)
	if insertUserErr != nil {
		return 0, nil, insertUserErr
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return 0, nil, commitErr
	}

	return primaryFarmID, farmIDs, nil
}

func (sm *subscriptionModel) CompletePendingRegistrationFromLogin(pending entity_public.PendingRegistration, customerID, subscriptionID, status string, periodEnd time.Time, tierKey string) error {
	ctx := context.Background()
	tx, err := sm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ownerDoc := pending.Cpf
	ownerDocType := 1
	if pending.OwnerDocument != nil && *pending.OwnerDocument != "" {
		ownerDoc = *pending.OwnerDocument
		if pending.OwnerDocumentType != nil {
			ownerDocType = *pending.OwnerDocumentType
		}
	}

	var ownerId int
	ownerRow := tx.QueryRow(ctx,
		`INSERT INTO owner (owner_document, owner_document_type) VALUES (@owner_document, @owner_document_type) RETURNING id`,
		pgx.NamedArgs{
			"owner_document":      ownerDoc,
			"owner_document_type": ownerDocType,
		},
	)
	oErr := ownerRow.Scan(&ownerId)
	if oErr != nil {
		return oErr
	}

	// Create any missing farms
	allIEs := []string{pending.InscricaoEstadual}
	if pending.AdditionalIEs != nil {
		allIEs = append(allIEs, pending.AdditionalIEs...)
	}

	for _, ie := range allIEs {
		farmRow := tx.QueryRow(ctx,
			`INSERT INTO farm (inscricao_estadual, uf)
			 VALUES (@inscricao_estadual, @uf) RETURNING id
			 ON CONFLICT (inscricao_estadual) DO NOTHING`,
			pgx.NamedArgs{
				"inscricao_estadual": ie,
			},
		)
		var farmId int
		farmErr := farmRow.Scan(&farmId)
		if farmErr != nil {
			return farmErr
		}

		_, fosErr := tx.Exec(ctx,
			`INSERT INTO farm_owner_subscription (farm_id, owner_id, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, tier_key)
			VALUES (@farm_id, @owner_id, @stripe_customer_id, @stripe_subscription_id, @subscription_status, @period_end, @tier_key)`,
			pgx.NamedArgs{
				"farm_id":                farmId,
				"owner_id":               ownerId,
				"stripe_customer_id":     customerID,
				"stripe_subscription_id": subscriptionID,
				"subscription_status":    status,
				"period_end":             periodEnd,
				"tier_key":               tierKey,
			})
		if fosErr != nil {
			return fosErr
		}
	}

	// Delete pending registration
	_, delErr := tx.Exec(ctx,
		`DELETE FROM pending_registration WHERE id = @id`,
		pgx.NamedArgs{"id": pending.Id})
	if delErr != nil {
		return delErr
	}

	return tx.Commit(ctx)
}
