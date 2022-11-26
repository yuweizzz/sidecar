package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"

	"github.com/yuweizzz/sidecar"
)

func help(filename string) {
	fmt.Println("Usage:")
	fmt.Println("      ", filename, "start [-conf tomlfile] [-dir workdir] [-daemon]")
	fmt.Println("      ", filename, "create-nginx-conf [-conf tomlfile] [-dir outputdir]")
	fmt.Println("      ", filename, "stop [-dir workdir]")
}

func readConfig(path string) *sidecar.Config {
	config_file_path := sidecar.DetectFile(path)
	if config_file_path == "" {
		panic("Run failed, conf file not exist.")
	} else {
		return sidecar.ReadConfig(config_file_path)
	}
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	tomlPathForStart := startCmd.String("conf", pwd+"/conf.toml", "conf")
	workDirForStart := startCmd.String("dir", pwd, "dir")
	runDaemon := startCmd.Bool("daemon", false, "daemon")
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	workDirForStop := stopCmd.String("dir", pwd, "dir")
	createNginxConfCmd := flag.NewFlagSet("create-nginx-conf", flag.ExitOnError)
	tomlPathForCreate := createNginxConfCmd.String("conf", pwd+"/conf.toml", "conf")
	workDirForCreate := createNginxConfCmd.String("dir", pwd, "dir")
	_, filename := path.Split(os.Args[0])

	if len(os.Args) < 2 {
		if os.Getenv("SPECIAL_MARK") == "ENABLED" {
			cfg := readConfig(os.Getenv("CONF_PATH"))
			run(cfg, os.Getenv("WORKDIR"), true)
		} else {
			help(filename)
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
		cfg := readConfig(*tomlPathForStart)
		if *runDaemon {
			cmd := &exec.Cmd{
				Path:   os.Args[0],
				Env:    []string{"SPECIAL_MARK=ENABLED", "CONF_PATH=" + *tomlPathForStart, "WORKDIR=" + *workDirForStart},
				Stdout: os.Stdout,
				Stderr: os.Stdout,
			}
			err := cmd.Start()
			if err != nil {
				panic(err)
			}
			os.Exit(0)
		} else {
			run(cfg, *workDirForStart, false)
		}
	case "stop":
		stopCmd.Parse(os.Args[2:])
		lockPath := *workDirForStop + "/sidecar-server.lock"
		pid := sidecar.ReadLock(lockPath)
		// if lock exist
		if pid != 0 {
			proc, _ := os.FindProcess(pid)
			if runtime.GOOS == "windows" {
				sidecar.RemoveLock(lockPath)
			}
			proc.Kill()
			fmt.Println("Now Server is stopped.")
		} else {
			fmt.Println("Now sidecar-server.lock is not exist, server is stopped")
		}
	case "create-nginx-conf":
		createNginxConfCmd.Parse(os.Args[2:])
		cfg := readConfig(*tomlPathForCreate)
		sidecar.RenderTemplateByConfig(*workDirForCreate, cfg)
	default:
		help(filename)
	}
}

func run(cfg *sidecar.Config, workdir string, backgroud bool) {
	daemon := &sidecar.Daemon{
		WorkDir:      workdir,
		CertPath:     workdir + "/sidecar.crt",
		PriKeyPath:   workdir + "/sidecar.pri",
		LockFilePath: workdir + "/sidecar-server.lock",
		LogLevel:     cfg.Sidecar.LogLevel,
	}
	daemon.Perpare(backgroud)
	pac := sidecar.NewPac(cfg)
	proxy := sidecar.NewProxyServer(cfg.Sidecar.OnlyListenIPv4, cfg.Sidecar.ProxyPort, daemon.Logger, pac)
	cache := sidecar.NewCertLRU(daemon.Cert, daemon.PriKey)
	forwarder := sidecar.NewNextProxyServer(proxy.Listener, cache, daemon.Logger, cfg.RemoteProxy.Server, cfg.RemoteProxy.ComplexPath, cfg.RemoteProxy.CustomHeaders)
	sidecar.Info("Now Server is running and pid is " + strconv.Itoa(daemon.Pid))
	go proxy.Run()
	go forwarder.Run()
	forwarder.WatchSignal()
	daemon.Clean()
}
