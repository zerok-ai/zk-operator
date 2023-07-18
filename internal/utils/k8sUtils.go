package utils

import (
	"context"
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

func getWorkLoadPatchData() string {
	return fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"zk-operator/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
}

func RestartStatefulSet(namespace string, statefulSet string) error {
	logger.Debug(LOG_TAG, "Restarting statefulSet ", statefulSet, " in namespace ", namespace)
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	statefulSetsClient := k8sClient.AppsV1().StatefulSets(namespace)
	_, err = statefulSetsClient.Patch(context.TODO(), statefulSet, types.StrategicMergePatchType, []byte(getWorkLoadPatchData()), metav1.PatchOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while restarting statefulSet ", statefulSet, " with error ", err)
		return err
	}
	return nil
}

func RestartDaemonSet(namespace string, daemonSet string) error {
	logger.Debug(LOG_TAG, "Restarting daemonset ", daemonSet, " in namespace ", namespace)
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	daemonSetsClient := k8sClient.AppsV1().DaemonSets(namespace)
	_, err = daemonSetsClient.Patch(context.TODO(), daemonSet, types.StrategicMergePatchType, []byte(getWorkLoadPatchData()), metav1.PatchOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while restarting daemonSet ", daemonSet, " with error ", err)
		return err
	}
	return nil
}

func RestartDeployment(namespace string, deployment string) error {
	logger.Debug(LOG_TAG, "Restarting deployment ", deployment, " in namespace ", namespace)
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	deploymentsClient := k8sClient.AppsV1().Deployments(namespace)
	_, err = deploymentsClient.Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(getWorkLoadPatchData()), metav1.PatchOptions{})
	if err != nil {
		logger.Error(LOG_TAG, "Error caught while restarting deployment ", deployment, " with error ", err)
		return err
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
		common.ZkInjectionKey: common.ZkInjectionValue,
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

func hasRestartLabel(namespace string, workLoadType int, name, labelKey, labelValue string) (bool, error) {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return false, err
	}

	var objLabels map[string]string

	switch workLoadType {
	case DEPLYOMENT:
		deployment, err := k8sClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		objLabels = deployment.ObjectMeta.Labels

	case STATEFULSET:
		statefulSet, err := k8sClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		objLabels = statefulSet.ObjectMeta.Labels

	case DAEMONSET:
		daemonSet, err := k8sClient.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		objLabels = daemonSet.ObjectMeta.Labels

	default:
		return false, fmt.Errorf("unsupported resource kind")
	}

	value, ok := objLabels[labelKey]
	logger.Debug(LOG_TAG, "Label value is ", value, " and ok is ", ok, " for workload ", name)
	return ok && value == labelValue, nil
}

func RestartMarkedNamespacesIfNeeded() error {
	namespaces, err := GetAllMarkedNamespaces()
	logger.Debug(LOG_TAG, "All marked namespaces are ", namespaces)

	if err != nil || namespaces == nil {
		logger.Error(LOG_TAG, "In restart marked namespaces, error caught while getting all marked namespaces ", err)
		return err
	}

	for _, namespace := range namespaces.Items {

		logger.Debug(LOG_TAG, " Checking for namespace ", namespace.ObjectMeta.Name)

		pods, err := GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			logger.Error(LOG_TAG, "Error caught while getting all non orchestrated pods ", err)
			return err
		}

		logger.Debug(LOG_TAG, " Non orchestrated pods for namespace ", namespace.ObjectMeta.Name, " are ", pods)

		workLoads, err := getWorkloadsForPods(pods)
		if err != nil {
			return err
		}

		logger.Debug(LOG_TAG, " Workloads for pods for namespace ", namespace.ObjectMeta.Name, " are ", workLoads)

		for workLoad := range workLoads {
			restart, err := hasRestartLabel(namespace.ObjectMeta.Name, workLoad.workLoadType, workLoad.name, common.ZkAutoRestartKey, common.ZkAutoRestartValue)
			if err != nil {
				logger.Error(LOG_TAG, "Error caught while checking if workload ", workLoad, " has restart label ", err)
			}
			if !restart {
				logger.Debug(LOG_TAG, "Workload ", workLoad.name, " and type ", workLoad.workLoadType, " does not have restart label, skipping")
				continue
			}
			switch workLoad.workLoadType {
			case DEPLYOMENT:
				err = RestartDeployment(namespace.ObjectMeta.Name, workLoad.name)
			case STATEFULSET:
				err = RestartStatefulSet(namespace.ObjectMeta.Name, workLoad.name)
			case DAEMONSET:
				err = RestartDaemonSet(namespace.ObjectMeta.Name, workLoad.name)
			default:
				logger.Error(LOG_TAG, "Unknown workload type ", workLoad.workLoadType)

			}
			if err != nil {
				logger.Error(LOG_TAG, "Error caught while restarting workload name ", workLoad, " with type ", workLoad.workLoadType, " with error ", err)
				return err
			}
		}
	}
	return nil
}

func getWorkloadsForPods(pods []corev1.Pod) (map[*WorkLoad]bool, error) {
	workLoads := make(map[*WorkLoad]bool)
	for _, pod := range pods {
		workLoad, err := getWorkloadForPod(&pod)
		if err != nil {
			logger.Error(LOG_TAG, "Error caught while getting all deployment for pod ", pod.Name, " with error ", err)
			return workLoads, err
		}
		workLoads[workLoad] = true
	}
	return workLoads, nil
}

func getWorkloadForPod(pod *corev1.Pod) (*WorkLoad, error) {
	ownerReferences := pod.GetOwnerReferences()
	namespace := pod.ObjectMeta.Namespace
	var workLoad *WorkLoad

	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}

	for _, ownerRef := range ownerReferences {
		if ownerRef.Kind == "ReplicaSet" {
			replicaSetName := ownerRef.Name
			replicaSet, err := clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaSetName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			ownerReferences := replicaSet.GetOwnerReferences()
			for _, ownerRef := range ownerReferences {
				if ownerRef.Kind == "Deployment" {
					workLoad = &WorkLoad{
						workLoadType: DEPLYOMENT,
						name:         ownerRef.Name,
					}
					break
				}
				if ownerRef.Kind == "StatefulSet" {
					workLoad = &WorkLoad{
						workLoadType: STATEFULSET,
						name:         ownerRef.Name,
					}
					break
				}
				if ownerRef.Kind == "DaemonSet" {
					workLoad = &WorkLoad{
						workLoadType: DAEMONSET,
						name:         ownerRef.Name,
					}
					break
				}
			}
			break
		}
	}

	return workLoad, nil
}
