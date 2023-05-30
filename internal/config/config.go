package config

type RedisConfig struct {
	Host            string `yaml:"host" env:"REDIS_HOST" env-description:"Database host"`
	Port            string `yaml:"port" env:"REDIS_PORT" env-description:"Database port"`
	ReadTimeout     int    `yaml:"readTimeout"`
	PollingInterval int    `yaml:"pollingInterval"`
	ImageDB         int    `yaml:"image_db"`
	VersionDB       int    `yaml:"version_db"`
}

type OperatorLoginConfig struct {
	Host                string `yaml:"host"`
	Path                string `yaml:"path"`
	ClusterKey          string `yaml:"clusterKeySecret"`
	ClusterKeyData      string `yaml:"clusterKeyData"`
	ClusterKeyNamespace string `yaml:"clusterKeyNamespace"`
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

type ZkCloudConfig struct {
	RestartPath string `yaml:"restartPath"`
	Port        string `yaml:"port"`
}

type InitContainerConfig struct {
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
}

type JavaToolOptionsConfig struct {
	OtelArgument string `yaml:"otelArgument"`
}

type ZkInjectorConfig struct {
	ZkCloud         ZkCloudConfig         `yaml:"zkcloud"`
	Redis           RedisConfig           `yaml:"redis"`
	Webhook         WebhookConfig         `yaml:"webhook"`
	ScenarioSync    ScenarioSyncConfig    `yaml:"scenario_sync"`
	OperatorLogin   OperatorLoginConfig   `yaml:"operator_login"`
	InitContainer   InitContainerConfig   `yaml:"init_container"`
	JavaToolOptions JavaToolOptionsConfig `yaml:"java_tool_options"`
}
