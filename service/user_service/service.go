package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/user_model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

func createToken(username string, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"email":    email,
			"exp":      time.Now().Add(time.Hour * 20).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func Login(email string, psswd string) (string, *entity_public.Toast) {
	um := user_model.GetUserModel()
	user, err := um.AuthUser(email, psswd)

	if err != nil || user == nil {
		toast := entity_public.GetWarningToast("Credenciais inválidas", "")
		return "", &toast
	}

	token, tokenErr := createToken(user.Name, user.Email)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return "", &toast
	}
	return token, nil
}
