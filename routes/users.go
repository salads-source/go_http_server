package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/salads-source/go_http_server/models"
)

func signup(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request"})
		return
	}

	err = user.Save()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user, try again later"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}
