package main

import (
	"armazenda/model/grain_model"
	"armazenda/router/grain_router"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	grain_model.InitGrainMap()
	router := gin.Default()
	html := template.Must(template.ParseGlob("templates/*"))
	router.SetHTMLTemplate(html)

	router.Static("/assets/static/css", "./assets/static/css")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", gin.H{})
	})

	router.GET("/grao", grain_router.GetGrains)

	router.GET("/grao/form", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addEntryDialog", gin.H{})
	})

	router.POST("/grao", grain_router.AddGrain)
	router.Run(":3000")
}
