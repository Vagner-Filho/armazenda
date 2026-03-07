package humidity_progression_model

import (
	"context"
	"errors"
	"fmt"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type HumidityProgressionModel struct {
	pool *pgxpool.Pool
}

var humidityProgressionModelImpl *HumidityProgressionModel

func InitHumidityProgressionModel(pool *pgxpool.Pool) (*HumidityProgressionModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}
	if humidityProgressionModelImpl == nil {
		humidityProgressionModelImpl = &HumidityProgressionModel{pool: pool}
	}
	return humidityProgressionModelImpl, nil
}

func GetHumidityProgressionModel() *HumidityProgressionModel {
	if humidityProgressionModelImpl == nil {
		panic("humidity progression model hasn't been initialized")
	}
	return humidityProgressionModelImpl
}

// GetProgression retrieves a progression by ID with all its tiers
func (hm *HumidityProgressionModel) GetProgression(id uint32) (entity_public.HumidityProgression, *model_error.ModelError) {
	var progression entity_public.HumidityProgression

	// Get progression details
	row, err := hm.pool.Query(context.Background(), `
		SELECT id, name, farm_id, is_system_default 
		FROM humidity_progression 
		WHERE id = $1
	`, id)
	if err != nil {
		return progression, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	progression, collectErr := pgx.CollectExactlyOneRow(row, pgx.RowToStructByPos[entity_public.HumidityProgression])
	if collectErr != nil {
		if errors.Is(collectErr, pgx.ErrNoRows) {
			return progression, &model_error.ModelError{Message: "Progressão de humidade não encontrada"}
		}
		return progression, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	// Get tiers
	tiersRow, err := hm.pool.Query(context.Background(), `
		SELECT id, progression_id, threshold_humidity, discount_value 
		FROM humidity_progression_tier 
		WHERE progression_id = $1 
		ORDER BY threshold_humidity ASC
	`, id)
	if err != nil {
		return progression, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	tiers, collectErr := pgx.CollectRows(tiersRow, pgx.RowToStructByPos[entity_public.HumidityProgressionTier])
	if collectErr != nil {
		return progression, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	progression.Tiers = tiers
	return progression, nil
}

// GetDiscountForHumidity finds the appropriate discount value for a given humidity
// Returns the discount_value from the tier with highest threshold <= humidity
func (hm *HumidityProgressionModel) GetDiscountForHumidity(progressionId *uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError) {
	// If progressionId is nil, get system default
	var pid uint32
	if progressionId == nil {
		sysDefault, err := hm.GetSystemDefaultProgression()
		if err != nil {
			return decimal.Zero, err
		}
		pid = sysDefault
	} else {
		pid = *progressionId
	}

	// Find the appropriate tier
	var discountValue decimal.Decimal
	err := hm.pool.QueryRow(context.Background(), `
		SELECT discount_value 
		FROM humidity_progression_tier 
		WHERE progression_id = $1 AND threshold_humidity <= $2
		ORDER BY threshold_humidity DESC 
		LIMIT 1
	`, pid, humidity).Scan(&discountValue)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No tier found for this humidity (humidity < lowest threshold)
			return decimal.Zero, nil
		}
		return decimal.Zero, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return discountValue, nil
}

// GetSystemDefaultProgression returns the system default progression ID
func (hm *HumidityProgressionModel) GetSystemDefaultProgression() (uint32, *model_error.ModelError) {
	var id uint32
	err := hm.pool.QueryRow(context.Background(), `
		SELECT id FROM humidity_progression WHERE is_system_default = TRUE
	`).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &model_error.ModelError{Message: "System default progression not found", IsServerErr: true}
		}
		return 0, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return id, nil
}

// AddProgression creates a new progression with tiers
func (hm *HumidityProgressionModel) AddProgression(name string, farmId uint32, tiers []entity_public.HumidityProgressionTier) (uint32, *model_error.ModelError) {
	// Validate tier count
	if len(tiers) < 1 {
		return 0, &model_error.ModelError{Message: "A progressão deve ter pelo menos 1 faixa"}
	}
	if len(tiers) > 30 {
		return 0, &model_error.ModelError{Message: "A progressão não pode ter mais de 30 faixas"}
	}

	ctx := context.Background()
	tx, err := hm.pool.Begin(ctx)
	if err != nil {
		return 0, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	defer tx.Rollback(ctx)

	// Insert progression
	var progressionId uint32
	err = tx.QueryRow(ctx, `
		INSERT INTO humidity_progression (name, farm_id, is_system_default) 
		VALUES ($1, $2, FALSE) 
		RETURNING id
	`, name, farmId).Scan(&progressionId)
	if err != nil {
		return 0, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	// Insert tiers
	for _, tier := range tiers {
		_, err = tx.Exec(ctx, `
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
			VALUES ($1, $2, $3)
		`, progressionId, tier.ThresholdHumidity, tier.DiscountValue)
		if err != nil {
			return 0, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return progressionId, nil
}

// UpdateProgression updates a progression and its tiers
func (hm *HumidityProgressionModel) UpdateProgression(id uint32, name string, tiers []entity_public.HumidityProgressionTier) *model_error.ModelError {
	// Validate tier count
	if len(tiers) < 1 {
		return &model_error.ModelError{Message: "A progressão deve ter pelo menos 1 faixa"}
	}
	if len(tiers) > 30 {
		return &model_error.ModelError{Message: "A progressão não pode ter mais de 30 faixas"}
	}

	// Check if this is the system default (cannot be modified)
	var isSystemDefault bool
	err := hm.pool.QueryRow(context.Background(), `
		SELECT is_system_default FROM humidity_progression WHERE id = $1
	`, id).Scan(&isSystemDefault)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model_error.ModelError{Message: "Progressão não encontrada"}
		}
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	if isSystemDefault {
		return &model_error.ModelError{Message: "Não é possível modificar a progressão padrão do sistema"}
	}

	ctx := context.Background()
	tx, err := hm.pool.Begin(ctx)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	defer tx.Rollback(ctx)

	// Update progression name
	_, err = tx.Exec(ctx, `
		UPDATE humidity_progression SET name = $1 WHERE id = $2
	`, name, id)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	// Delete old tiers
	_, err = tx.Exec(ctx, `
		DELETE FROM humidity_progression_tier WHERE progression_id = $1
	`, id)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	// Insert new tiers
	for _, tier := range tiers {
		_, err = tx.Exec(ctx, `
			INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
			VALUES ($1, $2, $3)
		`, id, tier.ThresholdHumidity, tier.DiscountValue)
		if err != nil {
			return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return nil
}

// DeleteProgression deletes a progression (cannot delete system default)
func (hm *HumidityProgressionModel) DeleteProgression(id uint32) *model_error.ModelError {
	// Check if this is the system default
	var isSystemDefault bool
	err := hm.pool.QueryRow(context.Background(), `
		SELECT is_system_default FROM humidity_progression WHERE id = $1
	`, id).Scan(&isSystemDefault)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model_error.ModelError{Message: "Progressão não encontrada"}
		}
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	if isSystemDefault {
		return &model_error.ModelError{Message: "Não é possível excluir a progressão padrão do sistema"}
	}

	// Check if progression is in use
	var count int
	err = hm.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM person_config WHERE humidity_progression_id = $1
	`, id).Scan(&count)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	if count > 0 {
		return &model_error.ModelError{Message: "Não é possível excluir: progressão está em uso por pessoas"}
	}

	err = hm.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM farm_config WHERE humidity_progression_id = $1
	`, id).Scan(&count)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}
	if count > 0 {
		return &model_error.ModelError{Message: "Não é possível excluir: progressão está em uso por fazendas"}
	}

	_, err = hm.pool.Exec(context.Background(), `
		DELETE FROM humidity_progression WHERE id = $1
	`, id)
	if err != nil {
		return &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	return nil
}

// ListProgressions lists all progressions for a farm (including system default)
func (hm *HumidityProgressionModel) ListProgressions(farmId uint32) ([]entity_public.HumidityProgression, *model_error.ModelError) {
	rows, err := hm.pool.Query(context.Background(), `
		SELECT id, name, farm_id, is_system_default 
		FROM humidity_progression 
		WHERE farm_id = $1 OR is_system_default = TRUE
		ORDER BY is_system_default DESC, name ASC
	`, farmId)
	if err != nil {
		return nil, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	progressions, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.HumidityProgression])
	if err != nil {
		return nil, &model_error.ModelError{Message: err.Error(), IsServerErr: true}
	}

	// Load tiers for each progression
	for i := range progressions {
		tiersRow, err := hm.pool.Query(context.Background(), `
			SELECT id, progression_id, threshold_humidity, discount_value 
			FROM humidity_progression_tier 
			WHERE progression_id = $1 
			ORDER BY threshold_humidity ASC
		`, progressions[i].Id)
		if err != nil {
			fmt.Printf("error loading tiers for progression %d: %v\n", progressions[i].Id, err)
			continue
		}

		tiers, err := pgx.CollectRows(tiersRow, pgx.RowToStructByPos[entity_public.HumidityProgressionTier])
		if err != nil {
			fmt.Printf("error collecting tiers for progression %d: %v\n", progressions[i].Id, err)
			continue
		}

		progressions[i].Tiers = tiers
	}

	return progressions, nil
}
