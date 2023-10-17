package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"log"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var mutex sync.Mutex
type DeploymentInfo struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Replicas       int32  `json:"replicas"`
	NodeScheduling string `json:"node_scheduling"`
}

func GetDeploymentCount() ([]DeploymentInfo, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var deploymentsInfo []DeploymentInfo

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

	// List Deployments in the cluster
	deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing Deployments: %v", err)
	}

	// Iterate through Deployments and gather information
	for _, deploy := range deployments.Items {
		deploymentInfo := DeploymentInfo{
			Namespace: deploy.Namespace,
			Name:      deploy.Name,
			Replicas:  *deploy.Spec.Replicas,
		}

		// Check node scheduling
		nodes, err := clientset.CoreV1().Pods(deploy.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + deploy.Name})
		if err != nil {
			return nil, fmt.Errorf("error listing Pods for Deployment %s: %v", deploy.Name, err)
		}

		info := color.New(color.FgGreen).SprintFunc()
		warning := color.New(color.FgRed).SprintFunc()

		// Check replica count and node scheduling
		if deploymentInfo.Replicas == 1 {
			deploymentInfo.NodeScheduling = warning("WARNING") + " Replica count is 1. Consider increasing it for high availability."
			} else if deploymentInfo.Replicas < 1 {
				deploymentInfo.NodeScheduling = warning("WARNING") + " Replica count is less than 1. Please update the Deployment."
			} else {
				deploymentInfo.NodeScheduling = info("INFO") + fmt.Sprintf(" Replica count is %d.", deploymentInfo.Replicas)
			}
	
		// Only append node scheduling details if there are Pods scheduled
		if len(nodes.Items) > 0 {
			nodeMap := make(map[string]bool)
			var probeMessages []string
	
			// Iterate through Pods and check the nodes and probes
			for _, pod := range nodes.Items {
				nodeName := pod.Spec.NodeName

				nodeMap[nodeName] = true
				for _, container := range pod.Spec.Containers {
					livenessProbe := container.LivenessProbe
					readinessProbe := container.ReadinessProbe
	
					if livenessProbe != nil && livenessProbe.FailureThreshold != 0 {
						probeMessages = append(probeMessages, fmt.Sprintf("Liveness probe failure threshold for container %s in pod %s: %d", container.Name, pod.Name, livenessProbe.FailureThreshold))
					}
	
					if readinessProbe != nil {
						if readinessProbe.FailureThreshold != 0 {
							probeMessages = append(probeMessages, fmt.Sprintf("Readiness probe failure threshold for container %s in pod %s: %d", container.Name, pod.Name, readinessProbe.FailureThreshold))
						} else {
							probeMessages = append(probeMessages, fmt.Sprintf("Readiness probe passed for container %s in pod %s", container.Name, pod.Name))
						}
					}
				}
			}

			// Determine node scheduling status
			if len(nodeMap) > 1 {
				deploymentInfo.NodeScheduling += " " + info("INFO") + " Pods are scheduled on different nodes." + strings.Join(probeMessages, " ")
			} else if len(nodeMap) == 1 {
				deploymentInfo.NodeScheduling += " " + warning("WARNING") + " Pods are scheduled on the same node." + strings.Join(probeMessages, " ")
			} else {
				deploymentInfo.NodeScheduling += " " + warning("WARNING") + " No Pods are scheduled." + strings.Join(probeMessages, " ")
			}
		
			deploymentsInfo = append(deploymentsInfo, deploymentInfo)
		}
		}
		return deploymentsInfo, nil
	}

func GetDeploymentInfoHandler(w http.ResponseWriter, r *http.Request) {
	deploymentsInfo, err := GetDeploymentCount()
		if err != nil {
			log.Printf("Error getting replica set information: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errorMessage := map[string]string{"error": err.Error()}
			json.NewEncoder(w).Encode(errorMessage)
			return
		}
		
	w.Header().Set("Content-Type", "application/json")
		
	json.NewEncoder(w).Encode(deploymentsInfo)
}

func PrintDeploymentInfo(deploymentsInfo []DeploymentInfo) {
	fmt.Println("DEPLOYMENT INFO")
		for _, deployment := range deploymentsInfo {
			fmt.Printf("Namespace: %s\n Name: %s\n Replicas: %d\n Node Scheduling: %s\n",
				deployment.Namespace, deployment.Name, deployment.Replicas, deployment.NodeScheduling)
		}
}