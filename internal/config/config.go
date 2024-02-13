package config

import "github.com/zerok-ai/zk-utils-go/storage/redis/config"
import logsConfig "github.com/zerok-ai/zk-utils-go/logs/config"

type HttpServerConfig struct {
	Port string `yaml:"port"`
}

type ClusterContextConfig struct {
	Path      string `yaml:"path"`
	CloudAddr string `yaml:"cloudAddr"`
	Port      string `yaml:"port"`
}

type AppConfig struct {
	Redis          config.RedisConfig    `yaml:"redis"`
	Http           HttpServerConfig      `yaml:"http"`
	LogsConfig     logsConfig.LogsConfig `yaml:"logs"`
	ClusterContext ClusterContextConfig  `yaml:"clusterContext"`
}
