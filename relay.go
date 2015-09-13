package main

import (
	"fmt"
	"io"
	"sync"

	"github.com/sorcix/irc"
)

const ServerName = "neobnc"

type User struct {
	Nick string
	User string
	Host string
	Name string
}

type Relay struct {
	sync.Mutex
	user User
	// TODO: Handle many clients
	connected *Client
	// TODO: Handle many servers
	server *irc.Conn
	prefix *irc.Prefix
}

func (r *Relay) Join(c *Client) error {
	message, err := c.DecodeWhen(irc.NICK)
	if err != nil {
		return err
	}
	u := User{}
	u.Nick = message.Params[0]

	message, err = c.DecodeWhen(irc.USER)
	if err != nil {
		return err
	}
	u.User = message.Params[0]
	u.Name = message.Trailing

	r.Lock()
	defer r.Unlock()

	r.user = u
	r.connected = c
	logger.Infof("Client joined: %s", r.prefix)

	r.prefix = &irc.Prefix{Name: ServerName}
	err = c.Encode(&irc.Message{
		Prefix:   r.prefix,
		Command:  irc.RPL_WELCOME,
		Params:   []string{u.User},
		Trailing: fmt.Sprintf("Welcome!"),
	})
	if err != nil {
		return err
	}

	if r.server == nil {
		// XXX: Unhardcode this
		//r.Connect("localhost:6667")
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
			if r.server == nil {
				// Skip relay
				// TODO: Buffer these to relay later? Or some subset of commands?
			} else if err = r.server.Encode(msg); err != nil {
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
		Params:   []string{r.user.Nick, "0", "*"},
		Trailing: r.user.Name,
	})
	conn.Encode(&irc.Message{
		Command: irc.NICK,
		Params:  []string{r.user.Nick},
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
