package main

import (
	"errors"
	"net"

	"github.com/sorcix/irc"
)

var ErrBufferExceeded = errors.New("buffer exceeded")

const clientMsgBuffer = 10

type Client struct {
	net.Conn
	*irc.Encoder
	*irc.Decoder
	buffer chan *irc.Message
}

// NewClient returns a new Client
func NewClient(conn net.Conn) (*Client, error) {
	return &Client{
		Conn:    conn,
		Encoder: irc.NewEncoder(conn),
		Decoder: irc.NewDecoder(conn),
		buffer:  make(chan *irc.Message, clientMsgBuffer),
	}, nil
}

// Decode will decode the next message, including buffered messages
func (c *Client) Decode() (*irc.Message, error) {
	select {
	case m := <-c.buffer:
		return m, nil
	default:
		return c.Decoder.Decode()
	}
}

// DecodeWhen will buffer messages until the desired command is seen
func (c *Client) DecodeWhen(command string) (*irc.Message, error) {
	for {
		message, err := c.Decoder.Decode()
		if err != nil {
			return nil, err
		}
		if message.Command == command {
			return message, nil
		}
		select {
		case c.buffer <- message:
		default:
			return nil, ErrBufferExceeded
		}
	}
}
