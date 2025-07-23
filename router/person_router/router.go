package person_router

import (
	entity_public "armazenda/entity/public"
	person_service "armazenda/service/person"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func addLegalPerson(c *gin.Context) {
	var newLegalPerson entity_public.LegalPerson
	err := c.Bind(&newLegalPerson)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newLegalPerson.Person.Farm = farm
	person, toast := person_service.AddLegalPerson(newLegalPerson)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusCreated, "person-list-item", person)
}

func addNaturalPerson(c *gin.Context) {
	var newPersonal entity_public.NaturalPerson
	err := c.Bind(&newPersonal)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newPersonal.Person.Farm = farm
	person, toast := person_service.AddNaturalPerson(newPersonal)
	if toast != nil {
		if toast.Type == entity_public.ErrorToast {
			c.Header("HX-Trigger", string(toast.ToJson()))
			c.Status(http.StatusInternalServerError)
			return
		}
		if toast.Type == entity_public.WarningToast {
			c.Header("HX-Trigger", string(toast.ToJson()))
			c.Status(http.StatusUnprocessableEntity)
			return
		}
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusOK, "person-list-item", person)
}

func getPersonForm(c *gin.Context) {
	c.HTML(http.StatusOK, "person-form", gin.H{})
}

func getPersonPage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	people, toast := person_service.FilterPerson(entity_public.PersonFilter{}, farm)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusOK, "person.html", gin.H{
		"People": people,
	})
}

func UsePersonRoutes(router *gin.Engine) {
	router.GET("/pessoa", getPersonPage)
	router.GET("/person/form", getPersonForm)
	router.POST("/person/natural", addNaturalPerson)
	router.POST("/person/legal", addLegalPerson)
}
