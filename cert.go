package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

func HandleHttp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	path, _ := CreateDirForCert()
	key,_ := GenAndSavePriKey(path)
	crt,_ := CreateAndSaveRootCert(path, key)
	server := &http.Server{
		Addr: ":6699",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			HandleHttp(w, r)
		}),
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // disable http2
		TLSConfig: &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return GenTLSCertificate(chi.ServerName, crt, key)
			},
		},
	}
	server.ListenAndServeTLS("", "")
}
