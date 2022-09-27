package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/yuweizzz/sidecar"
)

func help() {
	fmt.Println("Usage:")
	fmt.Println("       sidecar-server start [-conf tomlfile] [-dir workdir] [-daemon]")
	fmt.Println("       sidecar-server create-nginx-conf [-conf tomlfile] [-dir outputdir]")
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
	tomlPathForStart := startCmd.String("conf", pwd+"/conf.toml", "conf")
	workDirForStart := startCmd.String("dir", pwd, "dir")
	runDaemon := startCmd.Bool("daemon", false, "daemon")
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	workDirForStop := stopCmd.String("dir", pwd, "dir")
	createNginxConfCmd := flag.NewFlagSet("create-nginx-conf", flag.ExitOnError)
	tomlPathForCreate := createNginxConfCmd.String("conf", pwd+"/conf.toml", "conf")
	workDirForCreate := createNginxConfCmd.String("dir", pwd, "dir")

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
		startCmd.Parse(os.Args[2:])
		cfg := readConfig(*tomlPathForStart)
		if *runDaemon {
			cmd := &exec.Cmd{
				Path:   "sidecar-server",
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
		pid := sidecar.ReadLock(*workDirForStop + "/sidecar-server.lock")
		// if lock exist
		if pid != 0 {
			proc, _ := os.FindProcess(pid)
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
