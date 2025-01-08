package crop_model

import (
	entity_public "armazenda/entity/public"
	"armazenda/utils"
	"time"
)

var crops = []entity_public.Crop{
	{
		Id:        0,
		Name:      "Safra 2024",
		StartDate: time.Now().AddDate(0, -9, -3).Format(utils.TimeLayout),
	},
	{
		Id:        1,
		Name:      "Safra 2023",
		StartDate: time.Now().AddDate(-1, -5, -7).Format(utils.TimeLayout),
	},
	{
		Id:        2,
		Name:      "Safra 2022",
		StartDate: time.Now().AddDate(-2, 0, -7).Format(utils.TimeLayout),
	},
	{
		Id:        3,
		Name:      "Entressafra 2024",
		StartDate: time.Now().Format(utils.TimeLayout),
	},
}

func GetCrops() []entity_public.Crop {
	return crops
}

func AddCrop(c entity_public.Crop) entity_public.Crop {
	newCrop := entity_public.Crop{
		Name:      c.Name,
		StartDate: c.StartDate,
		Id:        crops[len(crops)-1].Id + 1,
	}
	crops = append(crops, newCrop)
	return newCrop
}
