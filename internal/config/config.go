package config

import "github.com/zerok-ai/zk-utils-go/storage/redis/config"
import logsConfig "github.com/zerok-ai/zk-utils-go/logs/config"

type OperatorLoginConfig struct {
	Host                string `yaml:"host"`
	Path                string `yaml:"path"`
	ClusterKey          string `yaml:"clusterKeySecret"`
	ClusterKeyData      string `yaml:"clusterKeyData"`
	ClusterKeyNamespace string `yaml:"clusterKeyNamespace"`
	MaxRetries          int    `yaml:"maxRetries"`
}

type ScenarioSyncConfig struct {
	Host            string `yaml:"host"`
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
	Port               string `yaml:"port"`
	ExceptionPath      string `yaml:"exceptionPath"`
	ClusterContextPath string `yaml:"clusterContextPath"`
}
type ZkCloudConfig struct {
	RestartPath string `yaml:"restartPath"`
	Port        string `yaml:"port"`
}

type InitContainerConfig struct {
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
}

type InstrumentationConfig struct {
	OtelArgument    string `yaml:"otelArgument"`
	PollingInterval int    `yaml:"pollingInterval"`
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
}
