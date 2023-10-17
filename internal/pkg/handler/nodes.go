package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"log"
	"github.com/mitchellh/go-homedir"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ErrorResponse struct {
    Error string `json:"error"`
}

type PodCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastProbeTime      metav1.Time `json:"lastProbeTime"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

type NodeInfo struct {
	Name          string         `json:"name"`
	NumberOfPods  int            `json:"numberOfPods"`
	Pods          []string       `json:"pods"`
	PodConditions []PodCondition `json:"podConditions"`
}

type PodResponse struct {
	Name           string           `json:"name"`
	PodConditions  []PodCondition   `json:"podConditions"`
}

type NodeResponse struct {
	Name         string         `json:"name"`
	NumberOfPods int            `json:"numberOfPods"`
	Pods         []PodResponse  `json:"pods"`
}
type FullNodeResponse struct {
	TotalNodes int            `json:"TotalNodes"`
	NodeNames  []string       `json:"NodeNames"`
	Nodes      []NodeResponse `json:"Nodes"`
}


func GetNodeInfo() ([]NodeInfo, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var nodesInfo []NodeInfo

	// Load Kubernetes configuration
	var kubeconfigPath string
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("error determining home directory: %v", err)
	}
	kubeconfigPath = home + "/.kube/config"

	// Load kubeconfig file and create a Kubernetes clientset using clientcmd
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	// Get the list of nodes from the cluster
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing Nodes: %v", err)
	}

	// Iterate through nodes and gather information about pods scheduled on each node
	for _, node := range nodes.Items {
		nodeInfo := NodeInfo{
			Name:          node.Name,
			NumberOfPods:  0,
			Pods:          []string{},
			PodConditions: []PodCondition{},
		}

		// Get the list of pods scheduled on the current node
		pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{FieldSelector: "spec.nodeName=" + node.Name})
		if err != nil {
			return nil, fmt.Errorf("error listing Pods on node %s: %v", node.Name, err)
		}

		// Update the number of pods for the current node
		nodeInfo.NumberOfPods = len(pods.Items)

		// Iterate through pods and gather status information
		for _, pod := range pods.Items {
			podConditions := []PodCondition{}

			for _, condition := range pod.Status.Conditions {
				lastProbeTime := condition.LastProbeTime.Time
				if lastProbeTime.IsZero() {
					lastProbeTime = time.Now()
				}

				switch condition.Type {
				case v1.PodReady:
					if condition.Status == v1.ConditionTrue {
						podConditions = append(podConditions, PodCondition{
							Type:               "Ready",
							Status:             string(condition.Status),
							LastProbeTime:      metav1.Time{Time: lastProbeTime},
							LastTransitionTime: condition.LastTransitionTime,
						})
					} else {
						podConditions = append(podConditions, PodCondition{
							Type:               "Not Ready",
							Status:             string(condition.Status),
							LastProbeTime:      metav1.Time{Time: lastProbeTime},
							LastTransitionTime: condition.LastTransitionTime,
						})
					}
				case v1.PodInitialized:
					if condition.Status == v1.ConditionTrue {
						podConditions = append(podConditions, PodCondition{
							Type:               "Initialized",
							Status:             string(condition.Status),
							LastProbeTime:      metav1.Time{Time: lastProbeTime},
							LastTransitionTime: condition.LastTransitionTime,
						})
					} else {
						podConditions = append(podConditions, PodCondition{
							Type:               "Not Initialized",
							Status:             string(condition.Status),
							LastProbeTime:      metav1.Time{Time: lastProbeTime},
							LastTransitionTime: condition.LastTransitionTime,
						})
					}
				default:
					podConditions = append(podConditions, PodCondition{
						Type:               string(condition.Type),
						Status:             string(condition.Status),
						LastProbeTime:      metav1.Time{Time: lastProbeTime},
						LastTransitionTime: condition.LastTransitionTime,
					})
				}
			}
			nodeInfo.Pods = append(nodeInfo.Pods, pod.Name)
			nodeInfo.PodConditions = append(nodeInfo.PodConditions, podConditions...)
		}
		nodesInfo = append(nodesInfo, nodeInfo)
	}
	return nodesInfo, nil
}


func NodeInfoHandler(w http.ResponseWriter, r *http.Request, nodesInfo []NodeInfo) {
    var nodeResponses []NodeResponse
    var nodeNames []string

    // Collect node names and populate NodeResponses
    for _, node := range nodesInfo {
        nodeNames = append(nodeNames, node.Name)

        var podResponses []PodResponse
        for i, pod := range node.Pods {
            if i < len(node.PodConditions) {
                condition := node.PodConditions[i]
                podResponse := PodResponse{
                    Name: pod,
                    PodConditions: []PodCondition{
                        {
                            Type:               condition.Type,
                            Status:             condition.Status,
                            LastProbeTime:      condition.LastProbeTime,
                            LastTransitionTime: condition.LastTransitionTime,
                        },
                    },
                }
                podResponses = append(podResponses, podResponse)
            }
        }

        nodeResponse := NodeResponse{
            Name:         node.Name,
            NumberOfPods: len(podResponses),
            Pods:         podResponses,
        }

        nodeResponses = append(nodeResponses, nodeResponse)
    }

    // Create FullNodeResponse
    fullResponse := FullNodeResponse{
        TotalNodes: len(nodeNames),
        NodeNames:  nodeNames,
        Nodes:      nodeResponses,
    }

    // Convert response to JSON
    jsonData, err := json.Marshal(fullResponse)
    if err != nil {
        // Log the error for debugging purposes
        log.Printf("Error converting to JSON: %v", err)

        // Return an error response as JSON
        errorResponse := ErrorResponse{Error: "Internal server error"}
        jsonError, jsonErr := json.Marshal(errorResponse)
        if jsonErr != nil {
            // If there's an error even in generating the error response, return a plain text error message
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Set Content-Type header to application/json and status code to http.StatusInternalServerError
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write(jsonError)
        return
    }

    // Set Content-Type header to application/json and status code to http.StatusOK
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    // Write JSON response to the browser
    w.Write(jsonData)
}


func PrintNodeInfo(nodesInfo []NodeInfo) {
	for _, node := range nodesInfo {
		fmt.Println("NODE INFORMATION")
		fmt.Printf("Node: %s\n", node.Name)
		fmt.Printf("Number of Pods: %d\n", len(node.Pods))
		fmt.Println("Pods:")

		for i, pod := range node.Pods {
			fmt.Printf("Name: %s\n", pod)
			if i < len(node.PodConditions) {
				condition := node.PodConditions[i]
				fmt.Println("  - Pod Conditions:")
				fmt.Printf("    - Type: %s\n", condition.Type)
				fmt.Printf("    - Status: %s\n", condition.Status)
				fmt.Printf("    - LastProbeTime: %v\n", condition.LastProbeTime)
				fmt.Printf("    - LastTransitionTime: %v\n", condition.LastTransitionTime)
			}
			fmt.Println()
		}
	}
}







