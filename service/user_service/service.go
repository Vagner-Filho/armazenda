package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/farm_config_model"
	"armazenda/model/user_model"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var (
	googleOauthConfig    *oauth2.Config
	microsoftOauthConfig *oauth2.Config
)

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

	microsoftOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("MICROSOFT_REDIRECT_URL"),
		ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
		Scopes: []string{
			"user.read",
			"openid",
			"profile",
			"email",
		},
		Endpoint: microsoft.AzureADEndpoint("common"),
	}
}

func GetGoogleLoginURL() string {
	return googleOauthConfig.AuthCodeURL("state")
}

func GetMicrosoftLoginURL() string {
	return microsoftOauthConfig.AuthCodeURL("state")
}

type OAuthLoginResult struct {
	Token     string
	Username  string
	IsNewUser bool
	Email     string
	Name      string
	Farm      uint32
}

func LoginWithGoogle(code, ipAddress, userAgent string) (OAuthLoginResult, *entity_public.Toast) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao trocar código com Google", "")
		return OAuthLoginResult{}, &toast
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao obter dados do usuário do Google", "")
		return OAuthLoginResult{}, &toast
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao ler resposta do Google", "")
		return OAuthLoginResult{}, &toast
	}

	var googleUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(contents, &googleUser); err != nil {
		toast := entity_public.GetErrorToast("Falha ao processar dados do Google", "")
		return OAuthLoginResult{}, &toast
	}

	um := user_model.GetUserModel()
	user, err := um.GetUserByEmail(googleUser.Email)

	if err != nil || user == nil {
		return OAuthLoginResult{
			IsNewUser: true,
			Email:     googleUser.Email,
			Name:      googleUser.Name,
		}, nil
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	jwtToken, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	return OAuthLoginResult{
		Token:    jwtToken,
		Username: user.Name,
		Farm:     user.Farm,
	}, nil
}

