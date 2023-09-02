package config

import (
	"context"
	"fmt"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

var LOG_TAG = "AppInitContainerData"

type AppInitContainerData struct {
	Image  string
	Tag    string
	ticker *zktick.TickerTask
}

func (h *AppInitContainerData) Init(cfg ZkOperatorConfig) {
	//Creating a timer for periodic app init container sync.
	interval := cfg.InitContainer.PollingInterval
	var duration = 2 * time.Minute
	if interval > 0 {
		duration = time.Duration(cfg.ScenarioSync.PollingInterval) * time.Second
	}
	h.ticker = zktick.GetNewTickerTask("scenario_sync", duration, h.periodicSync)
}

func (h *AppInitContainerData) StartPeriodicSync() {
	h.updateInitContainerConfig()
	h.ticker.Start()
}

func (h *AppInitContainerData) periodicSync() {
	logger.Debug(LOG_TAG, "Init container data sync triggered.")
	h.updateInitContainerConfig()
}

func (h *AppInitContainerData) updateInitContainerConfig() {
	Data, err := GetDataFromConfigMap("zk-client", "zk-app-init-container")
	if err != nil {
		logger.Debug(LOG_TAG, "Error while reading data from app init container config map.")
		return
	}
	image, ok1 := Data["appInitContainerRepo"]
	tag, ok2 := Data["appInitContainerTag"]
	if ok1 && ok2 {
		h.Image = image
		h.Tag = tag
		logger.Debug(LOG_TAG, "Found image ", image, " with tag ", tag)
	}
}

func GetDataFromConfigMap(namespace, name string) (map[string]string, error) {
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

	return configMap.Data, nil
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
