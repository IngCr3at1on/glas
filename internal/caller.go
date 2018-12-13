package internal

import (
	"io"

	"github.com/pkg/errors"
	telnet "github.com/reiver/go-telnet"
)

type (
	// CallerConfig is a Caller configuration.
	CallerConfig struct {
		Out io.Writer
		In  io.Reader
	}

	// Caller implements the telnet.Caller interface.
	Caller struct {
		out   io.Writer
		in    io.Reader
		errCh chan error
	}
)

// NewCaller returns a new Caller.
func NewCaller(c *CallerConfig, errCh chan error) (*Caller, error) {
	if c.Out == nil {
		return nil, errors.New("c.Out cannot be nil")
	}

	if c.In == nil {
		return nil, errors.New("c.In cannot be nil")
	}

	if errCh == nil {
		return nil, errors.New("errCh cannot be nil")
	}

	return &Caller{
		out:   c.Out,
		in:    c.In,
		errCh: errCh,
	}, nil
}

// CallTELNET is called by the telnet client.
func (c *Caller) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	go func() {
		// Large buffer, pretty sure larger than any mud can support anyway lol,
		// terminate all copied data with \r\n.
		_, err := copy(w, c.in, 1024, true)
		if err != nil {
			c.errCh <- errors.Wrap(err, "copy input")
			return
		}
	}()

	// Handle output. A small buffer means many iterations but also that we
	// don't have to wait for it to fill.
	_, err := copy(c.out, r, 1, false)
	if err != nil {
		c.errCh <- errors.Wrap(err, "copy output")
	}
}
