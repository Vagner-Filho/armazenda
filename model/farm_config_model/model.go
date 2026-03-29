package farm_config_model

import (
	"armazenda/entity/public"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type farmConfigModel struct {
	pool *pgxpool.Pool
}

var farmConfigModelImpl *farmConfigModel

func InitFarmConfigModel(pool *pgxpool.Pool) (*farmConfigModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if farmConfigModelImpl == nil {
		farmConfigModelImpl = &farmConfigModel{
			pool: pool,
		}
	}

	return farmConfigModelImpl, nil
}

func GetFarmConfigModel() *farmConfigModel {
	if farmConfigModelImpl == nil {
		panic("farm config model hasnt been initialized")
	}
	return farmConfigModelImpl
}

func (fcm *farmConfigModel) UpsertFarmConfig(config *entity_public.Farm) error {
	query := `
		SELECT * FROM update_get_farm(@id, @inscricao_estadual, @name, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone_number, @storage_name, @humidity_progression_id);
	`
	_, err := fcm.pool.Exec(context.Background(), query, pgx.NamedArgs{
		"id":                      config.Id,
		"inscricao_estadual":      config.InscricaoEstadual,
		"name":                    config.Name,
		"street":                  config.Address.Street,
		"cep":                     config.Address.Cep,
		"number":                  config.Address.Number,
		"neighborhood":            config.Address.Neighborhood,
		"city":                    config.Address.City,
		"state":                   config.Address.State,
		"complement":              config.Address.Complement,
		"email":                   config.Address.Email,
		"phone_number":            config.Address.PhoneNumber,
		"storage_name":            config.StorageName,
		"humidity_progression_id": config.HumidityProgressionId,
	})
	return err
}

func (fcm *farmConfigModel) GetFarmConfig(farmID uint32) (*entity_public.Farm, error) {
	query := `
		SELECT f.id, f.inscricao_estadual, fc.name, fc.humidity_progression_id, fa.id, fa.street, fa.cep, fa.number, fac.complement, fa.neighborhood, fa.city, fa.state, fco.email, fco.phone_number, fc.storage_name
		FROM farm f
		LEFT JOIN farm_config fc ON f.id = fc.farm_id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN farm_address_complement fac ON fac.farm_address_id = fa.id
		LEFT JOIN farm_contact fco ON fco.farm_id = f.id
		WHERE f.id = $1;
	`
	rows, _ := fcm.pool.Query(context.Background(), query, farmID)
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[entity_public.Farm])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		fmt.Printf("%+v\n", err.Error())
		return nil, fmt.Errorf("failed to get farm config: %w", err)
	}
	result.Id = farmID
	return &result, nil
}

func (fcm *farmConfigModel) GetFarmByInscricaoEstadual(inscricaoEstadual string) (*entity_public.Farm, error) {
	query := `
		SELECT f.id, f.inscricao_estadual, fc.name, fc.humidity_progression_id, fa.id, fa.street, fa.cep, fa.number, fac.complement, fa.neighborhood, fa.city, fa.state, fco.email, fco.phone_number, fc.storage_name
		FROM farm f
		LEFT JOIN farm_config fc ON f.id = fc.farm_id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN farm_address_complement fac ON fac.farm_address_id = fa.id
		LEFT JOIN farm_contact fco ON fco.farm_id = f.id
		WHERE f.inscricao_estadual = $1;
	`
	rows, _ := fcm.pool.Query(context.Background(), query, inscricaoEstadual)
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[entity_public.Farm])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		fmt.Printf("%+v\n", err.Error())
		return nil, fmt.Errorf("failed to get farm config: %w", err)
	}
	return &result, nil
}

// SetFarmHumidityProgression updates the farm's default humidity progression
func (fcm *farmConfigModel) SetFarmHumidityProgression(farmID uint32, progressionID *uint32) error {
	ctx := context.Background()

	// Validate progression exists and is accessible (farm-specific or system default)
	if progressionID != nil {
		var exists bool
		err := fcm.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM humidity_progression 
				WHERE id = $1 AND (farm_id = $2 OR is_system_default = TRUE) AND is_active = TRUE
			)
		`, *progressionID, farmID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to validate progression: %w", err)
		}
		if !exists {
			return fmt.Errorf("progressão não encontrada ou não acessível")
		}
	}

	// Check if farm_config exists
	var configExists bool
	err := fcm.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM farm_config WHERE farm_id = $1)
	`, farmID).Scan(&configExists)
	if err != nil {
		return fmt.Errorf("failed to check farm config: %w", err)
	}

	if configExists {
		// Update existing farm_config
		_, err = fcm.pool.Exec(ctx, `
			UPDATE farm_config
			SET humidity_progression_id = $1
			WHERE farm_id = $2
		`, progressionID, farmID)
		if err != nil {
			return fmt.Errorf("failed to update farm config: %w", err)
		}
	} else {
		// Need to create farm_config with default values
		// Get farm's inscricao_estadual for name, use default storage_name
		var inscricaoEstadual string
		err = fcm.pool.QueryRow(ctx, `
			SELECT inscricao_estadual FROM farm WHERE id = $1
		`, farmID).Scan(&inscricaoEstadual)
		if err != nil {
			return fmt.Errorf("failed to get farm data: %w", err)
		}

		_, err = fcm.pool.Exec(ctx, `
			INSERT INTO farm_config (farm_id, name, storage_name, humidity_progression_id)
			VALUES ($1, $2, $3, $4)
		`, farmID, inscricaoEstadual, "Armazém", progressionID)
		if err != nil {
			return fmt.Errorf("failed to create farm config: %w", err)
		}
	}

	return nil
}
