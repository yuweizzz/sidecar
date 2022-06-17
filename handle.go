package sidecar

import (
	"io"
	"net"
	"net/http"
)

func HandleHttp(server string, subpath string, key string, value string, writer http.ResponseWriter, in_req *http.Request) {
	dest_url := in_req.URL
	dest_url.Scheme = "https"
	dest_url.Host = server
	in_path := dest_url.Path
	dest_url.Path = "/" + subpath + "/" + in_req.Host + in_path
	in_req.Host = server
	in_req.Header.Add(key, value)
	resp, _ := http.DefaultTransport.RoundTrip(in_req)
	for key, value := range resp.Header {
		writer.Header().Set(key, value[0])
	}
	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}

func HandleHttps(watcher *Listener, writer http.ResponseWriter, in_req *http.Request) {
	conn_local, conn := net.Pipe()
	writer.WriteHeader(http.StatusOK)
	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		http.Error(writer, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	conn_remote_proxy, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(conn_remote_proxy, conn_local)
	go transfer(conn_local, conn_remote_proxy)
	go func(){
		watcher.Chan <- conn
	}()
}

func transfer(write io.WriteCloser, read io.ReadCloser) {
	defer write.Close()
	defer read.Close()
	io.Copy(write, read)
}
