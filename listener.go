package sidecar

import (
	"net"
)

type Listener struct {
	Chan chan net.Conn
}

func (l *Listener) Accept() (net.Conn, error) {
	return <-l.Chan, nil
}

func (l *Listener) Close() error {
	return nil
}

func (l *Listener) Addr() net.Addr {
	return nil
}
