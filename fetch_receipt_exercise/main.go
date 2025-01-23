package main

import (
	"hello-go/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to Debug mode
	gin.SetMode(gin.DebugMode)

	router := gin.Default()
	routes.RegisterRoutes(router)

	router.Run("localhost:8000")
}
