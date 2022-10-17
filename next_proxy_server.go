package sidecar

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type NextProxy struct {
	Listener      *Listener
	server        *http.Server
	ca            *x509.Certificate
	privateKey    *rsa.PrivateKey
	logger        *os.File
	destination   string
	complexPath   string
	customHeaders map[string]string
}

func NewNextProxyServer(
	l *Listener, ca *x509.Certificate, pri *rsa.PrivateKey, fd *os.File,
	destination string, complex_path string, headers map[string]string,
) *NextProxy {
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ifWebSocketReq(r) {
				nextProxyHandleWs(destination, complex_path, headers, w, r)
			} else {
				nextProxyHandleHttp(destination, complex_path, headers, w, r)
			}
		}),
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		TLSConfig: &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if chi.ServerName == "" {
					return GenTLSCert(strings.Split(l.Dest(), ":")[0], ca, pri)
				}
				return GenTLSCert(chi.ServerName, ca, pri)
			},
		},
	}
	server.Handler = http.AllowQuerySemicolons(server.Handler)
	return &NextProxy{
		Listener:      l,
		server:        server,
		ca:            ca,
		privateKey:    pri,
		logger:        fd,
		destination:   destination,
		complexPath:   complex_path,
		customHeaders: headers,
	}
}

func ifWebSocketReq(in_req *http.Request) bool {
	if in_req.Header.Get("Upgrade") == "websocket" && in_req.Header.Get("Connection") == "Upgrade" {
		return true
	}
	return false
}

func nextProxyHandleHttp(server string, subpath string, headers map[string]string, writer http.ResponseWriter, in_req *http.Request) {
	dest_url := in_req.URL
	dest_url.Scheme = "https"
	dest_url.Host = server
	in_path := dest_url.Path
	dest_url.Path = "/" + subpath + "/" + in_req.Host + in_path
	in_req.Host = server
	for k, v := range headers {
		in_req.Header.Add(k, v)
	}
	resp, err := http.DefaultTransport.RoundTrip(in_req)
	if err != nil {
		return
	}
	for k, v := range resp.Header {
		writer.Header()[k] = v
	}
	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}

func nextProxyHandleWs(server string, subpath string, headers map[string]string, writer http.ResponseWriter, in_req *http.Request) {
	tls_conn, err := tls.Dial("tcp", server+":443", nil)
	if err != nil {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dest_url := in_req.URL
	dest_url.Scheme = "http"
	dest_url.Host = server
	in_path := dest_url.Path
	dest_url.Path = "/" + subpath + "/" + in_req.Host + in_path
	for k, v := range headers {
		in_req.Header.Add(k, v)
	}
	in_req.Host = server
	in_req.URL = dest_url
	in_req.RequestURI = dest_url.RequestURI()
	dump, err := httputil.DumpRequest(in_req, true)
	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		http.Error(writer, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	proxy, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
	}
	tls_conn.Write(dump)
	go transfer(proxy, tls_conn)
	go transfer(tls_conn, proxy)
}

func (p *NextProxy) Run() {
	p.server.ServeTLS(p.Listener, "", "")
}

func (p *NextProxy) WatchSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	Info("Awaiting signal......")
	<-done
}
