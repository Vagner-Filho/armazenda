package entity_public

type Address struct {
	Id           *uint16 `form:"id"`
	Street       *string `form:"street"`
	Cep          *string `form:"cep"`
	Number       *uint32 `form:"number"`
	Complement   *string `form:"complement"`
	Neighborhood *string `form:"neighborhood"`
	City         *string `form:"city"`
	State        *string `form:"state"`
	Email        *string `form:"email"`
	PhoneNumber  *string `form:"phoneNumber"`
}
