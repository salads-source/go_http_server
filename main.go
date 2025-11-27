package main

import (
	"github.com/gin-gonic/gin"
	"github.com/salads-source/go_http_server/db"
	"github.com/salads-source/go_http_server/routes"
)

func main() {
	db.InitDB()
	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") // loaclhost:8080
}
