package main

import (
	"bytes"
	"net"
)

type logConn struct {
	net.Conn
}

func (conn *logConn) Read(p []byte) (n int, err error) {
	logger.Debugf("<- %+q", bytes.Trim(p, "\x00"))
	return conn.Conn.Read(p)
}

func (conn *logConn) Write(p []byte) (n int, err error) {
	logger.Debugf("-> %+q", bytes.Trim(p, "\x00"))
	return conn.Conn.Write(p)
}

func LogConn(conn net.Conn) net.Conn {
	return &logConn{conn}
}
