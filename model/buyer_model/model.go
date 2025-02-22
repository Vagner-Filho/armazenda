package buyer_model

import (
	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type buyerModel struct {
	conn *pgx.Conn
}

var buyerModelImpl *buyerModel

func InitBuyerModel(conn *pgx.Conn) (*buyerModel, error) {
	if conn == nil {
		return nil, errors.New("conn cant be null")
	}

	if buyerModelImpl == nil {
		buyerModelImpl = &buyerModel{
			conn: conn,
		}
	}

	return buyerModelImpl, nil
}

func GetBuyerModel() *buyerModel {
	if buyerModelImpl == nil {
		panic("\nbuyer model hasnt been initialized\n")
	}
	return buyerModelImpl
}

func (bm *buyerModel) AddBuyerCompany(bc entity_public.BuyerCompany) (entity_public.BuyerDisplay, *model_error.ModelError) {
	row, queryErr := bm.conn.Query(context.Background(), `
			SELECT * FROM add_get_buyer_company(@ie, @cnpj, @fantasyName, @companyName)
		`, pgx.NamedArgs{"ie": bc.InscricaoEstadual, "cnpj": bc.Cnpj, "fantasyName": bc.FantasyName, "companyName": bc.CompanyName})
	if queryErr != nil {
		model_error.Logger(bm.conn, queryErr.Error())
		return entity_public.BuyerDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	buyer, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.BuyerDisplay])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
		return entity_public.BuyerDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return buyer, nil
}

func (bm *buyerModel) AddBuyerPerson(bp entity_public.BuyerPerson) (entity_public.BuyerDisplay, *model_error.ModelError) {
	row, queryErr := bm.conn.Query(context.Background(), `
			SELECT * FROM add_get_buyer_person(@ie, @cpf, @name)
		`, pgx.NamedArgs{"ie": bp.InscricaoEstadual, "cpf": bp.Cpf, "name": bp.Name})
	if queryErr != nil {
		model_error.Logger(bm.conn, queryErr.Error())
		return entity_public.BuyerDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	buyer, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByPos[entity_public.BuyerDisplay])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
		return entity_public.BuyerDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return buyer, nil
}

func (bm *buyerModel) GetBuyers() ([]entity_public.BuyerDisplay, *model_error.ModelError) {
	rows, queryErr := bm.conn.Query(context.Background(), `
		SELECT b.id, bc.companyname AS name FROM buyer b
		JOIN buyercompany bc ON b.id = bc.buyerid
		UNION
		SELECT b.id, bp.name FROM buyer b
		JOIN buyerperson bp ON b.id = bp.buyerid;
	`)
	if queryErr != nil {
		model_error.Logger(bm.conn, queryErr.Error())
		return []entity_public.BuyerDisplay{}, &model_error.ModelError{Message: queryErr.Error()}
	}

	buyers, collectErr := pgx.CollectRows(rows, pgx.RowToStructByPos[entity_public.BuyerDisplay])
	if collectErr != nil {
		model_error.Logger(bm.conn, collectErr.Error())
		return []entity_public.BuyerDisplay{}, &model_error.ModelError{Message: collectErr.Error(), IsServerErr: true}
	}

	return buyers, nil
}
