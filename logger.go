package main

import (
	"log"
	"os"
	"path/filepath"
)

func FetchFile() (*os.File, error) {
	pwd, _ := os.Getwd()
	fullpath := filepath.Join(pwd, "logs")
	_, err := os.Stat(fullpath)
	if os.IsNotExist(err) {
		os.Mkdir(fullpath, 0744)
	}
	fileName := "logs/proxy.log"
	fd, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	return fd, err
}

func LogRecord(level string, context string) {
	fd, err := FetchFile()

	if err != nil {
		log.Fatalln("open file error !")
	}
	defer fd.Close()

	Logger := log.New(fd, "["+level+"] ", log.LstdFlags)
	Logger.Println(context)
}
