package user_service

import (
	"armazenda/model/user_model"

	"golang.org/x/crypto/bcrypt"
)

var users = []user_model.User{
	{Name: "Admin", Password: GeneratePasswordHash("teste"), Login: "admin", Email: "teste@hotmail.com", Role: "admin"},
	{Name: "Generic user", Password: GeneratePasswordHash("teste"), Login: "user", Email: "teste@hotmail.com", Role: "user"},
}

type UserService struct{}

func GeneratePasswordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func (us *UserService) ValidateCredentials(Login, password string) bool {
	for _, user := range users {
		if user.Login == Login {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				return true
			}
		}
	}
	return false
}
