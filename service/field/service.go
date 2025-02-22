package field_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/field_model"
)

func GetFields() ([]entity_public.Field, *entity_public.Toast) {
	fModel, _ := field_model.GetFieldModel()
	fields, err := fModel.GetFields()

	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		return []entity_public.Field{}, &toast
	}

	return fields, nil
}
