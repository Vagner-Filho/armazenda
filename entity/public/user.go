package entity_public

type User struct {
	Email             string `form:"email" binding:"required"`
	Passwd            string `form:"passwd" binding:"required"`
	PasswdConfirm     string `form:"passwdConfirm" binding:"required"`
	InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
}
