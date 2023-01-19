package opclients

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func GetLabelSelectorForDeployment(Name string, Namespace string) string {
	clientSet := GetK8sClient()

	k8sClient := clientSet.AppsV1()

	deployment, _ := k8sClient.Deployments(Namespace).Get(context.Background(), Name, metav1.GetOptions{})

	labelSet := labels.Set(deployment.Spec.Selector.MatchLabels)

	return string(labelSet.AsSelector().String())
}

func GetLabelSelectorForService(Name string, Namespace string) string {
	k8sClient := GetK8sClient().CoreV1()

	listOptions := metav1.GetOptions{}

	service, _ := k8sClient.Services(Namespace).Get(context.Background(), Name, listOptions)

	labelSet := labels.Set(service.Spec.Selector)

	return string(labelSet.AsSelector().String())
}

func LabelPod(pod *v1.Pod, path string, value string) {
	k8sClient := GetK8sClient().CoreV1()
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, updateErr := k8sClient.Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if updateErr == nil {
		logMessage := fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value)
		fmt.Println(logMessage)
	} else {
		fmt.Println(updateErr)
	}
}

func GetPodsForDeployment(Name string, Namespace string) *v1.PodList {
	clientSet := GetK8sClient()

	options := metav1.ListOptions{
		LabelSelector: GetLabelSelectorForDeployment(Name, Namespace),
	}

	podList, err := clientSet.CoreV1().Pods(Namespace).List(context.Background(), options)

	if err != nil {
		fmt.Printf("Get Pods of deployment[%s] error:%v\n", Name, err)
		return nil
	} else {
		for _, pod := range podList.Items {
			logMessage := fmt.Sprintf("Pod found with name %s.", pod.GetName())
			fmt.Println(logMessage)
		}
	}

	return podList
}

func GetPodsForService(Name string, Namespace string) *v1.PodList {
	k8sClient := GetK8sClient().CoreV1()

	options := metav1.ListOptions{
		LabelSelector: GetLabelSelectorForService(Name, Namespace),
	}

	pods, err := k8sClient.Pods(Namespace).List(context.Background(), options)

	if err != nil {
		fmt.Printf("Get Pods of service[%s] error:%v\n", Name, err)
		return nil
	} else {
		for _, pod := range pods.Items {
			logMessage := fmt.Sprintf("Pod found with name %s.", pod.GetName())
			fmt.Println(logMessage)
		}
	}

	return pods
}

func GetK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
