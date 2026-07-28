package owner_subscription_model

import (
	"armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/product"
)

type ownerSubscriptionModel struct {
	pool *pgxpool.Pool
}

var ownerSubscriptionModelImpl *ownerSubscriptionModel

func InitOwnerSubscriptionModel(pool *pgxpool.Pool) (*ownerSubscriptionModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if ownerSubscriptionModelImpl == nil {
		ownerSubscriptionModelImpl = &ownerSubscriptionModel{
			pool: pool,
		}
	}

	return ownerSubscriptionModelImpl, nil
}

func GetOwnerSubscriptionModel() *ownerSubscriptionModel {
	if ownerSubscriptionModelImpl == nil {
		panic("\nowner subscription model hasnt been initialized\n")
	}
	return ownerSubscriptionModelImpl
}

func (osm *ownerSubscriptionModel) Create(farmId uint32, ownerDocument string, ownerDocumentType int, customerID, subscriptionID, status string, periodEnd time.Time, tierKey string) (uint32, error) {
	var ownerId uint32
	oRow := osm.pool.QueryRow(context.Background(),
		`INSERT INTO owner (owner_document, owner_document_type) VALUES (@owner_doc, @owner_doc_type) RETURNING id`,
		pgx.NamedArgs{
			"owner_doc":      ownerDocument,
			"owner_doc_type": ownerDocumentType,
		})
	oErr := oRow.Scan(&ownerId)

	var pgErr *pgconn.PgError
	if oErr != nil {
		if errors.As(oErr, &pgErr) && pgErr.Code != pgerrcode.UniqueViolation {
			return ownerId, oErr
		} else {
			// select from owner to get their id
			oRow := osm.pool.QueryRow(context.Background(),
				`SELECT id FROM owner WHERE owner_document = @owner_doc`,
				pgx.NamedArgs{
					"owner_doc": ownerDocument,
				})
			selectOwnerErr := oRow.Scan(&ownerId)
			if selectOwnerErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("owner_subscription_model err while selecting id from owner: %s", selectOwnerErr.Error()))
				return ownerId, selectOwnerErr
			}
		}
	}

	var id uint32
	err := osm.pool.QueryRow(context.Background(),
		`INSERT INTO farm_owner_subscription (farm_id, owner_id, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, tier_key)
		 VALUES (@farm_id, @owner_id, @stripe_customer_id, @stripe_subscription_id, @subscription_status, @period_end, @tier_key)
		 RETURNING id`,
		pgx.NamedArgs{
			"farm_id":                farmId,
			"owner_id":               ownerId,
			"stripe_customer_id":     customerID,
			"stripe_subscription_id": subscriptionID,
			"subscription_status":    status,
			"period_end":             periodEnd,
			"tier_key":               tierKey,
		}).Scan(&id)
	return id, err
}

func (osm *ownerSubscriptionModel) GetBySubscriptionID(subscriptionID string) (*entity_public.OwnerSubscription, error) {
	rows, err := osm.pool.Query(context.Background(),
		`SELECT id, owner_document, owner_document_type, stripe_customer_id, stripe_subscription_id, subscription_status, subscription_current_period_end, created_at, tier_key
		 FROM farm_owner_subscription WHERE stripe_subscription_id = @subscription_id`,
		pgx.NamedArgs{"subscription_id": subscriptionID})
	if err != nil {
		return nil, err
	}

	sub, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.OwnerSubscription])
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (osm *ownerSubscriptionModel) GetStatusAndTierByFarm(farmID uint32) (string, string, error) {
	var status string
	var tierKey sql.NullString
	err := osm.pool.QueryRow(context.Background(),
		`SELECT fos.subscription_status, fos.tier_key
		 FROM farm f
		 LEFT JOIN farm_owner_subscription fos ON f.id = fos.farm_id
		 WHERE f.id = @farm_id`,
		pgx.NamedArgs{"farm_id": farmID}).Scan(&status, &tierKey)
	if err != nil {
		return "", "", err
	}
	if !tierKey.Valid {
		return status, "", nil
	}
	return status, tierKey.String, nil
}

func (osm *ownerSubscriptionModel) CountFarmsByOwner(ownerDocument string, ownerDocumentType int) (int, error) {
	var count int
	err := osm.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM owner o
		 JOIN farm_owner_subscription fos ON o.id = fos.owner_id
		 WHERE o.owner_document = @owner_document AND o.owner_document_type = @owner_document_type`,
		pgx.NamedArgs{
			"owner_document":      ownerDocument,
			"owner_document_type": ownerDocumentType,
		}).Scan(&count)
	return count, err
}

func (osm *ownerSubscriptionModel) GetSubscriptionByFarm(farmID uint32) (*entity_public.OwnerSubscription, error) {
	rows, err := osm.pool.Query(context.Background(),
		`SELECT fos.id, o.owner_document, o.owner_document_type, fos.stripe_customer_id, fos.stripe_subscription_id, fos.subscription_status, fos.subscription_current_period_end, o.created_at, fos.tier_key
		 FROM farm_owner_subscription fos
		 JOIN owner o ON o.id = fos.owner_id
		 WHERE fos.farm_id = @farm_id`,
		pgx.NamedArgs{"farm_id": farmID})
	if err != nil {
		return nil, err
	}

	sub, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[entity_public.OwnerSubscription])
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (osm *ownerSubscriptionModel) GetTierKeyByStripeProductID(stripeProductID string) (string, error) {
	params := &stripe.ProductParams{}
	prod, pErr := product.Get(stripeProductID, params)
	if pErr != nil {
		model_error.GetLoggerModel().Log(pErr.Error())
		return "", pErr
	}
	tier := prod.Metadata["tier"]
	return tier, nil
}

func (osm *ownerSubscriptionModel) UpdateStatus(subscriptionID, status string, periodEnd time.Time) error {
	_, err := osm.pool.Exec(context.Background(),
		`UPDATE farm_owner_subscription
		 SET subscription_status = @status, subscription_current_period_end = @period_end
		 WHERE stripe_subscription_id = @subscription_id`,
		pgx.NamedArgs{
			"status":          status,
			"period_end":      periodEnd,
			"subscription_id": subscriptionID,
		})
	return err
}

func (osm *ownerSubscriptionModel) UpdateFromCheckout(id uint32, customerID, subscriptionID, status string, periodEnd time.Time, tierKey string) error {
	_, err := osm.pool.Exec(context.Background(),
		`UPDATE farm_owner_subscription
		 SET stripe_customer_id = @customer_id,
		     stripe_subscription_id = @subscription_id,
		     subscription_status = @status,
		     subscription_current_period_end = @period_end,
		     tier_key = @tier_key
		 WHERE id = @id`,
		pgx.NamedArgs{
			"id":              id,
			"customer_id":     customerID,
			"subscription_id": subscriptionID,
			"status":          status,
			"period_end":      periodEnd,
			"tier_key":        tierKey,
		})
	return err
}