func LoginWithMicrosoft(code, ipAddress, userAgent string) (OAuthLoginResult, *entity_public.Toast) {
	token, err := microsoftOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("%v", err.Error())
		toast := entity_public.GetErrorToast("Falha ao trocar código com Microsoft", "")
		return OAuthLoginResult{}, &toast
	}

	client := microsoftOauthConfig.Client(context.Background(), token)
	response, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao obter dados do usuário da Microsoft", "")
		return OAuthLoginResult{}, &toast
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		toast := entity_public.GetErrorToast("Falha ao ler resposta da Microsoft", "")
		return OAuthLoginResult{}, &toast
	}

	var microsoftUser struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}

	if err := json.Unmarshal(contents, &microsoftUser); err != nil {
		toast := entity_public.GetErrorToast("Falha ao processar dados da Microsoft", "")
		return OAuthLoginResult{}, &toast
	}

	email := microsoftUser.Mail
	if email == "" {
		email = microsoftUser.UserPrincipalName
	}

	um := user_model.GetUserModel()
	user, err := um.GetUserByEmail(email)

	if err != nil || user == nil {
		return OAuthLoginResult{
			IsNewUser: true,
			Email:     email,
			Name:      microsoftUser.DisplayName,
		}, nil
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	jwtToken, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	return OAuthLoginResult{
		Token:    jwtToken,
		Username: user.Name,
		Farm:     user.Farm,
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
	Username  string
	Email     string
	Farm      uint32
	Role      string
	Id        uint32
	SessionId string
	jwt.RegisteredClaims
}

func createToken(username string, email string, farm uint32, role string, id uint32, sessionId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		ArmazendaUserClaims{
			Username:  username,
			Email:     email,
			Farm:      farm,
			Role:      role,
			Id:        id,
			SessionId: sessionId,
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

func ValidateTokenAndSession(tokenString string) (bool, error) {
	// First verify token signature and expiration
	token, err := jwt.ParseWithClaims(tokenString, &ArmazendaUserClaims{}, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, fmt.Errorf("invalid token")
	}

	// Extract session_id from claims
	claims, ok := token.Claims.(*ArmazendaUserClaims)
	if !ok {
		return false, fmt.Errorf("invalid token claims")
	}

	// Validate session in database
	return ValidateSession(claims.SessionId)
}

// Session management functions

type Session struct {
	ID        uint32
	SessionID string
	UserID    uint32
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	IsActive  bool
}

func CreateSession(userID uint32, ipAddress, userAgent string) (string, error) {
	um := user_model.GetUserModel()
	pool := um.GetPool()

	sessionID := generateSessionID()
	expiresAt := time.Now().Add(time.Hour * 20) // Same as JWT expiration

	_, err := pool.Exec(context.Background(),
		`INSERT INTO user_session (session_id, user_id, expires_at, ip_address, user_agent) 
		 VALUES ($1, $2, $3, $4, $5)`,
		sessionID, userID, expiresAt, ipAddress, userAgent)

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func GetSession(sessionID string) (*Session, error) {
	um := user_model.GetUserModel()
	pool := um.GetPool()

	var session Session
	err := pool.QueryRow(context.Background(),
		`SELECT id, session_id, user_id, created_at, expires_at, ip_address, user_agent, is_active 
		 FROM user_session WHERE session_id = $1`,
		sessionID).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.CreatedAt,
		&session.ExpiresAt, &session.IPAddress, &session.UserAgent, &session.IsActive)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func DeleteSession(sessionID string) error {
	um := user_model.GetUserModel()
	pool := um.GetPool()

	_, err := pool.Exec(context.Background(),
		`DELETE FROM user_session WHERE session_id = $1`,
		sessionID)

	return err
}

func DeleteUserSessions(userID uint32) error {
	um := user_model.GetUserModel()
	pool := um.GetPool()

	_, err := pool.Exec(context.Background(),
		`DELETE FROM user_session WHERE user_id = $1`,
		userID)

	return err
}

func ValidateSession(sessionID string) (bool, error) {
	session, err := GetSession(sessionID)
	if err != nil {
		return false, nil
	}

	if !session.IsActive {
		return false, nil
	}

	if time.Now().After(session.ExpiresAt) {
		return false, nil
	}

	// Check if user is still active (not deactivated)
	um := user_model.GetUserModel()
	pool := um.GetPool()

	var inactiveCount int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM inactive_user WHERE user_id = $1`,
		session.UserID).Scan(&inactiveCount)

	if err != nil {
		return false, err
	}

	if inactiveCount > 0 {
		// User is deactivated, delete their session
		DeleteSession(sessionID)
		return false, nil
	}

	return true, nil
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CleanupExpiredSessions() (int64, error) {
	um := user_model.GetUserModel()
	pool := um.GetPool()

	result, err := pool.Exec(context.Background(),
		`DELETE FROM user_session WHERE expires_at < NOW()`)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

type credentials struct {
	Token    string
	Username string
	Farm     uint32
}

func Login(cpf string, passwd string, ipAddress, userAgent string) (credentials, *entity_public.Toast) {
	um := user_model.GetUserModel()
	user, err := um.AuthUser(cpf, passwd)

	if err != nil || user == nil {
		toast := entity_public.GetWarningToast("Credenciais inválidas", "")
		return credentials{}, &toast
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}

	token, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID)
	if tokenErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}
	return credentials{Token: token, Username: user.Name, Farm: user.Farm}, nil
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
	// Delete all user sessions before changing role
	err := DeleteUserSessions(userId)
	if err != nil {
		return err
	}
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

	// Delete all user sessions before changing role
	err = DeleteUserSessions(userId)
	if err != nil {
		return nil, entity_public.GetErrorToast("Falha ao remover sessões do usuário", "")
	}

	usr, err := um.RemoveAdmin(userId)
	if err != nil {
		return nil, entity_public.GetErrorToast("Falha ao remover privilégios de administrador", "")
	}

	return usr, entity_public.GetSuccessToast("Privilégios de administrador removidos", "")
}

func RemoveUser(userId uint32) error {
	um := user_model.GetUserModel()
	// Delete all user sessions before deactivating
	err := DeleteUserSessions(userId)
	if err != nil {
		return err
	}
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
