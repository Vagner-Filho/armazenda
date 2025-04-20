package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/user_model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

type ArmazendaUserClaims struct {
	Username string
	Email    string
	Farm     uint32
	jwt.RegisteredClaims
}

func createToken(username string, email string, farm uint32) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		ArmazendaUserClaims{
			Username: username,
			Email:    email,
			Farm:     farm,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 20)),
			},
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &ArmazendaUserClaims{}, func(token *jwt.Token) (any, error) {
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

type credentials struct {
	Token    string
	Username string
}

func Login(email string, passwd string) (credentials, *entity_public.Toast) {
	um := user_model.GetUserModel()
	user, err := um.AuthUser(email, passwd)

	if err != nil || user == nil {
		toast := entity_public.GetWarningToast("Credenciais inválidas", "")
		return credentials{}, &toast
	}

	token, tokenErr := createToken(user.Name, user.Email, user.Farm)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}
	return credentials{Token: token, Username: user.Name}, nil
}

func Create(newUser entity_public.NewUser) entity_public.Toast {
	if newUser.Passwd == newUser.PasswdConfirm {
		um := user_model.GetUserModel()
		created, err := um.CreateUser(newUser)

		if created == true {
			toast := entity_public.GetSuccessToast("Usuário criado", "")
			return toast
		}

		if err != nil {
			toast := entity_public.GetErrorToast(err.Error(), "")
			return toast
		}
	}
	toast := entity_public.GetWarningToast("Confirmação de senha incorreta", "")
	return toast
}
