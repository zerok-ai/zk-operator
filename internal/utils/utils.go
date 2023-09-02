package utils

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/common"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sync"
)

var LOG_TAG_UTILS = "utils"

func GetContainerRuntime(data string) (*common.ContainerRuntime, error) {
	var runtimeDetails common.ContainerRuntime
	err := json.Unmarshal([]byte(data), &runtimeDetails)
	if err != nil {
		return nil, err
	}
	return &runtimeDetails, nil
}

func SyncMapToString(m *sync.Map) (string, error) {
	resultMap := make(map[string]interface{})

	m.Range(func(key, value interface{}) bool {
		resultMap[fmt.Sprintf("%v", key)] = value
		return true
	})

	logger.Debug(LOG_TAG_UTILS, resultMap)

	mapBytes, err := json.Marshal(resultMap)
	if err != nil {
		return "", err
	}

	return string(mapBytes), nil
}

func StringToSyncMap(str string) (*sync.Map, error) {
	var resultMap map[string]interface{}

	err := json.Unmarshal([]byte(str), &resultMap)
	if err != nil {
		return nil, err
	}

	newMap := &sync.Map{}
	for key, value := range resultMap {
		newMap.Store(key, value)
	}

	return newMap, nil
}
func GetIndexOfEnv(envVars []corev1.EnvVar, targetEnv string) int {
	for index, envVar := range envVars {
		if envVar.Name == targetEnv {
			return index
		}
	}
	return -1
}

func GetCurrentNamespace() string {
	podNamespace := os.Getenv(common.NamespaceEnvVariable)
	return podNamespace
}

func RespCodeIsOk(status int) bool {
	if status > 199 && status < 300 {
		return true
	}
	return false

}
