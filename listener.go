package sidecar

import (
	"net"
)

type Listener struct {
	Chan chan net.Conn
	host string
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

func (l *Listener) Dest() string {
	return l.host
}

func (l *Listener) SetDest(host string) {
	l.host = host
}