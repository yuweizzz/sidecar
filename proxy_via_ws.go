package sidecar

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type ProxyViaWss struct {
	server         *http.Server
	logger         *os.File
	pac            *Pac
	port           string
	onlyListenIPv4 bool
	destination    string
	complexPath    string
	customHeaders  map[string]string
}

func NewProxyViaWss(fd *os.File, pac *Pac,
	onlyListenIPv4 bool, port int, destination string, complex_path string, headers map[string]string,
) *ProxyViaWss {
	server := &http.Server{
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return &ProxyViaWss{
		server:         server,
		logger:         fd,
		pac:            pac,
		port:           ":" + strconv.Itoa(port),
		onlyListenIPv4: onlyListenIPv4,
		destination:    destination,
		complexPath:    complex_path,
		customHeaders:  headers,
	}
}

func (i *ProxyViaWss) Run() {
	i.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ifHttpRequest(r.URL.Scheme) {
			proxyHandleHttp(w, r)
			return
		}
		if i.pac.Matcher == nil {
			i.proxyHandleHttpsToWss(w, r)
		} else {
			if i.pac.Compare(r) {
				i.proxyHandleHttpsToWss(w, r)
			} else {
				directHandleHttps(w, r)
			}
		}
	})
	if i.onlyListenIPv4 {
		l, err := net.Listen("tcp4", "0.0.0.0"+i.port)
		if err != nil {
			Panic(err)
		}
		i.server.Serve(l)
	}
	i.server.Addr = i.port
	i.server.ListenAndServe()
}

func (i *ProxyViaWss) proxyHandleHttpsToWss(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	_, err = io.WriteString(client_conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		Error("Error in connection establish: ", err)
		return
	}

	u := url.URL{Scheme: "wss", Host: i.destination, Path: "/" + i.complexPath + "/"}
	wss_req_headers := http.Header{}
	for k, v := range i.customHeaders {
		wss_req_headers.Set(k, v)
	}
	wss_req_headers.Set("destination", r.Host)
	Info("Send Request via Wss tunnel, Host: ", r.Host)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), wss_req_headers)
	if err != nil {
		Error("Error in proxy connect to remote Websocket: ", err)
		return
	}
	defer func() {
		c.Close()
	}()

	go func() {
		for {
			buff := make([]byte, 2048)
			length, err := client_conn.Read(buff)
			if err != nil {
				Debug("Error in read from tcp connect: ", err)
				return
			}
			err = c.WriteMessage(websocket.BinaryMessage, buff[:length])
			if err != nil {
				Debug("Error in write data to websocket: ", err)
				return
			}
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			Debug("Error in read data from websocket: ", err)
			return
		}
		_, err = client_conn.Write([]byte(message))
		if err != nil {
			Debug("Error in write data to tcp connect: ", err)
			return
		}
	}
}
