package main

import (
	"armazenda/model/grain_model"
	"armazenda/router/grain_router"
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
	grain_model.InitGrainMap()
	router := gin.Default()
	html := template.Must(template.ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(html)

    router.StaticFS("/public", http.FS(assetsFS))
	// router.Static("/assets/static/css", "./assets/static/css")
	// router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", gin.H{})
	})

	router.GET("/grao", grain_router.GetGrains)

	router.GET("/grao/form", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addEntryDialog", gin.H{})
	})

    router.GET("/grao/form/:id", grain_router.GetEntryForm)

	router.POST("/grao", grain_router.AddGrain)
    router.DELETE("/grao/:id", grain_router.DeleteGrain)

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
