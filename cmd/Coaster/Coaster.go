package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"io"
	"path/filepath"
	"time"
	"fmt"
	"github.com/BurntSushi/toml"
	"os/signal"
	"syscall"
	"log"
)

func CreateDirForCert() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	_, err = os.Stat("Certificates")
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll("Certificates", os.ModePerm)
		}
	}
	return filepath.Join(dir, "Certificates"), nil
}

func GenAndSavePriKey(path string) (*rsa.PrivateKey, error) {
	path = filepath.Join(path, "CA.pri")
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	fd, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	pem.Encode(fd, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pri),
	})
	fd.Close()
	return pri, nil
}

func CreateAndSaveRootCert(path string, pri *rsa.PrivateKey) (*x509.Certificate, error) {
	path = filepath.Join(path, "CA.crt")
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() / 100000),
		Subject: pkix.Name{
			CommonName: "Coaster Root Certificate",
			Organization: []string{"Coaster"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageEmailProtection,
			x509.ExtKeyUsageTimeStamping,
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageMicrosoftCommercialCodeSigning,
			x509.ExtKeyUsageMicrosoftServerGatedCrypto,
			x509.ExtKeyUsageNetscapeServerGatedCrypto,
		},
	}
	bytes, err := x509.CreateCertificate(rand.Reader, template, template, &pri.PublicKey, pri)
	if err != nil {
		return nil, err
	}
	fd, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	pem.Encode(fd, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bytes,
	})
	fd.Close()
	crt, err := x509.ParseCertificate(bytes)
	if err != nil {
		return nil, err
	}
	return crt, nil
}

func GenTLSCertificate(hostname string, crt *x509.Certificate, pri *rsa.PrivateKey) (*tls.Certificate, error){
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() / 100000),
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"Coaster"},
		},
		NotBefore:          time.Now().Add(-time.Hour * 48),
		NotAfter:           time.Now().Add(time.Hour * 24 * 365),
		SignatureAlgorithm: x509.SHA256WithRSA,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	template.DNSNames = []string{hostname}
	bytes, err := x509.CreateCertificate(rand.Reader, template, crt, &pri.PublicKey, pri)
	if err != nil {
		return nil, err
	}
	TLSCert := &tls.Certificate{
		Certificate: [][]byte{bytes},
		PrivateKey:  pri,
	}
	return TLSCert, nil
}

func HandleHttp( server string, subpath string, key string, value string, w http.ResponseWriter, r *http.Request) {
	var  body  io.Reader
	NewURL := server + "/"+ subpath + "/" + r.Host + r.URL.String()
	fmt.Println(NewURL)
	Re,_ := http.NewRequest(r.Method , NewURL , body)
	Re.Header.Add(key, value)
	Resp,_ := http.DefaultTransport.RoundTrip(Re)
	w.WriteHeader(Resp.StatusCode)
	io.Copy(w, Resp.Body)
}

func HandleHttps(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", "127.0.0.1:443", 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(client_conn, dest_conn)
	go transfer(dest_conn, client_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}




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

type config struct {
	Server string
	ComplexPath string
	CustomHeaderName string
	CustomHeaderValue string
}

func ReadConfig() * config {
		filePath, err := filepath.Abs("./conf.toml")
		if err != nil {
			panic(err)
		}
		fmt.Printf("parse toml file once. filePath: %s\n", filePath)
		if _ , err := toml.DecodeFile(filePath, &cfg); err != nil {
			panic(err)
		}
	return cfg
}

var (
	cfg *config
)

func main() {

	if os.Args[1] == "start" {
		ReadConfig()
		path, _ := CreateDirForCert()
		key,_ := GenAndSavePriKey(path)
		crt,_ := CreateAndSaveRootCert(path, key)
		server := &http.Server{
			Addr: ":443",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				HandleHttp(cfg.Server, cfg.ComplexPath , cfg.CustomHeaderName , cfg.CustomHeaderValue, w, r)
			}),
			IdleTimeout:  5 * time.Second,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // disable http2
			TLSConfig: &tls.Config{
				GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
					return GenTLSCertificate(chi.ServerName, crt, key)
				},
			},
		}
		proxy := &http.Server{
			Addr: ":4396",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				HandleHttps(w, r)
			}),
		}
		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go proxy.ListenAndServe()
		go func() {
			sig := <-sigs
			fmt.Println()
			fmt.Println(sig)
			done <- true
		}()
		LogRecord("info", "Start Server ... ")
		go server.ListenAndServeTLS("", "")
		LogRecord("info", "awaiting signal")
		<-done
		LogRecord("info", "exiting")
	}
}
