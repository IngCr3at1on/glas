package internal

import (
	"bytes"
	"io"
	"regexp"
	"strings"

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

var regex = regexp.MustCompile(`(\\(033|x1b)|)`)

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
	go c.write(w)
	c.read(r)
}

func (c *Caller) write(w telnet.Writer) {
	crlfBuffer := [2]byte{'\r', '\n'}
	var wbuf bytes.Buffer
	rbuf := make([]byte, 1024)

	for {
		nr, er := c.in.Read(rbuf)
		if nr > 0 {
			wbuf.Write(rbuf[:nr])
			wbuf.Write(crlfBuffer[:])

			nw, err := w.Write(wbuf.Bytes())
			if err != nil {
				c.errCh <- err
				return
			}

			if nw != wbuf.Len() {
				c.errCh <- io.ErrShortWrite
				return
			}

			wbuf.Reset()
		}
		if er != nil {
			if er != io.EOF {
				c.errCh <- er
			}
			break
		}
	}
}

func (c *Caller) read(r telnet.Reader) {
	rbuf := make([]byte, 1)
	var wbuf bytes.Buffer
	readLine := false

	for {
		nr, er := r.Read(rbuf)
		if nr > 0 {
			wbuf.Write(rbuf[:nr])
			if regex.Match(wbuf.Bytes()) {
				readLine = true
			}

			if readLine && strings.Contains(wbuf.String(), "\r\n") || !readLine {
				nw, err := c.out.Write(wbuf.Bytes())
				if err != nil {
					c.errCh <- err
					return
				}

				if nw != wbuf.Len() {
					c.errCh <- io.ErrShortWrite
					return
				}

				wbuf.Reset()
				readLine = false
			}
		}
		if er != nil {
			if er != io.EOF {
				c.errCh <- er
			}
			break
		}
	}
}
