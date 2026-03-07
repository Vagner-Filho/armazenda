package entity_public

type Farm struct {
	Id                    uint32  `form:"id"`
	InscricaoEstadual     string  `form:"inscricaoEstadual" binding:"required"`
	Name                  *string `form:"name"`
	HumidityProgressionId *uint32 `form:"humidityProgressionId" json:"humidityProgressionId"`
	Address
	StorageName *string `form:"storageName"`
}
