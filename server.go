package main

import (
	"errors"
	"net"
	"sync"

	"github.com/sorcix/irc"
)

var ErrInvalidRelayKey = errors.New("invalid relay key")

type Host struct {
	sync.Mutex
	relays map[string]*Relay
	Debug  bool
}

func NewHost() *Host {
	return &Host{
		relays: make(map[string]*Relay),
	}
}

// Start server on a given host listener
func (h Host) Start(l net.Listener) error {
	for {
		conn, err := l.Accept()

		if err != nil {
			logger.Errorf("Failed to accept connection: %v", err)
			return err
		}

		if h.Debug {
			conn = LogConn(conn)
		}

		// Goroutineify to resume accepting sockets early
		go func() {
			client, err := NewClient(conn)
			if err != nil {
				logger.Errorf("Failed to handshake: %v", err)
				return
			}
			err = h.Join(client)
			if err != nil {
				logger.Errorf("Failed to join: %v", err)
				return
			}
		}()
	}

	return nil
}

// Get returns a relay instance of the requested key
func (h *Host) Get(key string) (*Relay, error) {
	h.Lock()
	relay, ok := h.relays[key]
	if !ok {
		logger.Warningf("Creating relay with key: %q", key)
		relay = &Relay{}
		h.relays[key] = relay
		// TODO: Deny invalid keys
		// 	return nil, ErrInvalidRelayKey
	}
	h.Unlock()
	return relay, nil
}

// Join starts listening and relaying for Client
func (h Host) Join(c *Client) error {
	// TODO: Get this from password field
	message, err := c.DecodeWhen(irc.PASS)
	if err != nil {
		return err
	}
	if len(message.Params) != 1 {
		return ErrInvalidRelayKey
	}
	key := message.Params[0]
	relay, err := h.Get(key)
	if err != nil {
		return err
	}
	relay.Join(c)
	return nil
}
