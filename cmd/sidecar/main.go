package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"syscall"

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

	// for run as daemon
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
				cmd := &exec.Cmd{
					Path:   os.Args[0],
					Env:    []string{"SPECIAL_MARK=ENABLED", "CONF_PATH=" + *configClientPath, "TYPE=client"},
					Stdout: os.Stdout,
					Stderr: os.Stdout,
				}
				err := cmd.Start()
				if err != nil {
					panic(err)
				}
				os.Exit(0)
			} else {
				runClient(cfg)
			}
		case "stop":
			lockPath := cfg.Client.WorkDir + "/sidecar.lock"
			pid := sidecar.ReadLock(lockPath)
			// if lock exist
			if pid != 0 {
				proc, _ := os.FindProcess(pid)
				if runtime.GOOS == "windows" {
					sidecar.RemoveLock(lockPath)
					proc.Kill()
				} else {
					proc.Signal(syscall.SIGINT)
				}
				fmt.Println("Now Server is stopped.")
			} else {
				fmt.Println("Now sidecar.lock is not exist, server is stopped")
			}
		default:
			help(filename)
		}
	case "server":
		serverCmd.Parse(os.Args[2:])
		cfg := sidecar.ReadServerConfig(*configServerPath)
		switch *serverAction {
		case "start":
			if cfg.Server.RunAsDaemon {
				cmd := &exec.Cmd{
					Path:   os.Args[0],
					Env:    []string{"SPECIAL_MARK=ENABLED", "CONF_PATH=" + *configServerPath, "TYPE=server"},
					Stdout: os.Stdout,
					Stderr: os.Stdout,
				}
				err := cmd.Start()
				if err != nil {
					panic(err)
				}
				os.Exit(0)
			} else {
				runServer(cfg)
			}
		case "stop":
			lockPath := cfg.Server.WorkDir + "/sidecar.lock"
			pid := sidecar.ReadLock(lockPath)
			if pid != 0 {
				proc, _ := os.FindProcess(pid)
				if runtime.GOOS == "windows" {
					sidecar.RemoveLock(lockPath)
					proc.Kill()
				} else {
					proc.Signal(syscall.SIGINT)
				}
				fmt.Println("Now Server is stopped.")
			} else {
				fmt.Println("Now sidecar.lock is not exist, server is stopped")
			}
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
	daemon.LoadCertAndPriKey()
	pac := sidecar.NewPac(cfg.Client.RemoteServers[0], cfg.Client.GfwListUrl, cfg.Client.CustomProxyHosts)
	proxy := sidecar.NewProxyViaHttps(daemon.Logger, pac, cfg.Client.OnlyListenIPv4, cfg.Client.ProxyPort)
	cache := sidecar.NewCertLRU(daemon.Cert, daemon.PriKey)
	forwarder := sidecar.NewNextProxyServer(proxy.Listener, cache, daemon.Logger,
		cfg.Client.RemoteServers[0].Host, cfg.Client.RemoteServers[0].ComplexPath, cfg.Client.RemoteServers[0].CustomHeaders)
	sidecar.Info("Now Server is run as a Client.")
	sidecar.Info("Now Server is running and pid is " + strconv.Itoa(daemon.Pid))
	go proxy.Run()
	go forwarder.Run()
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
	server := sidecar.NewRemoteServer(cfg.Server.ServerPort, daemon.Logger, daemon.CertPath, daemon.PriKeyPath,
		cfg.Server.OnlyListenIPv4, cfg.Server.ComplexPath, cfg.Server.CustomHeaders)
	sidecar.Info("Now Server is run as a Server.")
	sidecar.Info("Now Server is running and pid is " + strconv.Itoa(daemon.Pid))
	go server.Run()
	daemon.WatchSignal()
}
