package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"clusterviz/internal/pkg/server"
	"net/http"
)

func main() {
	// Create Kubernetes clientset
	clientset, err := server.NewClientset()
    if err != nil {
        panic("Error creating Kubernetes client: " + err.Error())
    }
	// Create API instance with the clientset
	api := server.NewAPI(clientset)

	// Set up Gin router
	r := gin.Default()

	// Enable CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, struct{}{})
			return
		}
		c.Next()
	})

	api.Router = r

	// Health and version endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "1.0.0",
		})
	})

	// Cluster visualization endpoint
	r.GET("/clusterviz", api.GetClusterViz)

	// Run the server on port 8080
	if err := r.Run(":8080"); err != nil {
		fmt.Println("Error starting server: " + err.Error())
	}
}
