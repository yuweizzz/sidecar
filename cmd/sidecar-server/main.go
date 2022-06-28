package main

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/yuweizzz/sidecar"
)

func main() {
	pid := sidecar.ReadLock()
	// if lock exist
	if pid != 0 {
		alive := sidecar.DetectProcess(pid)
		// if process alive
		if alive {
			fmt.Println("Maybe Server is running and pid is", pid)
			fmt.Println("If Server is not running, remove sidecar-server.lock and retry.")
			panic("Run failed, sidecar-server.lock exist.")
			// if process not alive
		} else {
			sidecar.RemoveLock()
		}
	}
	var cfg *sidecar.Config
	config_file_path := sidecar.DetectFile("conf.toml")
	if config_file_path == "" {
		panic("Run failed, conf.toml not exist.")
	} else {
		cfg = sidecar.ReadConfig(config_file_path)
	}
	log_path := sidecar.CreateDirIfNotExist("log")
	if log_path == "" {
		panic("Create dir for log failed.")
	}
	cert_path := sidecar.CreateDirIfNotExist("certificate")
	if cert_path == "" {
		panic("Create dir for certificate failed.")
	}
	log_fd := sidecar.CreateFileIfNotExist("log/server.log")
	if log_fd == nil {
		log_fd = sidecar.OpenExistFile("log/server.log")
	}
	sidecar.LogRecord(log_fd, "info", "Start Server......")
	sidecar.LogRecord(log_fd, "info", "log location: "+log_path)
	sidecar.LogRecord(log_fd, "info", "certificate location: "+cert_path)
	var pri *rsa.PrivateKey
	var crt *x509.Certificate
	if pri_file_path := sidecar.DetectFile("certificate/sidecar.pri"); pri_file_path == "" {
		pri_fd := sidecar.CreateFileIfNotExist("certificate/sidecar.pri")
		pri = sidecar.GenAndSavePriKey(pri_fd)
		sidecar.LogRecord(log_fd, "info", "Generate new privatekey......")
	} else {
		pri = sidecar.ReadPriKey("certificate/sidecar.pri")
		sidecar.LogRecord(log_fd, "info", "Use exist privatekey......")
	}
	if crt_file_path := sidecar.DetectFile("certificate/sidecar.crt"); crt_file_path == "" {
		crt_fd := sidecar.CreateFileIfNotExist("certificate/sidecar.crt")
		crt = sidecar.GenAndSaveRootCert(crt_fd, pri)
		sidecar.LogRecord(log_fd, "info", "Generate new certificate......")
	} else {
		crt = sidecar.ReadRootCert("certificate/sidecar.crt")
		sidecar.LogRecord(log_fd, "info", "Use exist certificate......")
	}
	proxy := sidecar.NewProxyServer(cfg.ProxyPort, log_fd)
	forwarder := sidecar.NewNextProxyServer(proxy.Listener, crt, pri, log_fd, cfg.Server, cfg.ComplexPath, cfg.CustomHeaders)
	pid = os.Getpid()
	sidecar.WriteLock(pid)
	fmt.Println("Now Server is running and pid is", pid)
	go proxy.Run()
	go forwarder.Run()
	forwarder.WatchSignal()
	sidecar.LogRecord(log_fd, "info", "Except signal, exiting......")
	sidecar.RemoveLock()
	defer log_fd.Close()
}
