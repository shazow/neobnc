package main

import (
	"bytes"
	"net"
)

type logConn struct {
	net.Conn
}

func byteString(buf []byte) []byte {
	i := bytes.IndexByte(buf, '\x00')
	if i < 0 {
		return buf[:]
	}
	return buf[:i]
}

func (conn *logConn) Read(p []byte) (n int, err error) {
	logger.Debugf("<- %+q", byteString(p))
	return conn.Conn.Read(p)
}

func (conn *logConn) Write(p []byte) (n int, err error) {
	logger.Debugf("-> %+q", byteString(p))
	return conn.Conn.Write(p)
}

func LogConn(conn net.Conn) net.Conn {
	return &logConn{conn}
}
