package opclients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	DeploymentInformers map[string]*PodObserver
	ServiceInformers    map[string]*PodObserver
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

func LabelSpillAndSoakPods(podList *v1.PodList) {
	if podList == nil {
		fmt.Printf("Given podList is nil\n")
	} else if len(podList.Items) < 2 {
		fmt.Printf("Not enough pods to apply the configuration.\n")
	} else {
		spillPod := podList.Items[0]
		LabelPod(&spillPod, "/metadata/labels/zk-status", "enabled")
		LabelPod(&spillPod, "/metadata/labels/zk-route-mark", "spill")

		for i := 1; i < len(podList.Items); i++ {
			soakPod := podList.Items[i]
			LabelPod(&soakPod, "/metadata/labels/zk-status", "enabled")
			LabelPod(&soakPod, "/metadata/labels/zk-route-mark", "soak")
		}
	}
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
		fmt.Println(fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value))
	} else {
		fmt.Println(updateErr)
	}
}

func GetMapKey(Name string, Namespace string) string {
	return Namespace + "," + Name
}

func (client *K8sClient) LabelSpillAndSoakPodsForDeployment(Name string, Namespace string) {
	podList := GetPodsForDeployment(Name, Namespace)
	if podList == nil {
		fmt.Printf("Error while fetching podList for deployment %v.\n", Name)
	} else {
		LabelSpillAndSoakPods(podList)
	}
}

func (client *K8sClient) StartObservingPodsForDeployment(Name string, Namespace string) {
	clientSet := GetK8sClient()

	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = GetLabelSelectorForDeployment(Name, Namespace)
	})
	informers := informers.NewSharedInformerFactoryWithOptions(clientSet, 10*time.Second, informers.WithNamespace(Namespace), labelOptions)

	po := &PodObserver{
		informers: informers,
		target:    Deployment,
		Name:      Name,
		Namespace: Namespace,
		client:    client,
		ch:        make(chan struct{}),
	}
	deployKey := GetMapKey(Name, Namespace)
	if client.DeploymentInformers[deployKey] != nil {
		prevPodObserver := client.DeploymentInformers[deployKey]
		prevPodObserver.StopObservingPods()
	}
	client.DeploymentInformers[deployKey] = po
	po.StartObservingPods()
	fmt.Printf("The deploymentInformers map is %v.\n", client.DeploymentInformers)
}

func (client *K8sClient) StartObservingPodsForService(Name string, Namespace string) {
	clientSet := GetK8sClient()

	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = GetLabelSelectorForService(Name, Namespace)
	})

	informers := informers.NewSharedInformerFactoryWithOptions(clientSet, 10*time.Second, informers.WithNamespace(Namespace), labelOptions)

	po := &PodObserver{
		informers: informers,
		target:    Service,
		Name:      Name,
		Namespace: Namespace,
		client:    client,
		ch:        make(chan struct{}),
	}
	serviceKey := GetMapKey(Name, Namespace)
	if client.DeploymentInformers[serviceKey] != nil {
		prevPodObserver := client.ServiceInformers[serviceKey]
		prevPodObserver.StopObservingPods()
	}
	client.ServiceInformers[serviceKey] = po
	po.StartObservingPods()
	fmt.Printf("The serviceInformers map is %v.\n", client.ServiceInformers)
}

func (client *K8sClient) LabelSpillAndSoakPodsForService(Name string, Namespace string) {
	podList := GetPodsForService(Name, Namespace)
	if podList == nil {
		fmt.Printf("Error while fetching podList for service %v.\n", Name)
	} else {
		LabelSpillAndSoakPods(podList)
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
			fmt.Println(fmt.Sprintf("Pod found with name %s.", pod.GetName()))
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
			fmt.Println(fmt.Sprintf("Pod found with name %s.", pod.GetName()))
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

func PrintPodsInCluster() {

	k8sClient := GetK8sClient().CoreV1()

	listOptions := metav1.ListOptions{}

	name := "default"

	services, _ := k8sClient.Services(name).List(context.Background(), listOptions)

	for _, service := range services.Items {
		if name == "default" && service.GetName() == "kubernetes" {
			continue
		}
		fmt.Println("namespace", name, "serviceName:", service.GetName(), "serviceKind:", service.Kind, "serviceLabels:", service.GetLabels(), service.Spec.Ports, "serviceSelector:", service.Spec.Selector)

		// labels.Parser
		set := labels.Set(service.Spec.Selector)

		if pods, err := k8sClient.Pods(name).List(context.Background(), metav1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
			fmt.Printf("List Pods of service[%s] error:%v\n", service.GetName(), err)
		} else {
			for _, pod := range pods.Items {
				fmt.Println("Pod", pod.GetName(), pod.Spec.NodeName, pod.Spec.Containers)
				payload := []patchStringValue{{
					Op:    "replace",
					Path:  "/metadata/labels/testLabel",
					Value: "897889",
				}}
				payloadBytes, _ := json.Marshal(payload)

				_, updateErr := k8sClient.Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
				if updateErr == nil {
					fmt.Println(fmt.Sprintf("Pod %s labelled successfully.", pod.GetName()))
				} else {
					fmt.Println(updateErr)
				}
			}
		}
	}

}
