package internal

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"
)

const (
	_user = "$user"
	_pass = "$pass"
)

type (
	// Conn is a mud connection.
	Conn struct {
		Conn net.Conn

		connected bool
	}
)

// Dial a mud connection.
func Dial(ctx context.Context, addr string) (*Conn, error) {
	// FIXME: use this context for a timeout (and cancel), requires addressing issue in glas.Connect first.
	if addr == "" {
		return nil, errors.New("addr cannot be empty")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:      conn,
		connected: true,
	}, nil
}

// Listen on a mud connection writing back to w.
func (c *Conn) Listen(ctx context.Context, w io.Writer) error {
	defer c.Conn.Close()

out:
	for {
		select {
		case <-ctx.Done():
			break out
		default:
			if !c.connected {
				break out
			}

			if _, err := _copy(w, c.Conn); err != nil {
				c.connected = false
				return errors.Wrap(err, "_copy")
			}
		}
	}

	return nil
}
