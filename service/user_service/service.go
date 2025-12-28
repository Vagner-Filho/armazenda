package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/farm_config_model"
	"armazenda/model/user_model"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getSecret() []byte {
	var tokenSecret = os.Getenv("TOKEN_SCRT")
	if len(tokenSecret) == 0 {
		fmt.Printf("\nno token secret has been provided\n")
		os.Exit(1)
	}
	return []byte(tokenSecret)
}

var secretKey = getSecret()

func GetFarmFromToken(sessionId string) uint32 {
	allocatedClaims := &ArmazendaUserClaims{}
	token, err := jwt.ParseWithClaims(sessionId, allocatedClaims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil || token == nil || !token.Valid {
		return 0
	}

	retrievedClaims := token.Claims.(*ArmazendaUserClaims)
	return retrievedClaims.Farm
}

func GetClaimsFromToken(sessionId string) *ArmazendaUserClaims {
	allocatedClaims := &ArmazendaUserClaims{}
	token, err := jwt.ParseWithClaims(sessionId, allocatedClaims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil || token == nil || !token.Valid {
		return nil
	}

	retrievedClaims := token.Claims.(*ArmazendaUserClaims)
	return retrievedClaims
}

type ArmazendaUserClaims struct {
	Username string
	Email    string
	Farm     uint32
	Role     string
	jwt.RegisteredClaims
	Id uint32
}

func createToken(username string, email string, farm uint32, role string, id uint32) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		ArmazendaUserClaims{
			Username: username,
			Email:    email,
			Farm:     farm,
			Role:     role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 20)),
			},
			Id: id,
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

func Login(cpf string, passwd string) (credentials, *entity_public.Toast) {
	um := user_model.GetUserModel()
	user, err := um.AuthUser(cpf, passwd)

	if err != nil || user == nil {
		toast := entity_public.GetWarningToast("Credenciais inválidas", "")
		return credentials{}, &toast
	}

	token, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}
	return credentials{Token: token, Username: user.Name}, nil
}

func Create(newUser entity_public.NewUser) entity_public.Toast {
	if newUser.Passwd != newUser.PasswdConfirm {
		return entity_public.GetWarningToast("Confirmação de senha incorreta", "")
	}

	fcm := farm_config_model.GetFarmConfigModel()
	farm, err := fcm.GetFarmByInscricaoEstadual(newUser.InscricaoEstadual)
	if err != nil {
		return entity_public.GetErrorToast(err.Error(), "")
	}

	um := user_model.GetUserModel()

	if farm != nil {
		created, err := um.CreateUserApproval(newUser, farm.Id)
		if !created || err != nil {
			return entity_public.GetErrorToast(err.Error(), "")
		}
		return entity_public.GetSuccessToast("Usuário enviado para aprovação", "")
	}

	created, err := um.CreateUser(newUser)
	if !created || err != nil {
		fmt.Printf("%v", err.Error())
		return entity_public.GetErrorToast(err.Error(), "")
	}

	return entity_public.GetSuccessToast("Usuário criado", "")
}

func GetRoleFromToken(sessionId string) string {
	allocatedClaims := &ArmazendaUserClaims{}
	token, err := jwt.ParseWithClaims(sessionId, allocatedClaims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil || token == nil || !token.Valid {
		return ""
	}

	retrievedClaims := token.Claims.(*ArmazendaUserClaims)
	return retrievedClaims.Role
}

func IsAdmin(sessionId string) bool {
	return GetRoleFromToken(sessionId) == "admin"
}

func MakeAdmin(userId uint32) error {
	um := user_model.GetUserModel()
	return um.MakeAdmin(userId)
}

func RemoveAdmin(userId uint32, adminId uint32) (*entity_public.User, entity_public.Toast) {
	um := user_model.GetUserModel()
	user, err := um.GetUserById(adminId)
	if user.Role != "admin" {
		return nil, entity_public.GetWarningToast("Somente o admin pode realizar esta ação", "consulte um admin")
	}

	if err != nil {
		return nil, entity_public.GetErrorToast("Falha ao encontrar usuário", "")
	}

	adminCount, err := um.GetAdminCount(adminId)
	if adminCount <= 1 {
		return nil, entity_public.GetWarningToast("A fazenda precisa de no mínimo 1 admin", "torne outro usuário admin")
	}

	if err != nil {
		return nil, entity_public.GetErrorToast("Falha ao consultar admins", "")
	}

	usr, err := um.RemoveAdmin(userId)
	if err != nil {
		return nil, entity_public.GetErrorToast("Falha ao remover privilégios de administrador", "")
	}

	return usr, entity_public.GetSuccessToast("Privilégios de administrador removidos", "")
}

func RemoveUser(userId uint32) error {
	um := user_model.GetUserModel()
	return um.RemoveUser(userId)
}

func GetUsersByFarm(farmId uint32, isAdmin bool) ([]entity_public.PendingUser, error) {
	um := user_model.GetUserModel()
	return um.GetUsersByFarm(farmId, isAdmin)
}

func ActivateUser(userId uint32) error {
	um := user_model.GetUserModel()
	return um.ActivateUser(userId)
}
