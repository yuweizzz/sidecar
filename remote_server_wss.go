package sidecar

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type RemoteServerWss struct {
	server         *http.Server
	logger         *os.File
	port           int
	onlyListenIPv4 bool
	priKeyPath     string
	certPath       string
	complexPath    string
	customHeaders  map[string]string
}

func NewRemoteServerWss(
	fd *os.File, port int, only_listen_ipv4 bool, cert_path string, prikey_path string,
	complex_path string, headers map[string]string,
) *RemoteServerWss {
	server := &http.Server{
		IdleTimeout:  5 * time.Second,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return &RemoteServerWss{
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

func (ws *RemoteServerWss) proxyRequest(w http.ResponseWriter, req *http.Request) {
	for k, v := range ws.customHeaders {
		if req.Header.Get(k) != v {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		} else {
			req.Header.Del(k)
		}
	}
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		Debug("Websocket handshake err: ", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	endpointHost := req.Header.Get("destination")
	Info("Request destination is ", endpointHost)
	dest_conn, err := net.Dial("tcp", endpointHost)
	if err != nil {
		Debug("Error in tcp connect to destination: ", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	Debug("tcp connect to destination is: ", dest_conn.RemoteAddr())
	defer func() {
		dest_conn.Close()
		c.Close()
	}()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				Debug("Error in read from websocket: ", err)
				return
			}
			_, err = dest_conn.Write([]byte(message))
			if err != nil {
				Debug("Error in write data to tcp connect: ", err)
				return
			}
		}
	}()

	for {
		buff := make([]byte, 2048)
		length, err := dest_conn.Read(buff)
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

}

func (ws *RemoteServerWss) Run() {
	addr := ":" + strconv.Itoa(ws.port)
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/"+ws.complexPath+"/", ws.proxyRequest)
	if ws.onlyListenIPv4 {
		l, err := net.Listen("tcp4", "0.0.0.0"+addr)
		if err != nil {
			Panic(err)
		}
		ws.server.ServeTLS(l, ws.certPath, ws.priKeyPath)
	}
	ws.server.Addr = addr
	ws.server.ListenAndServeTLS(ws.certPath, ws.priKeyPath)
}
