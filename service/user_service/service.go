package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/farm_config_model"
	"armazenda/model/user_model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func GetGoogleLoginURL() string {
	return googleOauthConfig.AuthCodeURL("state")
}

type GoogleLoginResult struct {
	Token     string
	Username  string
	IsNewUser bool
	Email     string
	Name      string
}

func LoginWithGoogle(code string) (GoogleLoginResult, *entity_public.Toast) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao trocar código com Google", "")
		return GoogleLoginResult{}, &toast
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao obter dados do usuário do Google", "")
		return GoogleLoginResult{}, &toast
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao ler resposta do Google", "")
		return GoogleLoginResult{}, &toast
	}

	var googleUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(contents, &googleUser); err != nil {
		toast := entity_public.GetErrorToast("Falha ao processar dados do Google", "")
		return GoogleLoginResult{}, &toast
	}

	um := user_model.GetUserModel()
	user, err := um.GetUserByEmail(googleUser.Email)

	if err != nil || user == nil {
		return GoogleLoginResult{
			IsNewUser: true,
			Email:     googleUser.Email,
			Name:      googleUser.Name,
		}, nil
	}

	jwtToken, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return GoogleLoginResult{}, &toast
	}

	return GoogleLoginResult{
		Token:    jwtToken,
		Username: user.Name,
	}, nil
}

type PreRegistrationClaims struct {
	Email string
	Name  string
	jwt.RegisteredClaims
}

func CreatePreRegistrationToken(email, name string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		PreRegistrationClaims{
			Email: email,
			Name:  name,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 30 mins to finish registration
			},
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyPreRegistrationToken(tokenString string) (*PreRegistrationClaims, error) {
	claims := &PreRegistrationClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

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
	existsAndIsActive, err := um.ExistsAndIsActive(newUser.Cpf)
	if existsAndIsActive == true {
		return entity_public.GetWarningToast("CPF em uso em outro armazém", "Um adm precisa desativa-lo para cadastra-lo aqui")
	}

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
