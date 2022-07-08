package sidecar

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Daemon struct {
	Pid          int
	WorkDir      string
	CertPath     string
	PriKeyPath   string
	LockFilePath string
	Logger       *os.File
	PriKey       *rsa.PrivateKey
	Cert         *x509.Certificate
}

func (d *Daemon) Perpare(backgroud bool) {
	if backgroud {
		log_fd := CreateFileIfNotExist(d.WorkDir + "/server.log")
		if log_fd == nil {
			log_fd = OpenExistFile(d.WorkDir + "server.log")
		}
		d.Logger = log_fd
	} else {
		d.Logger = os.Stdout
	}
	pid := ReadLock(d.LockFilePath)
	// if lock exist
	if pid != 0 {
		alive := DetectProcess(pid)
		// if process alive
		if alive {
			fmt.Println("Maybe Server is running and pid is", pid)
			fmt.Println("If Server is not running, remove sidecar-server.lock and retry.")
			panic("Run failed, sidecar-server.lock exist.")
			// if process not alive
		} else {
			removeLock(d.LockFilePath)
		}
	}
	d.Pid = os.Getpid()
	if pri_file_path := DetectFile(d.PriKeyPath); pri_file_path == "" {
		pri_fd := CreateFileIfNotExist(d.PriKeyPath)
		d.PriKey = GenAndSavePriKey(pri_fd)
		LogRecord(d.Logger, "info", "Generate new privatekey......")
	} else {
		d.PriKey = ReadPriKey(d.PriKeyPath)
		LogRecord(d.Logger, "info", "Use exist privatekey......")
	}
	if crt_file_path := DetectFile(d.CertPath); crt_file_path == "" {
		crt_fd := CreateFileIfNotExist(d.CertPath)
		d.Cert = GenAndSaveRootCert(crt_fd, d.PriKey)
		LogRecord(d.Logger, "info", "Generate new certificate......")
	} else {
		d.Cert = ReadRootCert(d.CertPath)
		LogRecord(d.Logger, "info", "Use exist certificate......")
	}
	writeLock(d.Pid, d.LockFilePath)
}

func ReadLock(path string) (pid int) {
	lock := DetectFile(path)
	if lock == "" {
		return 0
	} else {
		bytes, err := ioutil.ReadFile(path)
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

func writeLock(pid int, path string) {
	pid_str := strconv.Itoa(pid)
	err := ioutil.WriteFile(path, []byte(pid_str), 0444)
	if err != nil {
		panic(err)
	}
}

func removeLock(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}

func (d *Daemon) Clean() {
	LogRecord(d.Logger, "info", "Except signal, exiting......")
	removeLock(d.LockFilePath)
}

//sidecar.LogRecord(log_fd, "info", "Start Server......")
//sidecar.LogRecord(log_fd, "info", "log location: "+log_path)
//sidecar.LogRecord(log_fd, "info", "certificate location: "+cert_path)
//log_path := sidecar.CreateDirIfNotExist("log")
//if log_path == "" {
//	panic("Create dir for log failed.")
//}
//cert_path := sidecar.CreateDirIfNotExist("certificate")
//if cert_path == "" {
//	panic("Create dir for certificate failed.")
//}
