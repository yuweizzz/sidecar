package main

import (
	"io"
	"net/http"
)

func HandleHttp(w http.ResponseWriter, r *http.Request) {
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func main() {
	server := &http.Server{
		Addr: ":6699",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			HandleHttp(w, r)
		}),
	}
	server.ListenAndServe()
}
