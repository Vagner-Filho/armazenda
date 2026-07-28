package user_service

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/farm_config_model"
	"armazenda/model/owner_subscription_model"
	"armazenda/model/subscription_model"
	"armazenda/model/user_model"
	"armazenda/service/billing_service"
	field_service "armazenda/service/field"
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
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v85"
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

	// Check subscription status before creating session
	status, tierKey := resolveSubscriptionForLogin(user.Farm, user.Email, user.Cpf)
	if status != "active" && status != "trialing" && status != "past_due" {
		toast := entity_public.GetWarningToast("Assinatura inativa", "subscription-inactive")
		return OAuthLoginResult{Farm: user.Farm}, &toast
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	jwtToken, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID, tierKey)
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

	// Check subscription status before creating session
	status, tierKey := resolveSubscriptionForLogin(user.Farm, email, user.Cpf)
	if status != "active" && status != "trialing" && status != "past_due" {
		toast := entity_public.GetWarningToast("Assinatura inativa", "subscription-inactive")
		return OAuthLoginResult{Farm: user.Farm}, &toast
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast := entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return OAuthLoginResult{}, &toast
	}

	jwtToken, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID, tierKey)
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
	TierKey   string
	jwt.RegisteredClaims
}

func createToken(username string, email string, farm uint32, role string, id uint32, sessionId string, tierKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		ArmazendaUserClaims{
			Username:  username,
			Email:     email,
			Farm:      farm,
			Role:      role,
			Id:        id,
			SessionId: sessionId,
			TierKey:   tierKey,
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

func GetTierKeyFromContext(c interface {
	Get(string) (interface{}, bool)
}) string {
	tierKey, exists := c.Get("tier_key")
	if !exists {
		return ""
	}
	if s, ok := tierKey.(string); ok {
		return s
	}
	return ""
}

func resolveSubscriptionForLogin(farmID uint32, email string, cpf string) (string, string) {
	osm := owner_subscription_model.GetOwnerSubscriptionModel()

	// 1. Check local subscription
	status, tierKey, subErr := osm.GetStatusAndTierByFarm(farmID)
	if subErr != nil {
		fmt.Printf("failed to fetch subscription status for farm %d: %v\n", farmID, subErr)
	}
	if status == "active" || status == "trialing" || status == "past_due" {
		return status, tierKey
	}

	// 2. Fallback A: Check Stripe via stored subscription ID
	subRecord, subRecordErr := osm.GetSubscriptionByFarm(farmID)
	if subRecordErr == nil && subRecord != nil && subRecord.StripeSubscriptionId != nil && *subRecord.StripeSubscriptionId != "" {
		stripeSub, stripeErr := billing_service.GetStripeSubscription(*subRecord.StripeSubscriptionId)
		if stripeErr == nil && stripeSub != nil {
			stripeStatus := string(stripeSub.Status)
			periodEnd := billing_service.GetSubscriptionPeriodEnd(stripeSub)
			updateErr := osm.UpdateStatus(*subRecord.StripeSubscriptionId, stripeStatus, periodEnd)
			if updateErr != nil {
				fmt.Printf("failed to sync subscription status from Stripe: %v\n", updateErr)
			}
			if stripeStatus == "active" || stripeStatus == "trialing" || stripeStatus == "past_due" {
				return stripeStatus, tierKey
			}
			status = stripeStatus
		} else if stripeErr != nil {
			fmt.Printf("failed to query Stripe subscription %s: %v\n", *subRecord.StripeSubscriptionId, stripeErr)
			// Fail open: if local status is empty, assume active to allow login
			if status == "" {
				return "active", tierKey
			}
		}
	}

	// 3. Fallback B: Check pending_registration records
	sm := subscription_model.GetSubscriptionModel()

	var pending *entity_public.PendingRegistration
	var pendingErr error

	if email != "" {
		pending, pendingErr = sm.GetPendingRegistrationByEmail(email)
	}
	if (pending == nil || pendingErr != nil) && cpf != "" {
		pending, pendingErr = sm.GetPendingRegistrationByCpf(cpf)
	}

	if pending != nil && pendingErr == nil && pending.StripeCheckoutSessionId != nil && *pending.StripeCheckoutSessionId != "" {
		checkoutSession, checkoutErr := billing_service.GetStripeCheckoutSession(*pending.StripeCheckoutSessionId)
		if checkoutErr == nil && checkoutSession != nil {
			if checkoutSession.Status == stripe.CheckoutSessionStatusComplete && checkoutSession.Subscription != nil {
				subscriptionID := checkoutSession.Subscription.ID
				customerID := ""
				if checkoutSession.Customer != nil {
					customerID = checkoutSession.Customer.ID
				}

				// Get subscription details
				stripeSub, subErr := billing_service.GetStripeSubscription(subscriptionID)
				var periodEnd time.Time
				var stripeStatus string
				if subErr == nil && stripeSub != nil {
					periodEnd = billing_service.GetSubscriptionPeriodEnd(stripeSub)
					stripeStatus = string(stripeSub.Status)
				} else {
					periodEnd = time.Now().Add(time.Hour * 24 * 30)
					stripeStatus = "active"
					if subErr != nil {
						fmt.Printf("failed to query Stripe subscription %s for pending reg: %v\n", subscriptionID, subErr)
					}
				}

				// Resolve tier key
				tierKey = ""
				if pending.StripePriceID != nil && *pending.StripePriceID != "" {
					resolvedTier, tierErr := billing_service.ResolveTierKey(*pending.StripePriceID)
					if tierErr == nil {
						tierKey = resolvedTier
					}
				}
				if tierKey == "" {
					tierKey = checkoutSession.Metadata["tier_key"]
				}
				if tierKey == "" {
					tierKey = "pro"
				}

				// Complete the pending registration
				completeErr := sm.CompletePendingRegistrationFromLogin(*pending, customerID, subscriptionID, stripeStatus, periodEnd, tierKey)
				if completeErr != nil {
					fmt.Printf("failed to complete pending registration from login: %v\n", completeErr)
				} else {
					if stripeStatus == "active" || stripeStatus == "trialing" || stripeStatus == "past_due" {
						return stripeStatus, tierKey
					}
					status = stripeStatus
				}
			}
		} else if checkoutErr != nil {
			fmt.Printf("failed to query Stripe checkout session %s: %v\n", *pending.StripeCheckoutSessionId, checkoutErr)
			// Fail open for pending registrations too
			if status == "" {
				return "active", tierKey
			}
		}
	}

	return status, tierKey
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
	var toast entity_public.Toast

	if err != nil || user == nil {
		if user == nil {
			sm := subscription_model.GetSubscriptionModel()
			pending, _ := sm.GetPendingRegistrationByCpf(cpf)
			if pending != nil {
				dErr := sm.DeletePendingRegistration(pending.Id)
				if dErr != nil {
					toast = entity_public.GetWarningToast("Instabilidade Interna", "volte mais tarde")
					return credentials{}, &toast
				}
				toast = entity_public.GetInfoToast("Criação de conta pendente. Por favor, crie novamente", "é necessário finalizar o pagamento")
				return credentials{}, &toast
			}
		}
		toast = entity_public.GetWarningToast("Credenciais inválidas", "")
		return credentials{}, &toast
	}

	// Check subscription status before creating session
	status, tierKey := resolveSubscriptionForLogin(user.Farm, user.Email, user.Cpf)
	if status != "active" && status != "trialing" && status != "past_due" {
		toast = entity_public.GetWarningToast("Assinatura inativa", "subscription-inactive")
		return credentials{Farm: user.Farm}, &toast
	}

	// Create session
	sessionID, sessionErr := CreateSession(user.Id, ipAddress, userAgent)
	if sessionErr != nil {
		toast = entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}

	token, tokenErr := createToken(user.Name, user.Email, user.Farm, user.Role, user.Id, sessionID, tierKey)
	if tokenErr != nil {
		toast = entity_public.GetErrorToast("Desculpe, houve um erro interno :(", "")
		return credentials{}, &toast
	}
	return credentials{Token: token, Username: user.Name, Farm: user.Farm}, nil
}

func Create(newUser entity_public.NewUser) entity_public.CreateUserResult {
	if newUser.Passwd != newUser.PasswdConfirm {
		return entity_public.CreateUserResult{
			Toast: entity_public.GetWarningToast("Confirmação de senha incorreta", ""),
		}
	}

	if len(newUser.AdditionalIEs) > 0 && newUser.Cpf != newUser.OwnerDocument {
		return entity_public.CreateUserResult{
			Toast: entity_public.GetWarningToast("CPF inválido para cadastrar mais de uma inscrição estadual", "CPF deve ser o mesmo do proprietário"),
		}
	}

	fcm := farm_config_model.GetFarmConfigModel()
	farm, err := fcm.GetFarmByInscricaoEstadual(newUser.InscricaoEstadual)
	if err != nil {
		return entity_public.CreateUserResult{
			Toast: entity_public.GetErrorToast(err.Error(), ""),
		}
	}

	um := user_model.GetUserModel()
	existsAndIsActive, err := um.ExistsAndIsActive(newUser.Cpf)
	if existsAndIsActive == true {
		return entity_public.CreateUserResult{
			Toast: entity_public.GetWarningToast("CPF em uso em outro armazém", "Um adm precisa desativa-lo para cadastra-lo aqui"),
		}
	}

	// Case 1: Farm exists by IE → employee joining existing farm
	if farm != nil {
		created, err := um.CreateUserApproval(newUser, farm.Id)
		if !created || err != nil {
			return entity_public.CreateUserResult{
				Toast: entity_public.GetErrorToast(err.Error(), ""),
			}
		}
		field_service.AddField(entity_public.Field{
			Name:     "Externo",
			Hectares: decimal.Zero,
			Farm:     farm.Id,
		})

		return entity_public.CreateUserResult{
			Toast: entity_public.GetSuccessToast("Usuário enviado para aprovação", ""),
		}
	}

	// Case 2: New owner registering with multiple IEs
	allIEs := []string{newUser.InscricaoEstadual}
	if newUser.AdditionalIEs != nil {
		allIEs = append(allIEs, newUser.AdditionalIEs...)
	}

	// Validate all IEs
	newFarmCount := 0
	for _, ie := range allIEs {
		existingFarm, lookupErr := fcm.GetFarmByInscricaoEstadual(ie)
		if lookupErr != nil {
			return entity_public.CreateUserResult{
				Toast: entity_public.GetErrorToast(lookupErr.Error(), ""),
			}
		}
		if existingFarm != nil {
			// IE already exists and is owned by someone else → reject
			if existingFarm.OwnerDocument != nil && *existingFarm.OwnerDocument != newUser.OwnerDocument {
				return entity_public.CreateUserResult{
					Toast: entity_public.GetWarningToast(
						fmt.Sprintf("Inscrição estadual %s já registrada por outro proprietário", ie),
						""),
				}
			}
		} else {
			newFarmCount++
		}
	}

	// If there are new farms to create, require payment
	if newFarmCount > 0 {
		quantity := int64(len(allIEs))
		newUser.Role = "admin"
		checkoutURL, toast := billing_service.CreatePendingAndCheckout(newUser, newUser.PriceID, quantity)
		return entity_public.CreateUserResult{
			Toast:       toast,
			CheckoutURL: checkoutURL,
		}
	}

	// All IEs already exist and are owned by this owner → no payment needed, just add access
	// (This is unlikely during initial registration, but handled for completeness)
	return entity_public.CreateUserResult{
		Toast: entity_public.GetWarningToast("Todas as inscrições estaduais informadas já estão cadastradas", ""),
	}
}

func GetFarmsByOwnerDocument(ownerDocument string, ownerDocumentType int) ([]entity_public.Farm, error) {
	fcm := farm_config_model.GetFarmConfigModel()
	return fcm.GetFarmsByOwnerDocument(ownerDocument, ownerDocumentType)
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

func GetTierKeyFromToken(sessionId string) string {
	allocatedClaims := &ArmazendaUserClaims{}
	token, err := jwt.ParseWithClaims(sessionId, allocatedClaims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil || token == nil || !token.Valid {
		return ""
	}

	retrievedClaims := token.Claims.(*ArmazendaUserClaims)
	return retrievedClaims.TierKey
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
