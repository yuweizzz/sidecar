package sidecar

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ProxyViaHttps struct {
	Listener       *Listener
	server         *http.Server
	logger         *os.File
	pac            *Pac
	port           string
	onlyListenIPv4 bool
}

func NewProxyViaHttps(fd *os.File, pac *Pac, onlyListenIPv4 bool, port int) *ProxyViaHttps {
	listener := &Listener{Chan: make(chan net.Conn)}
	server := &http.Server{
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return &ProxyViaHttps{
		Listener:       listener,
		server:         server,
		logger:         fd,
		pac:            pac,
		port:           ":" + strconv.Itoa(port),
		onlyListenIPv4: onlyListenIPv4,
	}
}

func (p *ProxyViaHttps) proxyHandleHttps(w http.ResponseWriter, r *http.Request) {
	proxy_income, proxy_output := net.Pipe()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	next_proxy, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	_, err = io.WriteString(next_proxy, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		Error("Error in connection establish: ", err)
		return
	}
	p.Listener.SetDest(r.Host)
	Info("Send Request via Proxy, Host: ", r.Host)
	go transfer(next_proxy, proxy_output)
	go transfer(proxy_output, next_proxy)
	go func() {
		p.Listener.Chan <- proxy_income
	}()
}

func (p *ProxyViaHttps) Run() {
	p.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ifHttpRequest(r.URL.Scheme) {
			proxyHandleHttp(w, r)
			return
		}
		if p.pac.Matcher == nil {
			p.proxyHandleHttps(w, r)
		} else {
			if p.pac.Compare(r) {
				p.proxyHandleHttps(w, r)
			} else {
				directHandleHttps(w, r)
			}
		}
	})
	p.server.Handler = http.AllowQuerySemicolons(p.server.Handler)
	if p.onlyListenIPv4 {
		l, err := net.Listen("tcp4", "0.0.0.0"+p.port)
		if err != nil {
			Panic(err)
		}
		p.server.Serve(l)
	}
	p.server.Addr = p.port
	p.server.ListenAndServe()
}
