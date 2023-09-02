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

func CreateProcessMap(m *sync.Map) (string, error) {
	resultMap := make(map[string]interface{})

	m.Range(func(key, value interface{}) bool {
		switch y := value.(type) {
		case *common.ContainerRuntime:
			if len(y.Languages) > 0 {
				lang := y.Languages[0]
				if lang == "java" {
					temp := make(map[string]string)
					temp["Process"] = y.Process
					temp["language"] = "java"
					if len(y.Cmd) > 0 {
						temp["CmdLine"] = y.Cmd[0]
					}
					temp["JAVA_TOOL_OPTIONS"] = y.EnvMap["JAVA_TOOL_OPTIONS"]
					resultMap[fmt.Sprintf("%v", key)] = temp
				}
			}
		}
		return true
	})

	logger.Debug(LOG_TAG_UTILS, resultMap)

	mapBytes, err := json.MarshalIndent(resultMap, "", " ")
	if err != nil {
		return "", err
	}

	return string(mapBytes), nil
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
	if len(envVars) == 0 {
		return -1
	}
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
