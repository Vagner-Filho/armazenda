package main

import (
	"armazenda/model/armazenda_database"
	"armazenda/model/crop_model"
	"armazenda/model/entry_model"
	"armazenda/model/field_model"
	"armazenda/model/vehicle_model"
	"armazenda/router/buyer_router"
	"armazenda/router/crop_router"
	"armazenda/router/departure_router"
	"armazenda/router/entry_router"
	"armazenda/router/field_router"
	"armazenda/router/user_router"
	"armazenda/router/vehicle_router"
	"armazenda/service/vehicle_service"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed assets
var assetsFS embed.FS

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
	crop_model.InitCropModel(conn)
	field_model.InitFieldModel(conn)
	vehicle_model.InitVehicleModel(conn)
	entry_model.InitEntryModel(conn)

	entry_model.InitGrainMap()

	router := gin.Default()

	html := template.Must(template.ParseFS(templatesFS, "templates/*.html", "templates/**/*.html"))
	router.SetHTMLTemplate(html)

	router.StaticFS("/public", http.FS(assetsFS))
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	user_router.UserRoutes(router)

	router.GET("/user/form", user_router.GetUserForm)

	router.GET("/romaneio", entry_router.GetRomaneioPage)

	router.GET("/entry/list", entry_router.GetEntryContent)

	router.GET("/entry/filters", entry_router.GetEntryFiltersForm)

	router.GET("/entry/form", func(c *gin.Context) {
		vehicles, _ := vehicle_service.GetVehicles()

		cropModel, _ := crop_model.GetCropModel()
		crops, _ := cropModel.GetCrops()

		fieldModel, _ := field_model.GetFieldModel()
		fields, _ := fieldModel.GetFields()
		c.HTML(http.StatusOK, "entry-form", gin.H{
			"Fields":   fields,
			"Vehicles": vehicles,
			"Crops":    crops,
		})
	})

	router.GET("/crop/form", crop_router.GetCropForm)
	router.POST("/crop", crop_router.AddCrop)

	router.GET("/entry/form/:id", entry_router.GetEntryForm)
	router.POST("/entry", entry_router.AddEntry)
	router.PUT("/entry/:id", entry_router.PutEntry)
	router.DELETE("/entry/:id", entry_router.DeleteEntry)
	router.POST("/entry/filter", entry_router.FilterEntries)
	router.POST("/field", field_router.AddField)
	router.GET("/entry/field/form", field_router.GetFieldForm)

	router.POST("/departure/filter", departure_router.FilterDepartures)
	router.GET("/departure/list", departure_router.GetDepartureContent)
	router.GET("/departure/form", departure_router.GetDepartureForm)
	router.GET("/buyer/form", buyer_router.GetBuyerForm)
	router.GET("/departure/form/:id", departure_router.GetFilledDepartureForm)
	router.POST("/departure", departure_router.AddDeparture)
	router.POST("/buyer/personal", buyer_router.AddBuyerPerson)
	router.POST("/buyer/company", buyer_router.AddBuyerCompany)
	router.PUT("/departure/:id", departure_router.PutDeparture)
	router.DELETE("/departure/:id", departure_router.DeleteDeparture)

	router.GET("/vehicle/form", vehicle_router.GetVehiclesForm)
	router.POST("/vehicle", vehicle_router.AddVehicle)

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
