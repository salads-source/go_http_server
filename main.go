package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/salads-source/go_http_server/db"
	"github.com/salads-source/go_http_server/models"
)

func main() {
	db.InitDB()
	server := gin.Default()

	server.GET("/events", getEvents)
	server.POST("/events", createEvent)
	server.Run(":8080") // loaclhost:8080
}

func getEvents(context *gin.Context) {
	events, err := models.GetAllEvents()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch events, try again later"})
	}

	context.JSON(http.StatusOK, events)
}

func createEvent(context *gin.Context) {
	var event models.Event
	err := context.ShouldBindJSON(&event)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
	}

	event.ID = 1
	event.UserID = 1

	err = event.Save()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save event, try again later"})
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Event created!", "event": event})
}
