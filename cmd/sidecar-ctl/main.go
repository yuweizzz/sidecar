package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/yuweizzz/sidecar"
)

func main() {
	var action string
	if len(os.Args) != 2 {
		action = ""
	} else {
		action = os.Args[1]
	}
	switch action {
	case "start":
		cmd := &exec.Cmd{
			Path:   "sidecar-server",
			Stdout: os.Stdout,
			Stderr: os.Stdout,
		}
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
	case "stop":
		pid := sidecar.ReadLock()
		// if lock exist
		if pid != 0 {
			syscall.Kill(pid, syscall.SIGINT)
			fmt.Println("Now Server is stopped.")
		} else {
			fmt.Println("Now sidecar-server.lock is not exist, server is stopped")
		}
	case "reload":
		fmt.Println("reload action")
	default:
		fmt.Println("unknow action")
	}
	os.Exit(0)
}
