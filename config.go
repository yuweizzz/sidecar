package sidecar

import (
	"net"
	"os"

	"github.com/BurntSushi/toml"
)

type CustomHeaders map[string]string

type Client struct {
	ProxyPort        int
	OnlyListenIPv4   bool
	RunAsDaemon      bool
	Mode             string
	WorkDir          string
	LogLevel         string
	GfwListUrl       string
	CustomProxyHosts []string
	PriKeyPath       string
	CertPath         string
	Resolver         string
	RemoteServers    []RemoteServerInfo
}

type RemoteServerInfo struct {
	Host          string
	ComplexPath   string
	CustomHeaders CustomHeaders
}

type Server struct {
	ServerPort     int
	OnlyListenIPv4 bool
	RunAsDaemon    bool
	Mode           string
	WorkDir        string
	Resolver       string
	LogLevel       string
	PriKeyPath     string
	CertPath       string
	ComplexPath    string
	CustomHeaders  CustomHeaders
}

type NginxConfig struct {
	ServerName           string
	ServerPort           int
	OnlyListenIPv4       bool
	Location             string
	Resolver             string
	SSLCertificate       string
	SSLPrivateKey        string
	WorkDir              string
	EnableListenHTTP2    bool
	EnableWebSocketProxy bool
	EnableModernTLSOnly  bool
	NginxCustomHeader    CustomHeaders
}

type Config struct {
	Client      Client
	Server      Server
	NginxConfig NginxConfig
}

func ReadClientConfig(path string) (cfg *Config) {
	configPath := DetectFile(path)
	if configPath == "" {
		panic("Run failed, config.toml not exist.")
	}
	pwd, _ := os.Getwd()
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
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
	return
}

func ReadServerConfig(path string) (cfg *Config) {
	configPath := DetectFile(path)
	if configPath == "" {
		panic("Run failed, config.toml not exist.")
	}
	pwd, _ := os.Getwd()
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
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

func ReadNginxTplConfig(path string) (cfg *Config) {
	configPath := DetectFile(path)
	if configPath == "" {
		panic("Run failed, config.toml not exist.")
	}
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	if cfg.NginxConfig.ServerName == "" {
		panic("As Server, you have to specify the Nginx ServerName.")
	}
	valid := net.ParseIP(cfg.NginxConfig.Resolver)
	if valid == nil {
		cfg.NginxConfig.Resolver = ""
	}
	if cfg.NginxConfig.Location == "" {
		panic("As Server, you have to specify the Nginx Location.")
	}
	if cfg.NginxConfig.SSLCertificate == "" {
		panic("As Server, you have to specify the Nginx SSL Certificate.")
	}
	if cfg.NginxConfig.SSLPrivateKey == "" {
		panic("As Server, you have to specify the Nginx SSL PrivateKey.")
	}
	return
}
