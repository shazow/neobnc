package main

import "net"

type Client struct {
	net.Conn
}

type Server struct {
}

// Start server on a given host listener
func (s Server) Start(host net.Listener) <-chan *Client {
	ch := make(chan *Client)

	go func() {
		defer close(ch)

		for {
			conn, err := host.Accept()

			if err != nil {
				logger.Errorf("Failed to accept connection: %v", err)
				return
			}

			// Goroutineify to resume accepting sockets early
			go func() {
				client, err := s.NewClient(conn)
				if err != nil {
					logger.Errorf("Failed to handshake: %v", err)
					return
				}
				ch <- client
			}()
		}
	}()

	return ch
}

// NewClient returns a new Client based on this Server.
func (s Server) NewClient(conn net.Conn) (*Client, error) {
	logger.Debugf("New client: %v", conn)
	return &Client{conn}, nil
}
