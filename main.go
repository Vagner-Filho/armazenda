package main

import (
	"armazenda/model/entry_model"
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
	html := template.Must(template.ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(html)

    router.StaticFS("/public", http.FS(assetsFS))
	// router.Static("/assets/static/css", "./assets/static/css")
	// router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", gin.H{})
	})

	router.GET("/grao", entry_router.GetEntries)

	router.GET("/grao/form", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addEntryDialog", gin.H{})
	})

    router.GET("/grao/form/:id", entry_router.GetEntryForm)

	router.POST("/grao", entry_router.AddEntry)
    router.PUT("/grao/:id", entry_router.PutEntry)
    router.DELETE("/grao/:id", entry_router.DeleteEntry)

    router.GET("/vehicle/plate", vehicle_router.GetVehiclesSelector)
    router.POST("/vehicle/plate", vehicle_router.PostPlate)

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
