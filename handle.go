package sidecar

import (
	"io"
	"net"
	"net/http"
)

func HandleHttp(server string, subpath string, key string, value string, writer http.ResponseWriter, in_req *http.Request) {
	var body io.Reader
	dest_url := server + "/" + subpath + "/" + in_req.Host + in_req.URL.String()
	out_req, _ := http.NewRequest(in_req.Method, dest_url, body)
	out_req.Header.Add(key, value)
	resp, _ := http.DefaultTransport.RoundTrip(out_req)
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
