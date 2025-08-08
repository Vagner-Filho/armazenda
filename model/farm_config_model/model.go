package farm_config_model

import (
	"armazenda/entity/public"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type farmConfigModel struct {
	conn *pgx.Conn
}

var farmConfigModelImpl *farmConfigModel

func InitFarmConfigModel(conn *pgx.Conn) (*farmConfigModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if farmConfigModelImpl == nil {
		farmConfigModelImpl = &farmConfigModel{
			conn: conn,
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
		SELECT * FROM update_get_farm(@id, @inscricao_estadual, @name, @street, @cep, @number, @neighborhood, @city, @state, @complement, @email, @phone_number, @humidity_discount);
	`
	_, err := fcm.conn.Exec(context.Background(), query, pgx.NamedArgs{
		"id":                 config.Id,
		"inscricao_estadual": config.InscricaoEstadual,
		"name":               config.Name,
		"street":             config.Address.Street,
		"cep":                config.Address.Cep,
		"number":             config.Address.Number,
		"neighborhood":       config.Address.Neighborhood,
		"city":               config.Address.City,
		"state":              config.Address.State,
		"complement":         config.Address.Complement,
		"email":              config.Address.Email,
		"phone_number":       config.Address.PhoneNumber,
		"humidity_discount":  config.HumidityDiscount,
	})
	return err
}

func (fcm *farmConfigModel) GetFarmConfig(farmID uint32) (*entity_public.Farm, error) {
	query := `
		SELECT f.id, f.inscricao_estadual, fc.name, COALESCE(fc.humidity_discount, 1.15), fa.id, fa.street, fa.cep, fa.number, fac.complement, fa.neighborhood, fa.city, fa.state, fco.email, fco.phone_number
		FROM farm f
		LEFT JOIN farm_config fc ON f.id = fc.farm_id
		LEFT JOIN farm_address fa ON fa.farm_id = f.id
		LEFT JOIN farm_address_complement fac ON fac.farm_address_id = fa.id
		LEFT JOIN farm_contact fco ON fco.farm_id = f.id
		WHERE f.id = $1;
	`
	rows, _ := fcm.conn.Query(context.Background(), query, farmID)
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
