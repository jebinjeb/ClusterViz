package main

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/gin-gonic/gin"
	"clusterviz/internal/pkg/server"
	"net/http"
)

func main() {
	var kubeconfigPath string
	flag.StringVar(&kubeconfigPath, "kubeconfig", "C:\\Users\\Administrator\\.kube\\config", "Path to the kubeconfig file")
	flag.Parse()

	// Create Kubernetes clientset
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic("Error building kubeconfig: " + err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
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

	r.LoadHTMLFiles("C:/Program Files/Go/src/clusterviz/internal/pkg/server/cluster_viz.html")

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
		panic("Error starting server: " + err.Error())
	}
}
