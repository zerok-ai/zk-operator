package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/common"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	"sync"
	"time"

	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

var LOG_TAG = "k8sutils"

func getPodsWithSelector(selector string, namespace string) (*corev1.PodList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	return pods, nil
}

func GetPodsWithLabel(labelKey, labelValue, namespace string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector(labelKey+"="+labelValue, namespace)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func GetPodsWithoutLabel(labelKey string, namespace string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector("!"+labelKey, namespace)
	if err != nil {
		logger.Error(LOG_TAG, "Error while getting pods without label ", err)
		return nil, err
	}
	return pods, nil
}

func RestartDeployment(namespace string, deployment string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	deploymentsClient := k8sClient.AppsV1().Deployments(namespace)
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"zk-operator/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	_, err = deploymentsClient.Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while restarting deployment ", err)
		return err
	}
	return nil
}

func RestartAllDeploymentsInNamespace(namespace string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	deployments, err := k8sClient.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Error getting deployments: ", err)
		return err
	}

	for _, deployment := range deployments.Items {
		logger.Error(LOG_TAG, "Restarting Deployment: ", deployment.ObjectMeta.Name)
		RestartDeployment(namespace, deployment.ObjectMeta.Name)
	}
	return nil
}

func GetAllMarkedNamespaces() (*corev1.NamespaceList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting client.")
		return nil, err
	}

	labelSelector := labels.Set{
		common.ZkInjectionKey:   common.ZkInjectionValue,
		common.ZkAutoRestartKey: "true",
	}.AsSelector()

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), listOptions)
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while getting list of namespaces ", err)
		return nil, err
	}

	return namespaces, nil
}

func GetNotOrchestratedPods(namespace string) ([]corev1.Pod, error) {
	podList := []corev1.Pod{}

	//Getting pods which does not have zk-status label.
	pods, err := GetPodsWithoutLabel(common.ZkOrchKey, namespace)
	if err != nil {
		return podList, err
	}
	podList = append(podList, pods.Items...)

	//Getting pods where zk-status is in process.
	pods, err = GetPodsWithLabel(common.ZkOrchKey, common.ZkOrchInProcess, namespace)
	if err != nil {
		return podList, err
	}
	podList = append(podList, pods.Items...)

	return podList, err
}

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

func CreateOrUpdateConfigMap(namespace, name string, imageMap *sync.Map) error {

	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting k8s client.")
		return err
	}

	configMaps := clientSet.CoreV1().ConfigMaps(namespace)

	data := make(map[string]string)
	logger.Debug(LOG_TAG, imageMap)
	jsonString, err := SyncMapToString(imageMap)
	if err != nil {
		logger.Error(LOG_TAG, "Error while converting scenario.Map to string ", err)
	}

	logger.Debug(LOG_TAG, "The json string is ", jsonString)
	data[common.ZkConfigMapKey] = jsonString

	configMap, err := configMaps.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		// If the ConfigMap doesn't exist, create it
		if errors.IsNotFound(err) {
			newConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Data: data,
			}
			_, err = configMaps.Create(context.TODO(), newConfigMap, metav1.CreateOptions{})
			return err
		}
		return err
	}

	configMap.Data = data
	_, err = configMaps.Update(context.TODO(), configMap, metav1.UpdateOptions{})
	return err
}

func GetDataFromConfigMap(namespace, name string) (*sync.Map, error) {
	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting k8s client.")
		return nil, err
	}

	configMaps := clientSet.CoreV1().ConfigMaps(namespace)

	configMap, err := configMaps.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	imageMap, err := StringToSyncMap(configMap.Data[common.ZkConfigMapKey])
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while unmarshalling the data from configmap ", err)
		return nil, err
	}

	return imageMap, nil
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

