package config

import "github.com/zerok-ai/zk-utils-go/storage/redis/config"

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

type ExceptionConfig struct {
	Path string `yaml:"path"`
	Port string `yaml:"port"`
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
	Exception       ExceptionConfig       `yaml:"exception"`
	ScenarioSync    ScenarioSyncConfig    `yaml:"scenario_sync"`
	OperatorLogin   OperatorLoginConfig   `yaml:"operator_login"`
	InitContainer   InitContainerConfig   `yaml:"init_container"`
	Instrumentation InstrumentationConfig `yaml:"instrumentation"`
	LogLevel        string                `yaml:"logLevel"`
}
