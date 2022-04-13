package Coaster

import (
	"github.com/BurntSushi/toml"
)

type config struct {
	Server            string
	ComplexPath       string
	CustomHeaderName  string
	CustomHeaderValue string
}

func ReadConfig(path string) (cfg *config) {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return
}
