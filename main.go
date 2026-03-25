package main

import (
	"armazenda/model/armazenda_database"
	"armazenda/model/crop_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	model_error "armazenda/model/error"
	"armazenda/model/farm_config_model"
	"armazenda/model/field_model"
	"armazenda/model/humidity_progression_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/model/report_model"
	"armazenda/model/stats_model"
	"armazenda/model/user_approval_model"
	"armazenda/model/user_model"

	"armazenda/model/vehicle_model"
	"armazenda/router/crop_router"
	"armazenda/router/departure_router"
	"armazenda/router/entry_router"
	"armazenda/router/farm_config_router"
	"armazenda/router/field_router"
	"armazenda/router/humidity_progression_router"
	"armazenda/router/person_router"
	"armazenda/router/report_router"
	"armazenda/router/stats_router"
	"armazenda/router/sync_router"
	"armazenda/router/template_router"
	"armazenda/router/user_approval_router"
	"armazenda/router/user_router"
	"armazenda/router/vehicle_router"
	"armazenda/service/user_service"
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed assets
var assetsFS embed.FS

func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func authenticate(c *gin.Context) {
	path := c.FullPath()
	if path == "/" || path == "/user" || path == "/login" || strings.Contains(path, "/public") || path == "/user/form" || path == "/auth/google/login" || path == "/auth/google/callback" || path == "/auth/microsoft/login" || path == "/auth/microsoft/callback" || path == "/user/microsoft-register" || path == "/user/google-register" {
		c.Next()
		return
	}

	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		c.Abort()
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		fmt.Printf("\nverifyErr:\n %+v\n", verifyErr)
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		c.Abort()
		return
	}
	c.Next()
}

func setPublicAssetsHeaders(c *gin.Context) {
	path := c.FullPath()
	if strings.Contains(path, "/public") && strings.Contains(path, "filepath") {
		c.Header("Cache-Control", "public, max-age=28800")
		if strings.Contains(c.Request.URL.Path, "htmx.min.js.gz") {
			c.Header("Content-Encoding", "gzip")
			c.Header("Content-Type", "application/javascript")
		}
		return
	}
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func setSecurityHeaders(c *gin.Context) {
	//nonce := c.Request.Header.Get("X-nonce")

	//if nonce == "" {
	//	nonce = generateNonce()
	//}

	nonce := generateNonce()
	c.Set("csp_nonce", nonce)

	c.Header("Content-Security-Policy",
		"script-src 'nonce-"+nonce+"' 'strict-dynamic' https: 'nonce-"+nonce+"' 'wasm-unsafe-eval' 'unsafe-inline'; "+
			"object-src 'none'; "+
			"base-uri 'none'; "+
			"frame-ancestors 'none'; ")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	c.Header("Cross-Origin-Opener-Policy", "same-origin")
}

func adminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCookie, cookieErr := c.Request.Cookie("session_id")
		if cookieErr != nil {
			c.HTML(http.StatusUnauthorized, "401", gin.H{})
			c.Abort()
			return
		}

		verifyErr := user_service.VerifyToken(sessionCookie.Value)
		if verifyErr != nil {
			c.HTML(http.StatusUnauthorized, "401", gin.H{})
			c.Abort()
			return
		}

		if !user_service.IsAdmin(sessionCookie.Value) {
			c.String(http.StatusForbidden, "Acesso negado. Apenas administradores podem realizar esta ação.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	pool, connErr := armazenda_database.GetDbPool()

	if connErr != nil {
		fmt.Printf("db connection error: %v \n", connErr.Error())
		return
	}

	if pool == nil {
		fmt.Print("db connection nil\n")
		return
	}

	defer pool.Close()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a connection: %v\n", err)
		return
	}
	armazenda_database.InitDb(conn.Conn())
	conn.Release()

	user_model.InitUserModel(pool)
	crop_model.InitCropModel(pool)
	field_model.InitFieldModel(pool)
	vehicle_model.InitVehicleModel(pool)
	entry_model.InitEntryModel(pool)
	departure_model.InitDepartureModel(pool)
	product_model.InitProductModel(pool)
	person_model.InitPersonModel(pool)
	humidity_progression_model.InitHumidityProgressionModel(pool)
	report_model.InitReportModel(pool)
	farm_config_model.InitFarmConfigModel(pool)
	user_approval_model.InitUserApprovalModel(pool)
	stats_model.InitStatsModel(pool)
	model_error.InitLoggerModel(pool)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if valuer, ok := field.Interface().(decimal.Decimal); ok {
				return valuer.String()
			}
			return nil
		}, decimal.Decimal{})
	}

	router := gin.Default()
	router.Use(setSecurityHeaders, authenticate, setPublicAssetsHeaders)

	funcMap := template.FuncMap{
		"deref": func(p *uint32) uint32 {
			if p == nil {
				return 0
			}
			return *p
		},
		"dict": dict,
		"decIsZero": func(v interface{}) bool {
			switch d := v.(type) {
			case decimal.Decimal:
				return d.IsZero()
			case float64:
				return d == 0
			default:
				return false
			}
		},
		"decIsNotZero": func(v interface{}) bool {
			switch d := v.(type) {
			case decimal.Decimal:
				return !d.IsZero()
			case float64:
				return d != 0
			default:
				return false
			}
		},
	}
	html := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html", "templates/**/*.html"))
	router.SetHTMLTemplate(html)

	// Initialize template router for serving pre-rendered templates
	template_router.InitTemplateRouter(templatesFS, html)

	router.StaticFS("/public", http.FS(assetsFS))

	router.GET("/", func(c *gin.Context) {
		nonce, _ := c.Get("csp_nonce")
		c.HTML(http.StatusOK, "login.html", gin.H{"CSPNonce": nonce.(string)})
	})

	user_router.UserRoutes(router)
	entry_router.UseEntryRoutes(router)
	departure_router.UseDepartureRoutes(router)
	crop_router.UseCropRoutes(router)
	field_router.UseFieldRoutes(router)
	vehicle_router.UseVehicleRouter(router)
	person_router.UsePersonRoutes(router)
	report_router.UseReportRoutes(router)
	farm_config_router.UseFarmConfigRouter(router)
	user_approval_router.UserApprovalRoutes(router)
	stats_router.UseStatsRoutes(router)
	sync_router.UseSyncRoutes(router)
	humidity_progression_router.UseHumidityProgressionRouter(router)
	humidity_progression_router.UseHumidityProgressionHtmlRoutes(router)
	template_router.UseTemplateRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	ipv6Addr := "::" // Listen on all IPv6 addresses
	if envIP := os.Getenv("IP"); envIP != "" {
		ipv6Addr = envIP
	}

	address := fmt.Sprintf("[%s]:%s", ipv6Addr, port)
	router.Run(address)
}
