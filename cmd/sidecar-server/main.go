package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/yuweizzz/sidecar"
)

func help() {
	fmt.Println("Usage:")
	fmt.Println("       sidecar-server start [-conf tomlfile] [-dir workdir] [-daemon]")
	fmt.Println("       sidecar-server stop [-dir workdir]")
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
	confPath := startCmd.String("conf", pwd+"/conf.toml", "conf")
	workDir := startCmd.String("dir", pwd, "dir")
	runDaemon := startCmd.Bool("daemon", false, "daemon")
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	stopWorkDir := stopCmd.String("dir", pwd, "dir")

	if len(os.Args) < 2 {
		if os.Getenv("SPECIAL_MARK") == "ENABLED" {
			cfg := readConfig(os.Getenv("CONF_PATH"))
			run(cfg, os.Getenv("WORKDIR"), true)
		} else {
			help()
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "start":
		fmt.Println("subcommand 'start'")
		startCmd.Parse(os.Args[2:])
		cfg := readConfig(*confPath)
		if *runDaemon {
			cmd := &exec.Cmd{
				Path:   "sidecar-server",
				Env:    []string{"SPECIAL_MARK=ENABLED", "CONF_PATH=" + *confPath, "WORKDIR=" + *workDir},
				Stdout: os.Stdout,
				Stderr: os.Stdout,
			}
			err := cmd.Start()
			if err != nil {
				panic(err)
			}
			os.Exit(0)
		} else {
			run(cfg, *workDir, false)
		}
	case "stop":
		stopCmd.Parse(os.Args[2:])
		pid := sidecar.ReadLock(*stopWorkDir + "/sidecar-server.lock")
		// if lock exist
		if pid != 0 {
			syscall.Kill(pid, syscall.SIGINT)
			fmt.Println("Now Server is stopped.")
		} else {
			fmt.Println("Now sidecar-server.lock is not exist, server is stopped")
		}
	default:
		help()
	}
}

//log_fd := os.Stdout
//sidecar.LogRecord(log_fd, "info", "Except signal, exiting......")

func run(cfg *sidecar.Config, workdir string, backgroud bool) {
	daemon := &sidecar.Daemon{
		WorkDir:      workdir,
		CertPath:     workdir + "/sidecar.crt",
		PriKeyPath:   workdir + "/sidecar.pri",
		LockFilePath: workdir + "/sidecar-server.lock",
	}
	daemon.Perpare(backgroud)
	pac := sidecar.NewPac(cfg)
	proxy := sidecar.NewProxyServer(cfg.ProxyPort, daemon.Logger, pac)
	forwarder := sidecar.NewNextProxyServer(proxy.Listener, daemon.Cert, daemon.PriKey, daemon.Logger, cfg.Server, cfg.ComplexPath, cfg.CustomHeaders)
	sidecar.LogRecord(daemon.Logger, "info", "Now Server is running and pid is "+strconv.Itoa(daemon.Pid))
	go proxy.Run()
	go forwarder.Run()
	forwarder.WatchSignal()
	daemon.Clean()
}
