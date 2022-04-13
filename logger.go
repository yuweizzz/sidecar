package Coaster

import (
	"log"
	"os"
)

func LogRecord(fd *os.File, level string, context string) {
	defer fd.Close()
	Logger := log.New(fd, "["+level+"] ", log.LstdFlags)
	Logger.Println(context)
}
