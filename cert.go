package sidecar

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

func ReadPriKey(name string) (pri *rsa.PrivateKey) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(raw)
	if pri, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		panic(err)
	}
	return
}

func GenAndSavePriKey(fd *os.File) (pri *rsa.PrivateKey) {
	defer fd.Close()
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	pem.Encode(fd, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pri),
	})
	return
}

func ReadRootCert(name string) (crt *x509.Certificate) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(raw)
	if crt, err = x509.ParseCertificate(block.Bytes); err != nil {
		panic(err)
	}
	return
}

func GenAndSaveRootCert(fd *os.File, pri *rsa.PrivateKey) (crt *x509.Certificate) {
	defer fd.Close()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() / 100000),
		Subject: pkix.Name{
			CommonName:   "Go-sidecar Root Certificate",
			Organization: []string{"Go-sidecar"},
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
		panic(err)
	}
	pem.Encode(fd, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bytes,
	})
	crt, err = x509.ParseCertificate(bytes)
	if err != nil {
		panic(err)
	}
	return
}

func GenTLSCert(hostname string, crt *x509.Certificate, pri *rsa.PrivateKey) (tls_cert *tls.Certificate, err error) {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() / 100000),
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"Go-sidecar"},
		},
		NotBefore:          time.Now().Add(-time.Hour * 48),
		NotAfter:           time.Now().Add(time.Hour * 24 * 365),
		SignatureAlgorithm: x509.SHA256WithRSA,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	template.DNSNames = []string{hostname}
	bytes, err := x509.CreateCertificate(rand.Reader, template, crt, &pri.PublicKey, pri)
	if err != nil {
		panic(err)
	}
	tls_cert = &tls.Certificate{
		Certificate: [][]byte{bytes},
		PrivateKey:  pri,
	}
	return
}
