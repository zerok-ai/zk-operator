package config

type RedisConfig struct {
	Host            string `yaml:"host" env:"REDIS_HOST" env-description:"Database host"`
	Port            string `yaml:"port" env:"REDIS_PORT" env-description:"Database port"`
	ReadTimeout     int    `yaml:"readTimeout"`
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

type ZkInjectorConfig struct {
	Redis   RedisConfig   `yaml:"redis"`
	Webhook WebhookConfig `yaml:"webhook"`
	Local   bool          `yaml:"local"`
	ZkCloud ZkCloudConfig `yaml:"zkcloud"`
}
