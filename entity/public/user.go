package entity_public

type User struct {
	Id                uint32 `form:"id"`
	Email             string `form:"email" binding:"required"`
	Name              string `form:"name" binding:"required"`
	Passwd            string `form:"passwd" binding:"required"`
	InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
	Farm              uint32 `form:"farm" binding:"gte=0"`
}

type NewUser struct {
	User
	PasswdConfirm string `form:"passwdConfirm" binding:"required"`
}

type SignInUser struct {
	Email  string `form:"email" binding:"required"`
	Passwd string `form:"passwd" binding:"required"`
}
