package config

import "github.com/zerok-ai/zk-utils-go/storage/redis/config"
import logsConfig "github.com/zerok-ai/zk-utils-go/logs/config"

type OperatorLoginConfig struct {
	Path                string `yaml:"path"`
	ClusterSecretName   string `yaml:"clusterSecretName"`
	ClusterKeyData      string `yaml:"clusterKeyData"`
	ApiKeyData          string `yaml:"apiKeyData"`
	ClusterKeyNamespace string `yaml:"clusterKeyNamespace"`
	MaxRetries          int    `yaml:"maxRetries"`
}

type ScenarioSyncConfig struct {
	Path            string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	DB              int    `yaml:"db"`
}

type WebhookConfig struct {
	Namespace string `yaml:"namespace"`
	Service   string `yaml:"service"`
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
	Port      string `yaml:"port"`
}

type HttpServerConfig struct {
	Port          string `yaml:"port"`
	ExceptionPath string `yaml:"exceptionPath"`
}
type ZkCloudConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type InitContainerConfig struct {
	PollingInterval int `yaml:"pollingInterval"`
}

type InstrumentationConfig struct {
	OtelArgument    string `yaml:"otelArgument"`
	PollingInterval int    `yaml:"pollingInterval"`
}

type ClusterContextConfig struct {
	Path      string `yaml:"path"`
	CloudAddr string `yaml:"cloudAddr"`
	Port      string `yaml:"port"`
}

type WspClientConfig struct {
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	Path              string `yaml:"path"`
	DestinationHeader string `yaml:"destinationHeader"`
}

type ZkOperatorConfig struct {
	ZkCloud         ZkCloudConfig         `yaml:"zkcloud"`
	Redis           config.RedisConfig    `yaml:"redis"`
	Webhook         WebhookConfig         `yaml:"webhook"`
	Http            HttpServerConfig      `yaml:"http"`
	ScenarioSync    ScenarioSyncConfig    `yaml:"scenarioSync"`
	OperatorLogin   OperatorLoginConfig   `yaml:"operatorLogin"`
	InitContainer   InitContainerConfig   `yaml:"initContainer"`
	Instrumentation InstrumentationConfig `yaml:"instrumentation"`
	LogsConfig      logsConfig.LogsConfig `yaml:"logs"`
	ClusterContext  ClusterContextConfig  `yaml:"clusterContext"`
	WspClient       WspClientConfig       `yaml:"wspClient"`
}
