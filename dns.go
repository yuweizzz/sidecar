package sidecar

import (
	"context"
	"net"
	"net/http"
	"time"
)

var (
	globalResolver   string
	defaultTransport http.RoundTripper
)

func ChangeResolver(ipAddr string) {
	globalResolver = ""
	defaultTransport = http.DefaultTransport.(*http.Transport).Clone()
	valid := net.ParseIP(ipAddr)
	if valid != nil {
		globalResolver = ipAddr
		customResolverDialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: time.Duration(5000) * time.Millisecond,
					}
					return d.DialContext(ctx, "udp", ipAddr+":53")
				},
			},
		}
		dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return customResolverDialer.DialContext(ctx, network, addr)
		}
		defaultTransport.(*http.Transport).DialContext = dialContext
	}
}
