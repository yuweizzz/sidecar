package sidecar

import (
	"github.com/BurntSushi/toml"
)

type SidecarPart struct {
	ProxyPort        int
	OnlyListenIPv4   bool
	LogLevel         string
	GfwListUrl       string
	CustomProxyHosts []string
}

type RemoteProxyPart struct {
	Server        string
	ComplexPath   string
	CustomHeaders map[string]string
}

type RemoteProxyConfPart struct {
	EnableListenHTTP2    bool
	EnableWebSocketProxy bool
	EnableModernTLSOnly  bool
	SSLCertificatePath   string
	SSLPrivateKeyPath    string
	NginxWorkDir         string
}

type Config struct {
	Sidecar         SidecarPart
	RemoteProxy     RemoteProxyPart
	RemoteProxyConf RemoteProxyConfPart
}

func ReadConfig(path string) (cfg *Config) {
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	return
}
