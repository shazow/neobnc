package main

import (
	"net"

	"github.com/sorcix/irc"
)

type ircServer struct {
	net.Conn
}

type Relay struct {
	// TODO: Handle many clients
	connected *Client
	// TODO: Handle many servers
	irc    *ircServer
	prefix *irc.Prefix
}

func (r *Relay) Join(c *Client) error {
	logger.Info("Successfully joined.")

	// TODO: Unhardcode all of this:

	r.prefix = &irc.Prefix{
		Name: "name",
		User: "user",
		Host: "host",
	}
	err := c.Encode(&irc.Message{
		Prefix:  r.prefix,
		Command: irc.RPL_WELCOME,
		Params:  []string{r.prefix.User, "Welcome!"},
	})
	if err != nil {
		return err
	}

	go func() {
		defer c.Close()
		for {
			_, err := c.Decode()
			if err != nil {
				logger.Error(err)
				return
			}
		}
	}()

	return nil
}
