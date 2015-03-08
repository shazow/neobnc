package main

import (
	"io"
	"sync"

	"github.com/sorcix/irc"
)

type Relay struct {
	sync.Mutex
	Nick string
	// TODO: Handle many clients
	connected *Client
	// TODO: Handle many servers
	server *irc.Conn
	prefix *irc.Prefix
}

func (r *Relay) Join(c *Client) error {
	r.Lock()
	defer r.Unlock()

	r.connected = c
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

	if r.server == nil {
		// XXX: Unhardcode this
		r.Connect("localhost:6667")
	}

	go func() {
		// Client loop
		defer c.Close()
		for {
			msg, err := c.Decode()
			if err == io.EOF {
				// Client closed
				logger.Info("Client closed.")
				return
			} else if err != nil {
				logger.Error("Client decode error:", err)
				return
			}
			logger.Debugf("Client: %s", msg)
			if err = r.server.Encode(msg); err != nil {
				logger.Error("Client encode error:", err)
				return
			}
		}
	}()

	return nil
}

func (r *Relay) Connect(addr string) error {
	logger.Infof("Connecting to %s", addr)
	if r.server != nil {
		r.server.Close()
	}
	conn, err := irc.Dial(addr)
	if err != nil {
		return nil
	}

	r.server = conn

	conn.Encode(&irc.Message{
		Command:  irc.USER,
		Params:   []string{r.Nick, "0", "*"},
		Trailing: r.Nick,
	})
	conn.Encode(&irc.Message{
		Command: irc.NICK,
		Params:  []string{r.Nick},
	})

	go func() {
		// Server loop
		defer conn.Close()
		for {
			msg, err := conn.Decode()
			if err == io.EOF {
				logger.Info("Server closed.")
				// TODO: Reconnect?
				return
			} else if err != nil {
				logger.Error("Server decode error:", err)
				return
			}
			logger.Debugf("Server: %s", msg)
			err = r.connected.Encode(msg)
			if err != nil {
				logger.Error("Sever encode error:", err)
				return
			}
		}
	}()

	return nil
}
