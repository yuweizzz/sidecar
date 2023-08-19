package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/yuweizzz/sidecar"
)

func help(filename string) {
	fmt.Println("Usage:")
	fmt.Println("      ", filename, "client -action {start|stop} [-conf ./config.toml]")
	fmt.Println("      ", filename, "server -action {start|stop|create-nginx-conf} [-conf ./config.toml]")
}

func main() {
	// basic info
	_, filename := path.Split(os.Args[0])
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// client
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	configClientPath := clientCmd.String("conf", pwd+"/config.toml", "the path of sidecar conf")
	clientAction := clientCmd.String("action", "", "action must in ['start', 'stop']")
	// server
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	configServerPath := serverCmd.String("conf", pwd+"/config.toml", "the path of sidecar conf")
	serverAction := serverCmd.String("action", "", "action must in ['start', 'stop', 'create-nginx-conf']")

	// run as daemon, use function sidecar.StartDaemonProcess() to start
	if len(os.Args) < 2 {
		if os.Getenv("SPECIAL_MARK") == "ENABLED" {
			path := os.Getenv("CONF_PATH")
			switch os.Getenv("TYPE") {
			case "client":
				cfg := sidecar.ReadClientConfig(path)
				runClient(cfg)
			case "server":
				cfg := sidecar.ReadServerConfig(path)
				runServer(cfg)
			}
		} else {
			help(filename)
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "client":
		clientCmd.Parse(os.Args[2:])
		cfg := sidecar.ReadClientConfig(*configClientPath)
		switch *clientAction {
		case "start":
			if cfg.Client.RunAsDaemon {
				sidecar.StartDaemonProcess(*configClientPath, "client")
			} else {
				runClient(cfg)
			}
		case "stop":
			sidecar.StopDaemonProcess(cfg.Client.WorkDir)
		default:
			help(filename)
		}
	case "server":
		serverCmd.Parse(os.Args[2:])
		cfg := sidecar.ReadServerConfig(*configServerPath)
		switch *serverAction {
		case "start":
			if cfg.Server.RunAsDaemon {
				sidecar.StartDaemonProcess(*configServerPath, "server")
			} else {
				runServer(cfg)
			}
		case "stop":
			sidecar.StopDaemonProcess(cfg.Server.WorkDir)
		case "create-nginx-conf":
			sidecar.RenderTemplateByConfig(cfg.Server.WorkDir, cfg)
		default:
			help(filename)
		}
	default:
		help(filename)
	}
}

func runClient(cfg *sidecar.Config) {
	daemon := &sidecar.Daemon{
		WorkDir:      cfg.Client.WorkDir,
		CertPath:     cfg.Client.CertPath,
		PriKeyPath:   cfg.Client.PriKeyPath,
		LockFilePath: cfg.Client.WorkDir + "/sidecar.lock",
		LogLevel:     cfg.Client.LogLevel,
	}
	daemon.Perpare(cfg.Client.RunAsDaemon)
	remoteServer := cfg.Client.RemoteServers[0]
	sidecar.ChangeResolver(cfg.Client.Resolver)
	switch cfg.Client.Mode {
	case "WSS":
		pac := sidecar.NewPac(remoteServer, cfg.Client.GfwListUrl, cfg.Client.CustomProxyHosts)
		proxy := sidecar.NewProxyViaWss(daemon.Logger, pac, cfg.Client.OnlyListenIPv4, cfg.Client.ProxyPort,
			remoteServer.Host, remoteServer.ComplexPath, remoteServer.CustomHeaders)
		go proxy.Run()
	default: // HTTPS
		daemon.LoadCertAndPriKey()
		pac := sidecar.NewPac(remoteServer, cfg.Client.GfwListUrl, cfg.Client.CustomProxyHosts)
		proxy := sidecar.NewProxyViaHttps(daemon.Logger, pac, cfg.Client.OnlyListenIPv4, cfg.Client.ProxyPort)
		cache := sidecar.NewCertLRU(daemon.Cert, daemon.PriKey)
		mitm := sidecar.NewMitMServer(proxy.Listener, cache, daemon.Logger,
			remoteServer.Host, remoteServer.ComplexPath, remoteServer.CustomHeaders)
		go proxy.Run()
		go mitm.Run()
	}
	sidecar.Info("Now Server is run as a Client.")
	daemon.WatchSignal()
}

func runServer(cfg *sidecar.Config) {
	daemon := &sidecar.Daemon{
		WorkDir:      cfg.Server.WorkDir,
		CertPath:     cfg.Server.CertPath,
		PriKeyPath:   cfg.Server.PriKeyPath,
		LockFilePath: cfg.Server.WorkDir + "/sidecar.lock",
		LogLevel:     cfg.Server.LogLevel,
	}
	daemon.Perpare(cfg.Server.RunAsDaemon)
	switch cfg.Server.Mode {
	case "WSS":
		server := sidecar.NewRemoteServerWss(daemon.Logger, cfg.Server.ServerPort, cfg.Server.OnlyListenIPv4,
			daemon.CertPath, daemon.PriKeyPath, cfg.Server.ComplexPath, cfg.Server.CustomHeaders)
		go server.Run()
	default: // HTTPS
		server := sidecar.NewRemoteServerHttps(daemon.Logger, cfg.Server.ServerPort, cfg.Server.OnlyListenIPv4,
			daemon.CertPath, daemon.PriKeyPath, cfg.Server.ComplexPath, cfg.Server.CustomHeaders)
		go server.Run()
	}
	sidecar.Info("Now Server is run as a Server.")
	daemon.WatchSignal()
}
