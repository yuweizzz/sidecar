package sidecar

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Client struct {
	ProxyPort        int
	OnlyListenIPv4   bool
	RunAsDaemon      bool
	WorkDir          string
	LogLevel         string
	GfwListUrl       string
	CustomProxyHosts []string
	PriKeyPath       string
	CertPath         string
	RemoteServers    []RemoteServerInfo
}

type RemoteServerInfo struct {
	Host          string
	ComplexPath   string
	CustomHeaders map[string]string
}

type Server struct {
	ServerPort     int
	OnlyListenIPv4 bool
	RunAsDaemon    bool
	Mode           string
	WorkDir        string
	LogLevel       string
	PriKeyPath     string
	CertPath       string
	Host           string
	ComplexPath    string
	CustomHeaders  map[string]string
	NginxConf      NginxConf
}

type NginxConf struct {
	EnableListenHTTP2    bool
	EnableWebSocketProxy bool
	EnableModernTLSOnly  bool
	SSLCertificatePath   string
	SSLPrivateKeyPath    string
	NginxWorkDir         string
}

type Config struct {
	Client Client
	Server Server
}

func ReadConfig(path string) (cfg *Config) {
	pwd, _ := os.Getwd()
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}
	if cfg.Client.WorkDir == "" {
		cfg.Client.WorkDir = pwd
	}
	if cfg.Client.PriKeyPath == "" {
		cfg.Client.PriKeyPath = pwd + "/sidecar-client.pri"
	}
	if cfg.Client.CertPath == "" {
		cfg.Client.CertPath = pwd + "/sidecar-client.crt"
	}
	if cfg.Client.RemoteServers == nil {
		panic("You have to config one remote server at least.")
	}
	if cfg.Server.WorkDir == "" {
		cfg.Server.WorkDir = pwd
	}
	if cfg.Server.PriKeyPath == "" {
		panic("As Server, you have to specify the PrivateKey.")
	}
	if cfg.Server.CertPath == "" {
		panic("As Server, you have to specify the Certificate.")
	}
	return
}
