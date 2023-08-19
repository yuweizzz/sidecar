package sidecar

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

func ifHttpRequest(scheme string) bool {
	if scheme == "http" {
		return true
	}
	return false
}

func proxyHandleHttp(w http.ResponseWriter, r *http.Request) {
	response, err := defaultTransport.RoundTrip(r)
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
	customDialer := &net.Dialer{
		// dail timeout maybe get: Unsolicited response received on idle HTTP channel
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
	// because scheme is https, so r.Host will be "host:port"
	dest_conn, err := customDialer.Dial("tcp", r.Host)
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

func transfer(write io.WriteCloser, read io.ReadCloser) {
	defer write.Close()
	defer read.Close()
	io.Copy(write, read)
}
