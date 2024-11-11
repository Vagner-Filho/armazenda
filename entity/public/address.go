package entity_public

type Address struct {
	Id           int8   `form:"id"`
	Street       string `form:"street" binding:"required"`
	Cep          string `form:"cep" binding:"required"`
	Number       int8   `form:"number"`
	Complement   string `form:"complement"`
	Neighborhood string `form:"neighborhood" binding:"required"`
	City         string `form:"city" binding:"required"`
	State        string `form:"state" binding:"required"`
	Email        string `form:"email"`
	PhoneNumber  string `form:"phoneNumber"`
}
