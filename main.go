package main

import (
	"armazenda/model/armazenda_database"
	"armazenda/model/crop_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	model_error "armazenda/model/error"
	"armazenda/model/farm_config_model"
	"armazenda/model/field_model"
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
	"armazenda/router/person_router"
	"armazenda/router/report_router"
	"armazenda/router/stats_router"
	"armazenda/router/user_approval_router"
	"armazenda/router/user_router"
	"armazenda/router/vehicle_router"
	"armazenda/service/user_service"
	"context"
	"embed"
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

func authenticate(c *gin.Context) {
	path := c.FullPath()
	if path == "/" || path == "/user" || path == "/login" || strings.Contains(path, "/public") || path == "/user/form" {
		c.Next()
		return
	}

	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		fmt.Printf("\nverifyErr:\n %+v\n", verifyErr)
		c.Status(http.StatusUnauthorized)
		return
	}
	c.Next()
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
	router.Use(authenticate)

	html := template.Must(template.ParseFS(templatesFS, "templates/*.html", "templates/**/*.html"))
	router.SetHTMLTemplate(html)

	router.StaticFS("/public", http.FS(assetsFS))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
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
