package internal

import (
	"bytes"
	"fmt"
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
	crlfBuffer := [2]byte{'\r', '\n'}

	go func() {
		rbuf := make([]byte, 1024)
		for {
			nr, err := c.in.Read(rbuf)
			if nr > 0 {
				var buf bytes.Buffer
				buf.Write(rbuf[:nr])
				buf.Write(crlfBuffer[:])

				nw, ew := w.Write(buf.Bytes())
				if ew != nil {
					c.errCh <- errors.Wrap(err, "w.Write")
					return
				}

				if len(buf.Bytes()) != nw {
					c.errCh <- io.ErrShortWrite
					return
				}
				if err != nil {
					if err != io.EOF {
						c.errCh <- err
					}
					break
				}
			}
		}

		fmt.Println("exiting input loop")
	}()

	// Handler output.
	_, err := Copy(c.out, r)
	if err != nil {
		c.errCh <- errors.Wrap(err, "Copy")
	}
}
