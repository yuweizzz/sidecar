package sidecar

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	ProxyPort            int
	Server               string
	ComplexPath          string
	GfwListUrl           string
	EnableListenHTTP2    bool
	EnableWebSocketProxy bool
	EnableModernTLSOnly  bool
	SSLCertificatePath   string
	SSLPrivateKeyPath    string
	NginxWorkDir         string
	CustomHeaders        map[string]string
}

func ReadConfig(path string) (cfg *Config) {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return
}
