package sidecar

import (
	"os"
	"path/filepath"
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
