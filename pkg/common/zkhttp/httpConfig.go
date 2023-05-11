package zkhttp

var HTTP_DEBUG bool = false

func Init(http_debug bool) {
	HTTP_DEBUG = http_debug
}

type HttpConfig struct {
	Debug bool `yaml:"debug" env:"HTTP_DEBUG" env-description:"Whether to pass debug information in the response or not"`
}
