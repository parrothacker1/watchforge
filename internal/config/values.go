package config

import (
	"sync/atomic"
)

var cfg atomic.Value

func init() {
	cfg.Store((*Config)(nil))
}

func SetConfig(c *Config) {
	cfg.Store(c)
}

func GetConfig() *Config {
	return cfg.Load().(*Config)
}
