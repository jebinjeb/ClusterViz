package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/fatih/color"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

var message string

type ReplicaSetResponse struct {
	Data []ReplicaSetInfo `json:"data"`
}

type ReplicaSetInfo struct {
	Namespace string     `json:"namespace"`
	Name      string    `json:"name"`  
	Replicas  int32      `json:"replicas"`
	NodeScheduling string   `json:"node_scheduling"`
	Nodes    string       `json:"nodes"`
}


func GetReplicaCount() ([]ReplicaSetInfo, error) {
	var replicaSetsInfo []ReplicaSetInfo

	// Load Kubernetes configuration
	var kubeconfigPath string
	home, err := homedir.Dir()
	if err != nil {
		return replicaSetsInfo, fmt.Errorf("error determining home directory: %v", err)
	}
	kubeconfigPath = home + "/.kube/config"

	// Load kubeconfig file and create a Kubernetes clientset using clientcmd
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return replicaSetsInfo, fmt.Errorf("error loading kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return replicaSetsInfo, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	// List ReplicaSets in the cluster
	replicaSets, err := clientset.AppsV1().ReplicaSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return replicaSetsInfo, fmt.Errorf("error listing ReplicaSets: %v", err)
	}

	// Iterate through ReplicaSets and gather information
	for _, rs := range replicaSets.Items {
		replicaSetInfo := ReplicaSetInfo{
			Namespace: rs.Namespace,
			Name:      rs.Name,
			Replicas:  *rs.Spec.Replicas,
		}

		// Get nodes where Pods of this ReplicaSet are scheduled
		nodes, err := clientset.CoreV1().Pods(rs.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + rs.Name})
		if err != nil {
			return replicaSetsInfo, fmt.Errorf("error listing Pods for ReplicaSet %s: %v", rs.Name, err)
		}

		var nodeNames []string
		for _, pod := range nodes.Items {
			nodeName := pod.Spec.NodeName
			if nodeName != "" {
				nodeNames = append(nodeNames, nodeName)
			}
		}
		// Join the nodes into a comma-separated string
		nodeList := strings.Join(nodeNames, ", ")
		replicaSetInfo.Nodes = nodeList
		
	
		infoColor := color.New(color.FgGreen).SprintFunc()
		warningColor := color.New(color.FgRed).SprintFunc()
		
		//for _, rs := range replicaSetsInfo {
		
	
		if replicaSetInfo.Replicas == 1 {
				message = fmt.Sprintf("[%s] Replica count is 1. Consider increasing it for high availability.", warningColor("WARNING"))
			} else if replicaSetInfo.Replicas < 1 {
				message = fmt.Sprintf("[%s] Replica count is less than 1. Please update the ReplicaSet.", warningColor("WARNING"))
			} else {
				message = fmt.Sprintf("[%s] Replica count is %d.", infoColor("INFO"), replicaSetInfo.Replicas)
			}
		//fmt.Println(message)
		var nodeScheduling string
		if len(nodeNames)  > 1 {
			nodeScheduling = infoColor("INFO") + " Pods are scheduled on different nodes."
		} else if len(nodeNames) == 1 {
			nodeScheduling = warningColor("WARNING") + " Pods are scheduled on the same node."
		} else {
			nodeScheduling = warningColor("WARNING") + " No Pods are scheduled."
		}
		fmt.Println(nodeScheduling)
		replicaSetInfo.NodeScheduling = nodeScheduling
		replicaSetsInfo = append(replicaSetsInfo, replicaSetInfo)
	}
		
	return replicaSetsInfo, nil
}

func GetReplicaSetInfoHandler(w http.ResponseWriter, r *http.Request) {
	replicaSetsInfo, err := GetReplicaCount()
	if err != nil {
		log.Printf("Error getting replica set information: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errorMessage := map[string]string{"error": err.Error()}
		json.NewEncoder(w).Encode(errorMessage)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := ReplicaSetResponse{
		Data: replicaSetsInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func PrintReplicaSetInfo(replicaSetsInfo []ReplicaSetInfo) {
	fmt.Println("REPLICASET INFO")
	for _, rs := range replicaSetsInfo {
		message = rs.NodeScheduling
		if rs.NodeScheduling != "" {
			if rs.NodeScheduling[:1] == "[" {
				message = rs.NodeScheduling
			} else {
				message = rs.NodeScheduling + " "
			}
		}

		fmt.Printf("Namespace: %s\nName: %s\nReplicas: %d\nNode Scheduling: %s\n",
			rs.Namespace, rs.Name, rs.Replicas, message)
	}
}