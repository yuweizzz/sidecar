package sidecar

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func DetectFile(name string) (path string) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return ""
	}
	path, err = filepath.Abs(name)
	if err != nil {
		panic(err)
	}
	return
}

func CreateDirIfNotExist(path string) (abspath string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0744)
		if err != nil {
			panic(err)
		}
	}
	abspath, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return
}

func CreateFileIfNotExist(file string) (fd *os.File) {
	fd, _ = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_EXCL|os.O_APPEND, 0644)
	return
}

func OpenExistFile(file string) (fd *os.File) {
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return
}

func DetectProcess(pid int) (alive bool) {
	process, err := os.FindProcess(pid)
	if err != nil {
		panic(err)
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			return false
		} else {
			return true
		}
	}
}

func ReadLock() (pid int) {
	lock := DetectFile("sidecar-server.lock")
	if lock == "" {
		return 0
	} else {
		bytes, err := ioutil.ReadFile("sidecar-server.lock")
		if err != nil {
			panic(err)
		}
		pid, err = strconv.Atoi(string(bytes))
		if err != nil {
			panic(err)
		}
		return
	}
}

func WriteLock(pid int) {
	pid_str := strconv.Itoa(pid)
	err := ioutil.WriteFile("sidecar-server.lock", []byte(pid_str), 0444)
	if err != nil {
		panic(err)
	}
}

func RemoveLock() {
	err := os.Remove("sidecar-server.lock")
	if err != nil {
		panic(err)
	}
}
