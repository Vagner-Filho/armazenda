package sync_router

import (
	"net/http"
	"time"

	"armazenda/service/departure_service"
	"armazenda/service/entry_service"
	"armazenda/service/person"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
)

// SyncEntry represents an entry for sync (simplified structure)
type SyncEntry struct {
	Id          uint32          `json:"id"`
	Field       uint16          `json:"field"`
	Crop        uint8           `json:"crop"`
	Vehicle     uint16          `json:"vehicle"`
	CargoWeight SyncCargoWeight `json:"cargoWeight"`
	Analysis    SyncAnalysis    `json:"analysis"`
	ArrivalDate time.Time       `json:"arrivalDate"`
	Farm        uint32          `json:"farm"`
	Origin      *uint32         `json:"origin,omitempty"`
	ModifiedAt  time.Time       `json:"modifiedAt"`
	Deleted     bool            `json:"deleted,omitempty"`
}

type SyncCargoWeight struct {
	GrossWeight float64 `json:"grossWeight"`
	Tare        float64 `json:"tare"`
	NetWeight   float64 `json:"netWeight"`
}

type SyncAnalysis struct {
	Humidity *float64 `json:"humidity,omitempty"`
	Damage   *float64 `json:"damage,omitempty"`
	Impurity *float64 `json:"impurity,omitempty"`
}

// SyncDeparture represents a departure for sync
type SyncDeparture struct {
	Id            uint32          `json:"id"`
	DepartureDate time.Time       `json:"departureDate"`
	Vehicle       uint16          `json:"vehicle"`
	Crop          uint8           `json:"crop"`
	CargoWeight   SyncCargoWeight `json:"cargoWeight"`
	Analysis      SyncAnalysis    `json:"analysis"`
	Farm          uint32          `json:"farm"`
	Recipient     *uint32         `json:"recipient,omitempty"`
	Origin        *uint32         `json:"origin,omitempty"`
	ModifiedAt    time.Time       `json:"modifiedAt"`
	Deleted       bool            `json:"deleted,omitempty"`
}

// SyncPerson represents a person for sync
type SyncPerson struct {
	Id           uint32           `json:"id"`
	Type         uint8            `json:"type"`
	Name         string           `json:"name"`
	Document     string           `json:"document"`
	IE           string           `json:"ie"`
	Farm         uint32           `json:"farm"`
	PersonConfig SyncPersonConfig `json:"personConfig"`
	ModifiedAt   time.Time        `json:"modifiedAt"`
	Deleted      bool             `json:"deleted,omitempty"`
}

type SyncPersonConfig struct {
	HumidityProgressionId *uint32 `json:"humidityProgressionId"`
	EntrySoyDiscount      float64 `json:"entrySoyDiscount"`
	EntryCornDiscount     float64 `json:"entryCornDiscount"`
}

func getEntriesForSync(c *gin.Context) {
	sid, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	farm := user_service.GetFarmFromToken(sid)
	if farm == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	sinceStr := c.Query("since")
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since parameter", "details": err.Error()})
		return
	}

	entries, err := entry_service.GetEntriesForSync(since, farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func getDeparturesForSync(c *gin.Context) {
	sid, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	farm := user_service.GetFarmFromToken(sid)
	if farm == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	sinceStr := c.Query("since")
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since parameter", "details": err.Error()})
		return
	}

	departures, err := departure_service.GetDeparturesForSync(since, farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, departures)
}

func getPeopleForSync(c *gin.Context) {
	sid, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	farm := user_service.GetFarmFromToken(sid)
	if farm == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	sinceStr := c.Query("since")
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since parameter", "details": err.Error()})
		return
	}

	people, err := person_service.GetPeopleForSync(since, farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, people)
}

func getSyncStatus(c *gin.Context) {
	sid, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	farm := user_service.GetFarmFromToken(sid)
	if farm == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	var req struct {
		Farm     uint32     `json:"farm" binding:"required"`
		LastSync *time.Time `json:"lastSync"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user has access to this farm
	if req.Farm != farm {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Get counts of modified records
	entryCount, err := entry_service.GetModifiedEntryCount(*req.LastSync, req.Farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	departureCount, err := departure_service.GetModifiedDepartureCount(*req.LastSync, req.Farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	personCount, err := person_service.GetModifiedPersonCount(*req.LastSync, req.Farm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasUpdates":      entryCount > 0 || departureCount > 0 || personCount > 0,
		"entriesCount":    entryCount,
		"departuresCount": departureCount,
		"peopleCount":     personCount,
		"serverTimestamp": time.Now().UTC(),
	})
}

func UseSyncRoutes(router *gin.Engine) {
	// API routes for sync - require authentication
	api := router.Group("/api")
	{
		api.GET("/entries", getEntriesForSync)
		api.GET("/departures", getDeparturesForSync)
		api.GET("/people", getPeopleForSync)
		api.POST("/sync/status", getSyncStatus)
	}
}
