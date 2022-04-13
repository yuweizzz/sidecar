package Coaster

import (
	"os"
	"path/filepath"
)

func DetectConfigFile() (path string) {
	_, err := os.Stat("conf.toml")
	if os.IsNotExist(err) {
		return ""
	}
	path, err = filepath.Abs("conf.toml")
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
