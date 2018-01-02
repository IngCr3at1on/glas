package glas

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (
	// Glas is our mud client.
	Glas struct {
		log *logrus.Entry

		out             io.Writer
		characterConfig *CharacterConfig
		conn            net.Conn

		errCh  chan error
		stopCh chan error
	}
)

// New returns a new instance of Glas.
func New(characterConfig *CharacterConfig, out io.Writer, errCh, stopCh chan error, log *logrus.Entry) (*Glas, error) {
	if characterConfig == nil {
		return nil, errors.New("config cannot be nil")
	}

	if out == nil {
		return nil, errors.New("out cannot be nil")
	}

	if errCh == nil {
		return nil, errors.New("errCh cannot be nil")
	}

	if stopCh == nil {
		return nil, errors.New("stopCh cannot be nil")
	}

	if err := characterConfig.Validate(); err != nil {
		return nil, errors.Wrap(err, "characterConfig.Validate")
	}

	if log == nil {
		log = logrus.NewEntry(logrus.New())
	}

	return &Glas{
		log:             log,
		out:             out,
		characterConfig: characterConfig,
		errCh:           errCh,
		stopCh:          stopCh,
	}, nil
}

// Start starts our mud client.
func (g *Glas) Start() {
	if err := g.connect(); err != nil {
		g.errCh <- err
		return
	}
}

// Send data to the mud.
func (g *Glas) Send(data ...interface{}) error {
	g.log.WithFields(logrus.Fields{
		"command": "Send",
		"data":    data,
	}).Debug("Called")

	for i, d := range data {
		switch d.(type) {
		case []byte:
			byt := d.([]byte)
			byt = append(byt, '\n')
			if _, err := g.conn.Write(byt); err != nil {
				return err
			}
		case string:
			str := d.(string)

			ok, err := g.characterConfig.aliases.maybeHandleAlias(g, str)
			if err != nil {
				return err
			}

			if !ok {
				if _, err := g.conn.Write([]byte(fmt.Sprintf("%s\n", str))); err != nil {
					return err
				}
			}
		default:
			return errors.New("Invalid data type")
		}

		if i+1 < len(data) {
			// TODO: make this configurable.
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}
