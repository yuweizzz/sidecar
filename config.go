package sidecar

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server            string
	ComplexPath       string
	CustomHeaderName  string
	CustomHeaderValue string
}

func ReadConfig(path string) (cfg *Config) {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return
}
