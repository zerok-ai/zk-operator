package utils

import (
	"context"
	"fmt"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var LOG_TAG = "k8sutils"

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// If incluster config failes, reading from kubeconfig.
		// However, this is not connecting to gcp clusters. Only working for kind now(probably minikube also).
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetSecretValue(namespace, secretName, dataKey string) (string, error) {
	logger.Debug(LOG_TAG, namespace, secretName, dataKey)
	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting k8s client.")
		return "", err
	}

	secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Failed to get secret: ", err)
		return "", err
	}

	value, ok := secret.Data[dataKey]

	if ok {
		logger.Debug(LOG_TAG, dataKey, value)
		return string(value), nil
	}

	return "", fmt.Errorf("secret Value not found for %v and key %v", secretName, dataKey)
}

// GetNumberOfNodes returns the number of nodes in the cluster.
func GetNumberOfNodes() (int, error) {

	logger.Debug(LOG_TAG, "Getting number of nodes in the cluster.")
	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting k8s client.")
		return -1, err
	}

	// Get the list of nodes
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return -1, err
	}

	// Return the count of nodes
	return len(nodes.Items), nil
}
