package apollo

import (
	"time"
)

const (
	defaultCluster        = "default"
	defaultNamespaces     = "application"
	defaultTimeout        = time.Second * 10
	defaultInterval       = time.Second
	defaultNotificationID = -1
)

type Config struct {
	server     string
	appId      string
	cluster    string
	namespaces string
	clientIp   string
	timeout    time.Duration
	interval   time.Duration
}

type ConfigOptions func(opt *Config)

func NewConfig(opts ...ConfigOptions) Config {
	options := Config{
		cluster:    defaultCluster,
		namespaces: defaultNamespaces,
		timeout:    defaultTimeout,
		interval:   defaultInterval,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func WithServer(server string) ConfigOptions {
	return func(opt *Config) {
		opt.server = server
	}
}

func WithAppId(appId string) ConfigOptions {
	return func(opt *Config) {
		opt.appId = appId
	}
}

func WithClientIp(clientIp string) ConfigOptions {
	return func(opt *Config) {
		opt.clientIp = clientIp
	}
}

func WithTimeout(timeout time.Duration) ConfigOptions {
	return func(opt *Config) {
		opt.timeout = timeout
	}
}

func WithInterval(interval time.Duration) ConfigOptions {
	return func(opt *Config) {
		opt.interval = interval
	}
}
