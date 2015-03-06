package main

import (
	"errors"
	"net"

	"github.com/dchest/uniuri"
	"github.com/sorcix/irc"
)

var ErrMismatchedPong = errors.New("mismatched pong")

type ircServer struct {
	net.Conn
}

type Relay struct {
	// TODO: Handle many clients
	connected *Client
	// TODO: Handle many servers
	irc *ircServer
}

func (r *Relay) Join(c *Client) error {
	pingKey := uniuri.New()
	c.Encode(&irc.Message{
		Command: irc.PING,
		Params:  []string{pingKey},
	})

	msg, err := c.DecodeWhen(irc.PONG)
	if err != nil {
		return err
	}

	if len(msg.Params) != 1 || msg.Params[0] != pingKey {
		return ErrMismatchedPong
	}

	logger.Info("Successfully joined.")
	r.connected = c
	return nil
}
