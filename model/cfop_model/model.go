package cfop_model

import (
	"context"
	"errors"
	"sort"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CfopModel reads the system-wide CFOP catalog and manages farm-tied CFOPs
// plus their per-farm use counters. The catalog (cfop) is read-only through
// the app; only the migration seeds it.
type CfopModel struct {
	pool *pgxpool.Pool
}

var cfopModelImpl *CfopModel

func InitCfopModel(pool *pgxpool.Pool) (*CfopModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if cfopModelImpl == nil {
		cfopModelImpl = &CfopModel{
			pool: pool,
		}
	}

	return cfopModelImpl, nil
}

func GetCfopModel() *CfopModel {
	if cfopModelImpl == nil {
		panic("\ncfop model hasnt been initialized\n")
	}
	return cfopModelImpl
}

// ListCfopsForFarm merges the system catalog with the farm's registered CFOPs
// (disjoint by construction: registering a catalog code is rejected) and joins
// the farm's use counters. The result is ordered by code; callers split it
// into "Mais utilizados" / "Todos os CFOPs" with GroupCfopOptions.
func (m *CfopModel) ListCfopsForFarm(farmID uint32) ([]entity_public.CfopOption, *model_error.ModelError) {
	rows, err := m.pool.Query(context.Background(), `
		SELECT c.code, c.description, c.origin_destination,
		       COALESCE(u.use_count, 0) AS use_count, false AS is_farm
		FROM cfop c
		LEFT JOIN farm_cfop_use u ON u.farm_id = @farmId AND u.code = c.code
		UNION ALL
		SELECT f.code, f.description, f.origin_destination,
		       COALESCE(u.use_count, 0) AS use_count, true AS is_farm
		FROM farm_cfop f
		LEFT JOIN farm_cfop_use u ON u.farm_id = f.farm_id AND u.code = f.code
		WHERE f.farm_id = @farmId
		ORDER BY code
	`, pgx.NamedArgs{"farmId": farmID})
	if err != nil {
		model_error.GetLoggerModel().Log("ListCfopsForFarm query error: " + err.Error())
		return nil, &model_error.ModelError{Message: "Falhamos ao buscar os CFOPs", IsServerErr: true}
	}

	options, collectErr := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity_public.CfopOption, error) {
		var option entity_public.CfopOption
		scanErr := row.Scan(&option.Code, &option.Description, &option.OriginDestination, &option.UseCount, &option.IsFarm)
		return option, scanErr
	})
	if collectErr != nil {
		model_error.GetLoggerModel().Log("ListCfopsForFarm collect error: " + collectErr.Error())
		return nil, &model_error.ModelError{Message: "Falhamos ao buscar os CFOPs", IsServerErr: true}
	}

	return options, nil
}

// CatalogHasCfop reports whether the code already exists in the system-wide
// catalog. The registration flow rejects catalog duplicates because catalog
// and farm CFOPs must stay disjoint (no per-farm description overrides).
func (m *CfopModel) CatalogHasCfop(code string) (bool, *model_error.ModelError) {
	var exists bool
	err := m.pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM cfop WHERE code = @code)`,
		pgx.NamedArgs{"code": code},
	).Scan(&exists)
	if err != nil {
		model_error.GetLoggerModel().Log("CatalogHasCfop query error: " + err.Error())
		return false, &model_error.ModelError{Message: "Falhamos ao verificar o CFOP", IsServerErr: true}
	}
	return exists, nil
}

// AddFarmCfop registers a farm-tied CFOP. A duplicate inside the same farm is
// reported as a user-facing error (warning toast); everything else is a server
// error. Catalog duplicates are rejected by the caller before reaching here.
func (m *CfopModel) AddFarmCfop(farmID uint32, code, description, originDestination string) (entity_public.CfopOption, *model_error.ModelError) {
	var option entity_public.CfopOption
	err := m.pool.QueryRow(context.Background(), `
		INSERT INTO farm_cfop (farm_id, code, description, origin_destination)
		VALUES (@farmId, @code, @description, @originDestination)
		RETURNING code, description, origin_destination
	`, pgx.NamedArgs{
		"farmId":            farmID,
		"code":              code,
		"description":       description,
		"originDestination": originDestination,
	}).Scan(&option.Code, &option.Description, &option.OriginDestination)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity_public.CfopOption{}, &model_error.ModelError{Message: "CFOP já cadastrado"}
		}
		model_error.GetLoggerModel().Log("AddFarmCfop error: " + err.Error())
		return entity_public.CfopOption{}, &model_error.ModelError{Message: "Falhamos ao cadastrar o CFOP", IsServerErr: true}
	}

	option.IsFarm = true
	return option, nil
}

// IncrementFarmCfopUse bumps the farm's use counter once for each code in
// codes (callers pass distinct codes). This is a local ranking update only:
// it never contacts SEFAZ and its failure must never block an emission.
func (m *CfopModel) IncrementFarmCfopUse(farmID uint32, codes []string) *model_error.ModelError {
	if len(codes) == 0 {
		return nil
	}

	_, err := m.pool.Exec(context.Background(), `
		INSERT INTO farm_cfop_use (farm_id, code, use_count, updated_at)
		SELECT @farmId, c.code, 1, CURRENT_TIMESTAMP
		FROM unnest(@codes::text[]) AS c(code)
		ON CONFLICT (farm_id, code) DO UPDATE
		SET use_count = farm_cfop_use.use_count + 1,
		    updated_at = CURRENT_TIMESTAMP
	`, pgx.NamedArgs{"farmId": farmID, "codes": codes})
	if err != nil {
		model_error.GetLoggerModel().Log("IncrementFarmCfopUse error: " + err.Error())
		return &model_error.ModelError{Message: "Falhamos ao atualizar o uso do CFOP", IsServerErr: true}
	}

	return nil
}

// GroupCfopOptions splits merged options into the "Mais utilizados" group
// (UseCount > 0, most used first, ties by code ASC) and "Todos os CFOPs" (code
// ASC). The most-used slice is empty when no code has been used yet.
func GroupCfopOptions(options []entity_public.CfopOption) (mostUsed, all []entity_public.CfopOption) {
	all = make([]entity_public.CfopOption, 0, len(options))
	mostUsed = make([]entity_public.CfopOption, 0, len(options))
	for _, option := range options {
		all = append(all, option)
		if option.UseCount > 0 {
			mostUsed = append(mostUsed, option)
		}
	}

	sort.SliceStable(all, func(i, j int) bool { return all[i].Code < all[j].Code })
	sort.SliceStable(mostUsed, func(i, j int) bool {
		if mostUsed[i].UseCount != mostUsed[j].UseCount {
			return mostUsed[i].UseCount > mostUsed[j].UseCount
		}
		return mostUsed[i].Code < mostUsed[j].Code
	})

	return mostUsed, all
}
