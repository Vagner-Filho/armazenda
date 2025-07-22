package main

import (
	"armazenda/model/armazenda_database"
	"armazenda/model/crop_model"
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/model/field_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/model/report_model"
	"armazenda/model/user_model"
	"armazenda/model/vehicle_model"
	"armazenda/router/crop_router"
	"armazenda/router/departure_router"
	"armazenda/router/entry_router"
	"armazenda/router/field_router"
	"armazenda/router/person_router"
	"armazenda/router/report_router"
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
		fmt.Printf("\ncookieErr:\n %+v", cookieErr)
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		c.AbortWithStatus(http.StatusUnauthorized)
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

	conn, connErr := armazenda_database.GetDbConnection()

	if connErr != nil {
		fmt.Printf("db connection error: %v \n", connErr.Error())
		return
	}

	if conn == nil {
		fmt.Print("db connection nil\n")
		return
	}

	defer conn.Close(context.Background())

	armazenda_database.InitDb(conn)
	user_model.InitUserModel(conn)
	crop_model.InitCropModel(conn)
	field_model.InitFieldModel(conn)
	vehicle_model.InitVehicleModel(conn)
	entry_model.InitEntryModel(conn)
	departure_model.InitDepartureModel(conn)
	product_model.InitProductModel(conn)
	person_model.InitPersonModel(conn)
	report_model.InitReportModel(conn)

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
