package sidecar

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	ProxyPort     int
	Server        string
	ComplexPath   string
	CustomHeaders map[string]string
}

func ReadConfig(path string) (cfg *Config) {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return
}
