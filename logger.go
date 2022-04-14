package sidecar

import (
	"log"
	"os"
)

func LogRecord(fd *os.File, level string, context string) {
	Logger := log.New(fd, "["+level+"] ", log.LstdFlags)
	Logger.Println(context)
}
