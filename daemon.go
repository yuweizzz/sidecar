package sidecar

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

type Daemon struct {
	Pid          int
	WorkDir      string
	CertPath     string
	PriKeyPath   string
	LockFilePath string
	LogLevel     string
	Logger       *os.File
	PriKey       *rsa.PrivateKey
	Cert         *x509.Certificate
}

func (d *Daemon) Perpare(backgroud bool) {
	if backgroud {
		fmt.Println(d.WorkDir)
		log_fd := CreateFileIfNotExist(d.WorkDir + "/sidecar.log")
		if log_fd == nil {
			log_fd = OpenExistFile(d.WorkDir + "/sidecar.log")
		}
		d.Logger = log_fd
	} else {
		d.Logger = os.Stdout
	}
	Initial(d.LogLevel, d.Logger)
	pid := ReadLock(d.LockFilePath)
	Info("Detect if Server is running .....")
	// if lock exist
	if pid != 0 {
		if runtime.GOOS == "linux" {
			alive := DetectProcess(pid)
			// if process alive
			if alive {
				exitWhenLocked(pid)
			} else {
				// if process not alive
				Info("File sidecar.lock exist, file path is ", d.LockFilePath, ", but process is not running.....")
				RemoveLock(d.LockFilePath)
			}
		} else {
			exitWhenLocked(pid)
		}
	}
	d.Pid = os.Getpid()
	writeLock(d.Pid, d.LockFilePath)
}

func (d *Daemon) LoadCertAndPriKey() {
	if pri_file_path := DetectFile(d.PriKeyPath); pri_file_path == "" {
		pri_fd := CreateFileIfNotExist(d.PriKeyPath)
		d.PriKey = GenAndSavePriKey(pri_fd)
		Info("Generate new privatekey, privatekey file save to ", d.PriKeyPath)
	} else {
		d.PriKey = ReadPriKey(d.PriKeyPath)
		Info("Use exist privatekey, file path is ", pri_file_path)
	}
	if crt_file_path := DetectFile(d.CertPath); crt_file_path == "" {
		crt_fd := CreateFileIfNotExist(d.CertPath)
		d.Cert = GenAndSaveRootCert(crt_fd, d.PriKey)
		Info("Generate new certificate, certificate file save to ", d.CertPath)
	} else {
		d.Cert = ReadRootCert(d.CertPath)
		Info("Use exist certificate, file path is ", crt_file_path)
	}
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

func RemoveLock(path string) {
	err := os.Remove(path)
	if err != nil {
		panic(err)
	}
}

func exitWhenLocked(pid int) {
	fmt.Println("Start Server failed because sidecar.lock exist, maybe Server is already running and pid is ", pid)
	fmt.Println("If you confirm Server is not running, remove sidecar.lock and retry.")
	os.Exit(2)
}

func (d *Daemon) WatchSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	Info("Awaiting signal......")
	<-done
	Info("Except signal, exiting......")
	Info("Remove sidecar.lock......")
	RemoveLock(d.LockFilePath)
}
