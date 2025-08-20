package entity_public

type User struct {
	Id                uint32 `form:"id"`
	Email             string `form:"email" binding:"required"`
	Name              string `form:"name" binding:"required"`
	Passwd            string `form:"passwd" binding:"required"`
	InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
	Farm              uint32 `form:"farm" binding:"gte=0"`
	Cpf               string `form:"cpf" binding:"len=11"`
}

type NewUser struct {
	User
	PasswdConfirm string `form:"passwdConfirm" binding:"required"`
}

type SignInUser struct {
	Cpf    string `form:"cpf" binding:"len=11"`
	Passwd string `form:"passwd" binding:"required"`
}

type UserApproval struct {
	Id                uint32 `form:"id"`
	Email             string `form:"email" binding:"required"`
	Name              string `form:"name" binding:"required"`
	Passwd            string `form:"passwd" binding:"required"`
	InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
	FarmID            uint32 `form:"farm_id" binding:"gte=0"`
	Cpf               string `form:"cpf" binding:"len=11"`
	Status            string `form:"status" binding:"required"`
}

type PendingUser struct {
	Id    uint32 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Cpf   string `json:"cpf"`
}

