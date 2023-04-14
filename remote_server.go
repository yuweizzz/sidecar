package sidecar

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type RemoteServer struct {
	server         *http.Server
	logger         *os.File
	port           string
	onlyListenIPv4 bool
	priKeyPath     string
	certPath       string
	complexPath    string
	customHeaders  map[string]string
}

func NewRemoteServer(
	port string, fd *os.File, cert_path string, prikey_path string, only_listen_ipv4 bool,
	complex_path string, headers map[string]string,
) *RemoteServer {
	server := &http.Server{
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return &RemoteServer{
		server:         server,
		logger:         fd,
		port:           port,
		onlyListenIPv4: only_listen_ipv4,
		priKeyPath:     prikey_path,
		certPath:       cert_path,
		complexPath:    complex_path,
		customHeaders:  headers,
	}
}

func (r *RemoteServer) proxyRequest(w http.ResponseWriter, req *http.Request) {
	for k, v := range r.customHeaders {
		if req.Header.Get(k) != v {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		} else {
			req.Header.Del(k)
		}
	}
	Infos := strings.SplitN(req.RequestURI, "/", 4)
	TrueURL, err := url.Parse("https://" + Infos[2] + "/" + Infos[3])
	req.URL = TrueURL
	req.Host = Infos[2]
	req.RequestURI = TrueURL.RequestURI()
	Info("Request Info After Rewrite: Host is", req.Host, ", Uri is", req.RequestURI)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (r *RemoteServer) Run() {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/"+r.complexPath+"/", r.proxyRequest)
	if r.onlyListenIPv4 {
		l, err := net.Listen("tcp4", "0.0.0.0"+r.port)
		if err != nil {
			Panic(err)
		}
		r.server.ServeTLS(l, r.certPath, r.priKeyPath)
	}
	r.server.Addr = r.port
	r.server.ListenAndServeTLS(r.certPath, r.priKeyPath)
}
