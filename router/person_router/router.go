package person_router

import (
	entity_public "armazenda/entity/public"
	person_service "armazenda/service/person"
	"armazenda/service/user_service"
	"armazenda/view"
	person_view "armazenda/view/person"
	"net/http"
	"strconv"

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

	ie := c.PostForm("inscricaoEstadual")
	newLegalPerson.Person.Ie = ie

	person, toast := person_service.AddLegalPerson(newLegalPerson)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusCreated, "person-list-item", person_view.PersonListItemView{
		PersonDisplay:    person,
		BaseTemplateData: view.BaseTemplateData{CSPNonce: nonce.(string)},
	})
}

func addNaturalPerson(c *gin.Context) {
	var newNatural entity_public.NaturalPerson
	err := c.Bind(&newNatural)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newNatural.Person.Farm = farm

	ie := c.PostForm("inscricaoEstadual")
	newNatural.Person.Ie = ie
	person, toast := person_service.AddNaturalPerson(newNatural)
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
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "person-list-item", person_view.PersonListItemView{
		PersonDisplay:    person,
		BaseTemplateData: view.BaseTemplateData{CSPNonce: nonce.(string)},
	})
}

func getPersonForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	progressions := person_view.GetProgressionsForForm(farm)
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "person-form", gin.H{"Progressions": progressions, "CSPNonce": nonce.(string)})
}

func getFilledNaturalPersonForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	person, t := person_view.GetFilledNaturalPersonForm(uint32(converted), farm)
	if t != nil {
		c.Header("HX-Trigger", string(t.ToJson()))
		return
	}
	nonce, _ := c.Get("csp_nonce")
	person.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "person-form", person)
}

func getFilledLegalPersonForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	pform, t := person_view.GetFilledLegalPersonForm(uint32(converted), farm)
	if t != nil {
		c.Header("HX-Trigger", string(t.ToJson()))
		return
	}
	nonce, _ := c.Get("csp_nonce")
	pform.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "person-form", pform)
}

func getPersonPage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	const limit = 15

	peopleData, toast := person_service.FilterPerson(entity_public.PersonFilter{}, farm, page, limit)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	pageData := gin.H{
		"People":      peopleData.People,
		"CurrentPage": peopleData.CurrentPage,
		"TotalPages":  peopleData.TotalPages,
		"NextPage":    peopleData.NextPage,
		"PrevPage":    peopleData.PrevPage,
		"HasNextPage": peopleData.HasNextPage,
		"HasPrevPage": peopleData.HasPrevPage,
		"NoContent":   peopleData.NoContent,
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "people-table", pageData)
		return
	}

	nonce, _ := c.Get("csp_nonce")
	pageData["CSPNonce"] = nonce.(string)
	c.HTML(http.StatusOK, "person.html", pageData)
}

func updateNaturalPerson(c *gin.Context) {
	var updatedNatural entity_public.NaturalPerson
	err := c.Bind(&updatedNatural)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	updatedNatural.Person.Farm = farm
	updatedNatural.Person.Id = uint32(converted)

	ie := c.PostForm("inscricaoEstadual")
	updatedNatural.Person.Ie = ie

	person, toast := person_service.UpdateNaturalPerson(updatedNatural)
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
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "person-list-item", person_view.PersonListItemView{
		PersonDisplay:    person,
		BaseTemplateData: view.BaseTemplateData{CSPNonce: nonce.(string)},
	})
}

func updateLegalPerson(c *gin.Context) {
	var updatedLegal entity_public.LegalPerson
	err := c.Bind(&updatedLegal)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	updatedLegal.Person.Farm = farm
	updatedLegal.Person.Id = uint32(converted)

	ie := c.PostForm("inscricaoEstadual")
	updatedLegal.Person.Ie = ie

	person, toast := person_service.UpdateLegalPerson(updatedLegal)
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
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "person-list-item", person_view.PersonListItemView{
		PersonDisplay:    person,
		BaseTemplateData: view.BaseTemplateData{CSPNonce: nonce.(string)},
	})
}

func UsePersonRoutes(router *gin.Engine) {
	router.GET("/pessoa", getPersonPage)
	router.GET("/person/form", getPersonForm)
	router.POST("/person/natural", addNaturalPerson)
	router.POST("/person/legal", addLegalPerson)
	router.GET("/person/legal/form/:id", getFilledLegalPersonForm)
	router.GET("/person/natural/form/:id", getFilledNaturalPersonForm)
	router.PUT("/person/natural/:id", updateNaturalPerson)
	router.PUT("/person/legal/:id", updateLegalPerson)
}
