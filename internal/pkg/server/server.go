package server

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"

    "github.com/gin-gonic/gin"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"

    "clusterviz/internal/pkg/handler"
)

type API struct {
    clientset *kubernetes.Clientset
    Router    *gin.Engine
}

func NewAPI(clientset *kubernetes.Clientset) *API {
    api := &API{
        clientset: clientset,
        Router:    gin.Default(),
    }
    api.setupRoutes()
    return api
}

func (api *API) setupRoutes() {
    api.Router.GET("/clusterviz", api.GetClusterViz)
}

func NewClientset() (*kubernetes.Clientset, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
    if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
    }

    config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    return clientset, nil
}

func (api *API) GetClusterViz(c *gin.Context) {
	// Use functions from replica.go and deployment.go to get the required information
	replicaInfo, err := handler.GetReplicaCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deploymentsInfo, err := handler.GetDeploymentCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.PrintDeploymentInfo(deploymentsInfo)
	handler.PrintReplicaSetInfo(replicaInfo)

	nodesInfo, err := handler.GetNodeInfo()
    if err != nil {
		fmt.Println("Error getting node info:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
    }
	
    handler.PrintNodeInfo(nodesInfo)

	// Create the response map with custom headings
	response := map[string]interface{}{
		"ClusterVisualization": gin.H{
			"Deployments": deploymentsInfo,
			"ReplicaSets": replicaInfo,
			//"Nodes": nodesInfo,
		},
	}
	handler.NodeInfoHandler(c.Writer, c.Request, nodesInfo)
	// Set the content type header to JSON
	c.Header("Content-Type", "application/json")
	
	// Convert the map to JSON and send it in the response
	c.JSON(http.StatusOK, response)

}