package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	req, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	if err != nil {
		fmt.Println("request error")
	}
	resp, _ := http.DefaultTransport.RoundTrip(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
