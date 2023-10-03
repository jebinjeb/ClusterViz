package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfigPath string
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Error determining home directory:", err)
		os.Exit(1)
	}
	kubeconfigPath = home + "/.kube/config"

	// Parse command line arguments for kubeconfig file path.
	flag.StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to the kubeconfig file")
	flag.Parse()

	// Load kubeconfig file and create a Kubernetes clientset using clientcmd.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Error loading kubeconfig: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// List Deployments in the cluster.
	deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing Deployments: %v\n", err)
		os.Exit(1)
	}

	// Create a map to store Deployments by namespace.
	deploymentsByNamespace := make(map[string][]string)

	// Iterate through the Deployments and collect them by namespace.
	for _, deployment := range deployments.Items {
		namespace := deployment.Namespace
		deploymentName := deployment.Name

		// Append the Deployment to the namespace in the map.
		deploymentsByNamespace[namespace] = append(deploymentsByNamespace[namespace], deploymentName)
	}
		// Print deployemnts by namespace.
	fmt.Println("Deployments in the cluster by namespace:")
	for namespace, deployList := range deploymentsByNamespace {
		fmt.Printf("\nNamespace: %s\n", namespace)
		for _, deployName := range deployList {
			fmt.Printf("Deployment:  %s\n", deployName)
		}
	}
	// Iterate through the Deployment and check node scheduling.
	fmt.Println("\nDeployment availability and node scheduling:")
	for _, deploy := range deployments.Items {
		fmt.Printf("\nNamespace: %s, Name: %s\n", deploy.Namespace, deploy.Name)

		// Check replica count.
		replicas := *deploy.Spec.Replicas
		info := color.New(color.FgGreen).SprintFunc()
		warning := color.New(color.FgRed).SprintFunc()

		if replicas == 1 {
			fmt.Printf("[%s] Replica count is 1. Consider increasing it for high availability.\n", info("INFO"))
		} else if replicas < 1 {
			fmt.Printf("[%s] Replica count is less than 1. Please update the ReplicaSet.\n", warning("WARNING"))
			continue
		} else {
			fmt.Printf("[%s] Replica count is %d.\n", info("INFO"), replicas)
		}

		// Check node scheduling.
		nodes, err := clientset.CoreV1().Pods(deploy.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + deploy.Name})
		if err != nil {
			fmt.Printf("[%s] Error listing Pods for Deployment %s: %v\n", color.RedString("WARNING"), deploy.Name, err)
			continue
		}

		// Create a map to track the nodes where Pods are scheduled.
		nodeMap := make(map[string]bool)

		// Iterate through Pods and check the nodes and probes.
		for _, pod := range nodes.Items {
			nodeName := pod.Spec.NodeName
			if _, exists := nodeMap[nodeName]; exists {
				fmt.Printf("[%s] Error: Pods are scheduled on the same node.\n", color.RedString("WARNING"))
				break
			}
			nodeMap[nodeName] = true

			// Check liveness probe and readiness probe.
			for _, container := range pod.Spec.Containers {
				livenessProbe := container.LivenessProbe
				readinessProbe := container.ReadinessProbe

				if livenessProbe != nil && livenessProbe.FailureThreshold != 0 {
					fmt.Printf("[%s] Liveness probe for container %s in pod %s has failure threshold: %d.\n", color.RedString("WARNING"), container.Name, pod.Name, livenessProbe.FailureThreshold)
				}

				if readinessProbe != nil {
					if readinessProbe.FailureThreshold != 0 {
						fmt.Printf("[%s] Readiness probe for container %s in pod %s has failure threshold: %d.\n", color.RedString("WARNING"), container.Name, pod.Name, readinessProbe.FailureThreshold)
					} else {
						fmt.Printf("[%s] Readiness probe for container %s in pod %s has passed.\n", color.GreenString("INFO"), container.Name, pod.Name)
					}
				}
			}
		}
	}

	
		}
	