func DeleteNamespaceWithRetry(namespaceName string, maxRetries int, retryDelay time.Duration) error {
	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Error(LOG_TAG, " Error while getting k8s client.")
		return err
	}

	for i := 0; i < maxRetries; i++ {
		err := clientSet.CoreV1().Namespaces().Delete(context.TODO(), namespaceName, metav1.DeleteOptions{})
		if err == nil {
			return nil
		}

		logger.Error(LOG_TAG, "Failed to delete namespace ", namespaceName, " retrying in ", retryDelay)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("failed to delete namespace %s after %d retries", namespaceName, maxRetries)
}

func RestartMarkedNamespacesIfNeeded() error {
	namespaces, err := GetAllMarkedNamespaces()

	if err != nil || namespaces == nil {
		logger.Error(LOG_TAG, "In restart marked namespaces, error caught while getting all marked namespaces ", err)
		return err
	}

	for _, namespace := range namespaces.Items {

		pods, err := GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			logger.Error(LOG_TAG, "Error caught while getting all non orchestrated pods ", err)
			return err
		}

		deployments, err := getDeploymentsForPods(pods)
		if err != nil {
			return err
		}

		for deploymentName := range deployments {
			err = RestartDeployment(namespace.ObjectMeta.Name, deploymentName)
			if err != nil {
				logger.Error(LOG_TAG, "Error caught while restaring deployment name ", deploymentName, " with error ", err)
				return err
			}
		}
	}
	return nil
}

func getDeploymentsForPods(pods []corev1.Pod) (map[string]bool, error) {
	deployments := make(map[string]bool)
	for _, pod := range pods {
		deploymentName, err := getDeploymentForAPod(&pod)
		if err != nil {
			logger.Error(LOG_TAG, "Error caught while getting all deployment for pod ", pod.Name, " with error ", err)
			return deployments, err
		}
		deployments[deploymentName] = true
	}
	return deployments, nil
}

func getDeploymentForAPod(pod *corev1.Pod) (string, error) {
	ownerReferences := pod.GetOwnerReferences()
	namespace := pod.ObjectMeta.Namespace
	var deploymentName string

	clientset, err := GetK8sClient()
	if err != nil {
		return "", err
	}

	for _, ownerRef := range ownerReferences {
		if ownerRef.Kind == "ReplicaSet" {
			replicaSetName := ownerRef.Name
			replicaSet, err := clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaSetName, metav1.GetOptions{})
			if err != nil {
				return "", err
			}

			ownerReferences := replicaSet.GetOwnerReferences()
			for _, ownerRef := range ownerReferences {
				if ownerRef.Kind == "Deployment" {
					deploymentName = ownerRef.Name
					break
				}
			}
			break
		}
	}

	return deploymentName, nil
}

func restartWorkloads(restartRequestObj common.RestartRequest) error {
	namespace := restartRequestObj.Namespace
	all := restartRequestObj.All
	if all {
		return RestartAllDeploymentsInNamespace(namespace)
	} else {
		deployment := restartRequestObj.Deployment
		return RestartDeployment(namespace, deployment)
	}
}

func LabelPod(pod *corev1.Pod, path string, value string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, updateErr := k8sClient.CoreV1().Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if updateErr == nil {
		logMessage := fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value)
		logger.Debug(LOG_TAG, logMessage)
		return updateErr
	} else {
		logger.Error(LOG_TAG, updateErr)
	}
	return nil
}

func GetPodsMatchingLabels(labelsMap map[string]string, namespace string) (*corev1.PodList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	labelSet := labels.Set(labelsMap)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSet.AsSelector().String(),
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	return pods, err
}

func GetAllNonOrchestratedPods() ([]corev1.Pod, error) {
	allPodsList := []corev1.Pod{}
	namespaces, err := GetAllMarkedNamespaces()

	if err != nil {
		logger.Error(LOG_TAG, "Error caught while getting list of namespaces ", err)
		return nil, err
	}

	for _, namespace := range namespaces.Items {
		logger.Debug(LOG_TAG, "Checking for namespace %v.\n", namespace)
		pods, err := GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			err = fmt.Errorf("error getting non orchestrated pods from namespace %v", namespace)
			return nil, err
		}
		allPodsList = append(allPodsList, pods...)
	}
	return allPodsList, nil
}