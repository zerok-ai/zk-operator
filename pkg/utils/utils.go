package utils

import (
	"encoding/json"
	"github.com/zerok-ai/zk-operator/pkg/common"
	"os"

	corev1 "k8s.io/api/core/v1"
)

func UnmarshalFromString(data string, out interface{}) error {
	err := json.Unmarshal([]byte(data), out)
	if err != nil {
		return err
	}
	return nil
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

func GetCurrentNamespace() string {
	podNamespace := os.Getenv(common.NamespaceEnvVariable)
	return podNamespace
}
