package sidecar

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

type MitMServer struct {
	Listener      *Listener
	server        *http.Server
	logger        *os.File
	destination   string
	complexPath   string
	customHeaders map[string]string
}

func NewMitMServer(
	l *Listener, cache *CertLRU, fd *os.File,
	destination string, complex_path string, headers map[string]string,
) *MitMServer {
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ifWebSocketReq(r) {
				MitMHandleWs(destination, complex_path, headers, w, r)
			} else {
				MitMHandleHttp(destination, complex_path, headers, w, r)
			}
		}),
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		TLSConfig: &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				sni := chi.ServerName
				if sni == "" {
					sni = strings.Split(l.Dest(), ":")[0]
				}
				return cache.GetCert(sni)
			},
		},
	}
	server.Handler = http.AllowQuerySemicolons(server.Handler)
	return &MitMServer{
		Listener:      l,
		server:        server,
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

func MitMHandleHttp(server string, subpath string, headers map[string]string, writer http.ResponseWriter, in_req *http.Request) {
	dest_url := in_req.URL
	dest_url.Scheme = "https"
	dest_url.Host = server
	in_path := dest_url.Path
	Debug("Send Https Request to Remote Proxy, Host: ", in_req.Host, ", Uri: ", in_path)
	dest_url.Path = "/" + subpath + "/" + in_req.Host + in_path
	in_req.Host = server
	for k, v := range headers {
		in_req.Header.Add(k, v)
	}
	resp, err := defaultTransport.RoundTrip(in_req)
	if err != nil {
		return
	}
	for k, v := range resp.Header {
		writer.Header()[k] = v
	}
	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}

func MitMHandleWs(server string, subpath string, headers map[string]string, writer http.ResponseWriter, in_req *http.Request) {
	customDialer := &net.Dialer{
		Timeout: time.Duration(20) * time.Second,
	}
	if globalResolver != "" {
		customDialer.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(5000) * time.Millisecond,
				}
				return d.DialContext(ctx, "udp", globalResolver+":53")
			},
		}
	}
	// tls_conn, err := tls.Dial("tcp", server+":443", nil)
	tls_conn, err := tls.DialWithDialer(customDialer, "tcp", server+":443", nil)
	if err != nil {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dest_url := in_req.URL
	dest_url.Scheme = "http"
	dest_url.Host = server
	in_path := dest_url.Path
	dest_url.Path = "/" + subpath + "/" + in_req.Host + in_path
	Debug("Send WebSocket Request to Remote Proxy, Host: ", in_req.Host, ", Uri: ", in_path)
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

func (p *MitMServer) Run() {
	p.server.ServeTLS(p.Listener, "", "")
}
