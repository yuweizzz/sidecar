package sidecar

import (
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

type Proxy struct {
	Listener *Listener
	server   *http.Server
	port     string
	logger   *os.File
}

func NewProxyServer(port int, fd *os.File) *Proxy {
	port_info := ":" + strconv.Itoa(port)
	listener := &Listener{Chan: make(chan net.Conn)}
	server := &http.Server{
		Addr: port_info,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ifHttpRequest(r.URL.Scheme) {
				proxyHandleHttp(w, r)
			} else {
				proxyHandleHttps(listener, w, r)
			}
		}),
	}
	return &Proxy{
		Listener: listener,
		server:   server,
		port:     port_info,
		logger:   fd,
	}
}

func ifHttpRequest(scheme string) bool {
	if scheme == "http" {
		return true
	}
	return false
}

func proxyHandleHttp(w http.ResponseWriter, r *http.Request) {
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()
	for k, v := range response.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func proxyHandleHttps(l *Listener, w http.ResponseWriter, r *http.Request) {
	proxy_income, proxy_output := net.Pipe()
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	next_proxy, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(next_proxy, proxy_output)
	go transfer(proxy_output, next_proxy)
	go func() {
		l.Chan <- proxy_income
	}()
}

func (p *Proxy) Run() {
	p.server.ListenAndServe()
}

func transfer(write io.WriteCloser, read io.ReadCloser) {
	defer write.Close()
	defer read.Close()
	io.Copy(write, read)
}
