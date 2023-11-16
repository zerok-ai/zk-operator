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

type ConfigSyncConfig struct {
	CloudPath       string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	ApiPath         string `yaml:"apiPath"`
}

type ScenarioSyncConfig struct {
	Path            string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	DB              int    `yaml:"db"`
}

type ObfuscationSyncConfig struct {
	Path            string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	DB              int    `yaml:"db"`
}

type IntegrationSyncConfig struct {
	Path            string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	DB              int    `yaml:"db"`
}

type HttpServerConfig struct {
	Port          string `yaml:"port"`
	ExceptionPath string `yaml:"exceptionPath"`
}
type ZkCloudConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
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

type ExecutorAttributesSyncConfig struct {
	Path            string `yaml:"path"`
	PollingInterval int    `yaml:"pollingInterval"`
	DB              int    `yaml:"db"`
}

type ZkOperatorConfig struct {
	ZkCloud                ZkCloudConfig                `yaml:"zkcloud"`
	Redis                  config.RedisConfig           `yaml:"redis"`
	Http                   HttpServerConfig             `yaml:"http"`
	ScenarioSync           ScenarioSyncConfig           `yaml:"scenarioSync"`
	IntegrationSync        IntegrationSyncConfig        `yaml:"integrationSync"`
	ConfigurationSync      ConfigSyncConfig             `yaml:"configurationSync"`
	ObfuscationSync        ObfuscationSyncConfig        `yaml:"obfuscationRulesSync"`
	ExecutorAttributesSync ExecutorAttributesSyncConfig `yaml:"executorAttributesSync"`
	OperatorLogin          OperatorLoginConfig          `yaml:"operatorLogin"`
	LogsConfig             logsConfig.LogsConfig        `yaml:"logs"`
	ClusterContext         ClusterContextConfig         `yaml:"clusterContext"`
	WspClient              WspClientConfig              `yaml:"wspClient"`
}
