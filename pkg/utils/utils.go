package utils

import (
	"encoding/json"

	"github.com/zerok-ai/zk-operator/pkg/common"

	corev1 "k8s.io/api/core/v1"
)

func GetContainerRuntime(data string) (*common.ContainerRuntime, error) {
	var runtimeDetails common.ContainerRuntime
	err := json.Unmarshal([]byte(data), &runtimeDetails)
	if err != nil {
		return nil, err
	}
	return &runtimeDetails, nil
}

func ToJsonString(iInstance interface{}) *string {
	if iInstance == nil {
		return nil
	}
	bytes, err := json.Marshal(iInstance)
	if err != nil {
		return nil
	}
	iString := string(bytes)
	return &iString
}

func GetIndexOfEnv(envVars []corev1.EnvVar, targetEnv string) int {
	for index, envVar := range envVars {
		if envVar.Name == targetEnv {
			return index
		}
	}
	return -1
}
