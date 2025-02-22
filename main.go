package main

import (
	"armazenda/model/entry_model"
	"armazenda/router/departure_router"
	"armazenda/router/entry_router"
	"armazenda/router/vehicle_router"
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
	entry_model.InitGrainMap()
	router := gin.Default()
	html := template.Must(template.ParseFS(templatesFS, "templates/*.html", "templates/**/*.html"))
	println(html.DefinedTemplates())
	router.SetHTMLTemplate(html)

	router.StaticFS("/public", http.FS(assetsFS))
	// router.Static("/assets/static/css", "./assets/static/css")
	// router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", gin.H{})
	})

	router.GET("/romaneio", entry_router.GetEntries)

	router.GET("/entry/list", entry_router.GetEntriesTable)

	router.GET("/entry/form", func(c *gin.Context) {
		var fields []entry_router.Field
		for _, field := range entry_router.GetFields() {
			newF := entry_router.Field{}
            newF.Selected = false
            newF.Name = field.Name
            newF.Id = field.Id
			fields = append(fields, newF)
		}
        var vehicles []entry_router.Vehicle
        for _, vehicle := range vehicle_router.GetVehicles() {
            newV := entry_router.Vehicle{}
            newV.Name = vehicle.Name
            newV.Plate = vehicle.Plate
            vehicles = append(vehicles, newV)
        }
		c.HTML(http.StatusOK, "add-entry-dialog", gin.H{
			"Fields":   fields,
			"Vehicles": vehicles,
		})
	})

	router.GET("/entry/form/:id", entry_router.GetEntryForm)

	router.POST("/entry", entry_router.AddEntry)
	router.PUT("/entry/:id", entry_router.PutEntry)
	router.DELETE("/entry/:id", entry_router.DeleteEntry)

	router.POST("/entry/field", entry_router.AddField)
	router.GET("/entry/field/form", entry_router.GetFieldForm)

	router.GET("/departure/list", departure_router.GetDepartures)
	router.GET("/departure/form", departure_router.GetDepartureForm)
	router.GET("/departure/form/:id", departure_router.GetFilledDepartureForm)
    router.POST("/departure", departure_router.AddDeparture)
    router.PUT("/departure/:id", departure_router.PutDeparture)

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
