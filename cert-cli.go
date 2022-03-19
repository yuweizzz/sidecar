package main

import (
    "log"
    "os"
    "os/exec"
)

func main() {
    cmd := &exec.Cmd{
        Path: "cert",
        Args: os.Args,
    }
    //异步启动子进程
    err := cmd.Start()
    if err != nil {
        log.Println(os.Getpid(), "启动子进程失败:", err)
    } else {
        //执行成功
        log.Println(os.Getpid(), ":", "启动子进程成功:", "->", cmd.Process.Pid, "\n ")
    }

    os.Exit(0)
}