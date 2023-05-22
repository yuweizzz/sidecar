package sidecar

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
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

// show windows console
// https://stackoverflow.com/questions/23743217/printing-output-to-a-command-window-when-golang-application-is-compiled-with-ld
func console(show bool) {
	getWin := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
	showWin := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	hwnd, _, _ := getWin.Call()
	if hwnd == 0 {
		return
	}
	if show {
		var SW_RESTORE uintptr = 9
		showWin.Call(hwnd, SW_RESTORE)
	} else {
		var SW_HIDE uintptr = 0
		showWin.Call(hwnd, SW_HIDE)
	}
}

func readLock(path string) (pid int) {
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
	Info("Remove sidecar.lock done.")
}

func exitWhenLocked(pid int) {
	fmt.Println("Start Server failed because sidecar.lock exist, maybe Server is already running and pid is ", pid)
	fmt.Println("If you confirm Server is not running, remove sidecar.lock and retry.")
	os.Exit(2)
}

func (d *Daemon) Perpare(backgroud bool) {
	console(!backgroud)
	if backgroud {
		log_fd := CreateFileIfNotExist(d.WorkDir + "/sidecar.log")
		if log_fd == nil {
			log_fd = OpenExistFile(d.WorkDir + "/sidecar.log")
		}
		d.Logger = log_fd
	} else {
		d.Logger = os.Stdout
	}
	Initial(d.LogLevel, d.Logger)
	pid := readLock(d.LockFilePath)
	Info("Detect if Server is running .....")
	// if lock exist
	if pid != 0 {
		exitWhenLocked(pid)
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

func (d *Daemon) WatchSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	Info("Now Server is running and pid is " + strconv.Itoa(d.Pid))
	Info("Awaiting signal......")
	<-done
	Info("Except signal, exiting......")
}

// run in backgroud
func StartDaemonProcess(configPath string, serviceType string) {
	cmd := &exec.Cmd{
		Path:   os.Args[0],
		Env:    []string{"SPECIAL_MARK=ENABLED", "CONF_PATH=" + configPath, "TYPE=" + serviceType},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func StopDaemonProcess(workDir string) {
	lockPath := workDir + "/sidecar.lock"
	pid := readLock(lockPath)
	// if lock exist
	if pid != 0 {
		proc, _ := os.FindProcess(pid)
		removeLock(lockPath)
		proc.Kill()
	} else {
		Error("Now sidecar.lock is not exist, server is stopped")
	}
}
