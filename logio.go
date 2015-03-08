package main

import (
	"fmt"
	"net"
)

type logConn struct {
	net.Conn
}

func (conn *logConn) Read(p []byte) (n int, err error) {
	fmt.Printf("<- %s", p)
	return conn.Conn.Read(p)
}

func (conn *logConn) Write(p []byte) (n int, err error) {
	fmt.Printf("-> %s", p)
	return conn.Conn.Write(p)
}

func LogConn(conn net.Conn) net.Conn {
	return &logConn{conn}
}
