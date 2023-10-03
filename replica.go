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

func replica() {

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

	// List ReplicaSets in the cluster.
	replicaSets, err := clientset.AppsV1().ReplicaSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing ReplicaSets: %v\n", err)
		os.Exit(1)
	}
	// Create a map to store ReplicaSets by namespace.
	replicaSetsByNamespace := make(map[string][]string)

	// Iterate through the ReplicaSets and collect them by namespace.
	for _, rs := range replicaSets.Items {
		namespace := rs.Namespace
		replicaSetName := rs.Name

		// Append the ReplicaSet to the namespace in the map.
		replicaSetsByNamespace[namespace] = append(replicaSetsByNamespace[namespace], replicaSetName)
	}
	// Print ReplicaSets by namespace.
	fmt.Println("ReplicaSets in the cluster by namespace:")
	for namespace, rsList := range replicaSetsByNamespace {
		fmt.Printf("Namespace: %s\n", namespace)
		for _, rsName := range rsList {
			fmt.Printf("Replicaset:  %s\n\n", rsName)
		}
	}

	// Iterate through the ReplicaSets and check node scheduling.
	fmt.Println("ReplicaSets count and availability:")
	for _, rs := range replicaSets.Items {
		fmt.Printf("\nNamespace: %s, Name: %s\n", rs.Namespace, rs.Name)

		// Check replica count.
		replicas := *rs.Spec.Replicas
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
		nodes, err := clientset.CoreV1().Pods(rs.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + rs.Name})
		if err != nil {
			fmt.Printf("[%s] Error listing Pods for ReplicaSet %s: %v\n\n", warning("WARNING"), rs.Name, err)
			continue
		}
		
		// Create a map to track the nodes where Pods are scheduled.
		nodeMap := make(map[string]bool)

		// Iterate through Pods and check the nodes.
		for _, pod := range nodes.Items {
			nodeName := pod.Spec.NodeName
			if _, exists := nodeMap[nodeName]; exists {
				fmt.Printf("[%s] Error: Pods are scheduled on the same node.\n", warning("WARNING"))
				break
			}
			nodeMap[nodeName] = true
		}
		if len(nodeMap) > 1 {
			fmt.Printf("[%s] Pods are scheduled on different nodes.\n", info("INFO"))
		} else {
			fmt.Printf("[%s] Pods are scheduled on the same node.\n", warning("WARNING"))
		}
	}
}
