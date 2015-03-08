package main

import (
	"errors"
	"net"

	"github.com/dchest/uniuri"
	"github.com/sorcix/irc"
)

var ErrBufferExceeded = errors.New("buffer exceeded")
var ErrMismatchedPong = errors.New("mismatched pong")

const clientMsgBuffer = 10

type Client struct {
	net.Conn
	*irc.Encoder
	*irc.Decoder

	buffer chan *irc.Message
}

// NewClient returns a new Client
func NewClient(conn net.Conn) (*Client, error) {
	// XXX: Debugging here
	//conn = LogConn(conn)

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

func (c *Client) searchBuffer(command string) (message *irc.Message, err error) {
	buffer := make(chan *irc.Message, cap(c.buffer))

Fill:
	for {
		select {
		case message = <-c.buffer:
			if message.Command == command {
				break Fill
			}
			// Put message back in our temporary buffer
			select {
			case buffer <- message:
			default:
				err = ErrBufferExceeded
			}
		default:
			break Fill
		}
	}

	close(buffer)

	for m := range buffer {
		select {
		case c.buffer <- m:
		default:
			err = ErrBufferExceeded
			return
		}
	}
	return
}

// DecodeWhen will buffer messages until the desired command is seen
func (c *Client) DecodeWhen(command string) (*irc.Message, error) {
	message, err := c.searchBuffer(command)
	if err != nil {
		return message, err
	}
	if message != nil {
		return message, nil
	}

	// FIXME: This is shitty, rewrite the buffer to not use a chan.
	//   There is a race condition if message is buffered between here.
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

func (c *Client) Ping() error {
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
	return nil
}
