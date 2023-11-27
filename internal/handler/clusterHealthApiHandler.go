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

var ClusterHealthHandlerTag = "ClusterHealthHandler"

var namespace = "zk-client"
var prefix = "zk-"

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Healthy bool
}

// Equals compares this ServiceHealth with another instance
func (sh ServiceHealth) Equals(other ServiceHealth) bool {
	return sh.Healthy == other.Healthy
}

type ClusterHealthRequestPayload struct {
	Services map[string]ServiceHealth `json:"services"`
}

type ClusterHealthHandler struct {
	config            config.ZkOperatorConfig
	latestUpdateTime  time.Time
	serviceHealthData map[string]ServiceHealth
	ticker            *zktick.TickerTask
}

func NewClusterHealthHandler(config config.ZkOperatorConfig) *ClusterHealthHandler {
	ch := &ClusterHealthHandler{
		config:            config,
		serviceHealthData: make(map[string]ServiceHealth),
	}
	var duration = time.Duration(config.ClusterHeathSync.SyncInterval) * time.Second
	ch.ticker = zktick.GetNewTickerTask("ClusterHeathSync", duration, ch.PeriodicSync)
	return ch
}

func (ch *ClusterHealthHandler) StartPeriodicSync() {
	ch.PeriodicSync()
	ch.ticker.Start()
}

// CheckServicesHealth checks the health of services with a specific prefix in a namespace
func (ch *ClusterHealthHandler) CheckServicesHealth(clientset *kubernetes.Clientset, namespace string, prefix string) (map[string]ServiceHealth, error) {
	serviceHealthMap := make(map[string]ServiceHealth)

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
				serviceHealthMap[fmt.Sprintf("%s/%s", namespace, service.Name)] = ServiceHealth{Healthy: false}
				continue
			}
			defer resp.Body.Close()

			// Check if the response status code is 200
			healthy := resp.StatusCode == http.StatusOK
			serviceHealthMap[fmt.Sprintf("%s/%s", namespace, service.Name)] = ServiceHealth{Healthy: healthy}
		}
	}
	return serviceHealthMap, nil
}

// FindServiceHealthDiff finds the difference between two service health maps
func (ch *ClusterHealthHandler) FindServiceHealthDiff(oldMap, newMap map[string]ServiceHealth) map[string]ServiceHealth {
	diff := make(map[string]ServiceHealth)

	for key, newValue := range newMap {
		oldValue, exists := oldMap[key]
		if !exists || !oldValue.Equals(newValue) {
			diff[key] = newValue
		}
	}

	return diff
}

func (ch *ClusterHealthHandler) GetServiceStatusPayload() (*ClusterHealthRequestPayload, error) {
	clientSet, err := utils.GetK8sClient()

	// Get the latest service health data
	serviceHealthData, err := ch.CheckServicesHealth(clientSet, namespace, prefix)
	if err != nil {
		return nil, err
	}

	// Find the difference between the latest service health data and the previous one
	diff := ch.FindServiceHealthDiff(ch.serviceHealthData, serviceHealthData)

	// Update the latest service health data
	ch.serviceHealthData = serviceHealthData

	// Update the latest update time
	ch.latestUpdateTime = time.Now()

	return &ClusterHealthRequestPayload{Services: diff}, nil
}

func (ch *ClusterHealthHandler) PeriodicSync() {
	port := ch.config.ZkCloud.Port
	protocol := "http"
	if port == "443" {
		protocol = "https"
	}

	zklogger.Debug(ClusterHealthHandlerTag, "Starting periodic sync for cluster health.")

	syncUrl := protocol + "://" + ch.config.ZkCloud.Host + ":" + ch.config.ZkCloud.Port + ch.config.ClusterHeathSync.Path

	payload, err := ch.GetServiceStatusPayload()
	if err != nil {
		zklogger.Error(ClusterHealthHandlerTag, "Error getting service status payload:", err)
		return
	}

	client := &http.Client{}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zklogger.Error(ClusterHealthHandlerTag, "Error getting marshalling payload:", err)
		return
	}

	req, err := http.NewRequest("POST", syncUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		zklogger.Error(ClusterHealthHandlerTag, "Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		zklogger.Error(ClusterHealthHandlerTag, "Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Log the response status code
	zklogger.Debug(ClusterHealthHandlerTag, "Response Status Code: %d\n", resp.StatusCode)
}
