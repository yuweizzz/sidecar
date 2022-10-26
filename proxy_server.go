package sidecar

import (
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Proxy struct {
	Listener       *Listener
	server         *http.Server
	port           string
	logger         *os.File
	onlyListenIPv4 bool
}

func NewProxyServer(onlyListenIPv4 bool, port int, fd *os.File, pac *Pac) *Proxy {
	listener := &Listener{Chan: make(chan net.Conn)}
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ifHttpRequest(r.URL.Scheme) {
				proxyHandleHttp(w, r)
				return
			}
			if pac.Matcher == nil {
				proxyHandleHttps(listener, w, r)
			} else {
				if pac.Compare(r) {
					proxyHandleHttps(listener, w, r)
				} else {
					directHandleHttps(w, r)
				}
			}
		}),
	}
	server.Handler = http.AllowQuerySemicolons(server.Handler)
	return &Proxy{
		Listener:       listener,
		server:         server,
		port:           ":" + strconv.Itoa(port),
		logger:         fd,
		onlyListenIPv4: onlyListenIPv4,
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

func directHandleHttps(w http.ResponseWriter, r *http.Request) {
	// connect Method is empty scheme, add "https" here
	r.URL.Scheme = "https"
	// dail timeout maybe get: Unsolicited response received on idle HTTP channel
	dest_conn, err := net.DialTimeout("tcp", r.Host, 20*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	_, err = io.WriteString(client_conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	Info("Send Request Directly, Host: ", r.Host)
	go transfer(client_conn, dest_conn)
	go transfer(dest_conn, client_conn)
}

func proxyHandleHttps(l *Listener, w http.ResponseWriter, r *http.Request) {
	proxy_income, proxy_output := net.Pipe()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	next_proxy, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	_, err = io.WriteString(next_proxy, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	l.SetDest(r.Host)
	Info("Send Request via Proxy, Host: ", r.Host)
	go transfer(next_proxy, proxy_output)
	go transfer(proxy_output, next_proxy)
	go func() {
		l.Chan <- proxy_income
	}()
}

func (p *Proxy) Run() {
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

func transfer(write io.WriteCloser, read io.ReadCloser) {
	defer write.Close()
	defer read.Close()
	io.Copy(write, read)
}
