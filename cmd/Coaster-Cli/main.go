package main

import (
	"os"
    "io/ioutil"
    "os/exec"
	"fmt"
    "strconv"
    "syscall"

	//"github.com/yuweizzz/Coaster"
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
                Path: "Coaster",
                Args: os.Args,
            }
            err := cmd.Start()
            if err != nil {
                panicWapper(err)
            } else {
                pid := strconv.Itoa(cmd.Process.Pid)
                err = ioutil.WriteFile("Coaster.lock", []byte(pid), 0644)
                if err != nil {
                    panicWapper(err)
                }
                fmt.Println("Now Server is running, pid:", pid)
            }
        case "stop":
            bytes, err := ioutil.ReadFile("Coaster.lock")
            if err != nil {
                panicWapper(err)
            }
            pid, err := strconv.Atoi(string(bytes))
            if err != nil {
                panicWapper(err)
            }
            syscall.Kill(pid, syscall.SIGINT)
            err = os.Remove("Coaster.lock")
            if err != nil {
                panicWapper(err)
            }
            fmt.Println("Now Server is stopped.")
        case "reload":
            fmt.Println("reload action")
        default:
            fmt.Println("unknow action")
    }
    os.Exit(0)
}

func panicWapper(err error) {
    fmt.Println("Coaster-Cli Run Failed!")
    panic(err)
}
