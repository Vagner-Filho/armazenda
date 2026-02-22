package entity_public

type Address struct {
	Id           *uint16 `form:"id" json:"id,omitempty"`
	Street       *string `form:"street" json:"street,omitempty"`
	Cep          *string `form:"cep" json:"cep,omitempty"`
	Number       *uint32 `form:"number" json:"number,omitempty"`
	Complement   *string `form:"complement" json:"complement,omitempty"`
	Neighborhood *string `form:"neighborhood" json:"neighborhood,omitempty"`
	City         *string `form:"city" json:"city,omitempty"`
	State        *string `form:"state" json:"state,omitempty"`
	Email        *string `form:"email" json:"email,omitempty"`
	PhoneNumber  *string `form:"phoneNumber" json:"phoneNumber,omitempty"`
}
