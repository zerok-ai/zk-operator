package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"strings"
	"time"
)

var ClusterStatusHandlerTag = "ClusterStatusHandler"

var namespace = "zk-client"
var prefix = "zk-"

// ServiceStatus represents the health status of a service
type ServiceStatus struct {
	Healthy bool
}

// Equals compares this ServiceStatus with another instance
func (sh ServiceStatus) Equals(other ServiceStatus) bool {
	return sh.Healthy == other.Healthy
}

type ClusterStatusRequestPayload struct {
	NumberOfNodes int                      `json:"number_of_nodes"`
	Services      map[string]ServiceStatus `json:"services"`
}

type ClusterStatusHandler struct {
	config            config.ZkOperatorConfig
	latestUpdateTime  time.Time
	serviceStatusData map[string]ServiceStatus
	ticker            *zktick.TickerTask
}

func NewClusterStatusHandler(config config.ZkOperatorConfig) *ClusterStatusHandler {
	ch := &ClusterStatusHandler{
		config:            config,
		serviceStatusData: make(map[string]ServiceStatus),
	}
	var duration = time.Duration(config.ClusterHeathSync.SyncInterval) * time.Second
	ch.ticker = zktick.GetNewTickerTask("ClusterHeathSync", duration, ch.PeriodicSync)
	return ch
}

func (ch *ClusterStatusHandler) StartPeriodicSync() {
	ch.PeriodicSync()
	ch.ticker.Start()
}

// CheckServicesStatus checks the status of services with a specific prefix in a namespace
func (ch *ClusterStatusHandler) CheckServicesStatus(clientset *kubernetes.Clientset, namespace string, prefix string) (map[string]ServiceStatus, error) {
	serviceStatusMap := make(map[string]ServiceStatus)

	// Get the list of services in the specified namespace
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Iterate over services and check if they match the prefix
	for _, service := range services.Items {
		if strings.HasPrefix(service.Name, prefix) {
			// Construct the URL for the /healthz endpoint of the service
			url := fmt.Sprintf("http://%s.%s.svc.cluster.local/healthz", service.Name, namespace)

			// Create an HTTP client with a timeout
			client := http.Client{
				Timeout: 5 * time.Second,
			}

			// Make the HTTP GET request
			resp, err := client.Get(url)
			if err != nil {
				serviceStatusMap[fmt.Sprintf("%s/%s", namespace, service.Name)] = ServiceStatus{Healthy: false}
				continue
			}
			defer resp.Body.Close()

			// Check if the response status code is 200
			healthy := resp.StatusCode == http.StatusOK
			serviceStatusMap[fmt.Sprintf("%s/%s", namespace, service.Name)] = ServiceStatus{Healthy: healthy}
		}
	}
	return serviceStatusMap, nil
}

// FindServiceStatusDiff finds the difference between two service status maps
func (ch *ClusterStatusHandler) FindServiceStatusDiff(oldMap, newMap map[string]ServiceStatus) map[string]ServiceStatus {
	diff := make(map[string]ServiceStatus)

	for key, newValue := range newMap {
		oldValue, exists := oldMap[key]
		if !exists || !oldValue.Equals(newValue) {
			diff[key] = newValue
		}
	}

	return diff
}

func (ch *ClusterStatusHandler) GetServiceStatusPayload() (*ClusterStatusRequestPayload, error) {
	clientSet, err := utils.GetK8sClient()

	// Get the latest service status data
	serviceStatusData, err := ch.CheckServicesStatus(clientSet, namespace, prefix)
	if err != nil {
		return nil, err
	}

	// Find the difference between the latest service status data and the previous one
	diff := ch.FindServiceStatusDiff(ch.serviceStatusData, serviceStatusData)

	// Update the latest service status data
	ch.serviceStatusData = serviceStatusData

	// Update the latest update time
	ch.latestUpdateTime = time.Now()

	//This method will return number of nodes as -1 in case of an error.
	numberOfNodes, err := utils.GetNumberOfNodes()
	if err != nil {
		zklogger.Error(ClusterStatusHandlerTag, "Error getting number of nodes:", err)
	}

	return &ClusterStatusRequestPayload{Services: diff, NumberOfNodes: numberOfNodes}, nil
}

func (ch *ClusterStatusHandler) PeriodicSync() {
	port := ch.config.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	zklogger.Debug(ClusterStatusHandlerTag, "Starting periodic sync for cluster status.")

	syncUrl := protocol + "://" + ch.config.ZkCloud.Host + ":" + ch.config.ZkCloud.Port + ch.config.ClusterHeathSync.Path

	payload, err := ch.GetServiceStatusPayload()
	if err != nil {
		zklogger.Error(ClusterStatusHandlerTag, "Error getting service status payload:", err)
		return
	}

	client := &http.Client{}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zklogger.Error(ClusterStatusHandlerTag, "Error getting marshalling payload:", err)
		return
	}

	req, err := http.NewRequest("POST", syncUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		zklogger.Error(ClusterStatusHandlerTag, "Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		zklogger.Error(ClusterStatusHandlerTag, "Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Log the response status code
	zklogger.Debug(ClusterStatusHandlerTag, "Response Status Code: %d\n", resp.StatusCode)
}
