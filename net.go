package glas

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	_user = "$user"
	_pass = "$pass"
)

type (
	conn struct {
		net.Conn

		connected bool
		glas      *Glas
	}
)

func (g *Glas) connect(input string) error {
	if g.conn.connected {
		return nil
	}

	// We were passed a file that may or may not already be loaded into our
	// character map.
	if strings.HasSuffix(input, ".toml") {
		name, err := g.loadCharacterConfig(input)
		if err != nil {
			return errors.Wrap(err, "g.loadCharacterConfig")
		}

		input = name
	}

	cc := g.characters.getCharacter(input)
	if cc != nil {
		input = cc.Address
	}

	if err := g.conn.connect(input, cc); err != nil {
		return err
	}

	g.currentCharacter = cc
	return nil
}

func (c *conn) connect(address string, cc *CharacterConfig) (err error) {
	if c.connected {
		return nil
	}

	c.Conn, err = net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.connected = true

	time.Sleep(time.Millisecond * 100)

	if cc != nil && cc.AutoLogin != nil {
		if err := c.handleAutoLogin(cc); err != nil {
			return errors.Wrap(err, "handleAutoLogin")
		}
	}

	go c.handleConnection()

	return nil
}

func (c *conn) disconnect() {
	if !c.connected {
		return
	}

	_ = c.Conn.Close()
	c.connected = false
}

func (c *conn) handleAutoLogin(cc *CharacterConfig) error {
	for _, str := range cc.AutoLogin {
		switch str {
		case _user:
			if cc.Name == "" {
				return errors.New("autologin not possible: character name not set")
			}

			if err := c.glas.Send(cc.Name); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		case _pass:
			if cc.Password == "" {
				return errors.New("autologin not possible: character password not set")
			}

			// TODO: decode/decrypt this from whatever value it's in (something other
			// than plain text or what's the point)...

			if err := c.glas.Send(cc.Password); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		default:
			if err := c.glas.Send(str); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		}

		// TODO: make this configurable.
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (c *conn) handleConnection() {
	tag := "handleConnection"

	rd := bufio.NewReader(c.Conn)

	for {
		select {
		case <-c.glas.stopCh:
			return
		default:
			if !c.connected {
				return
			}

			// FIXME: This doesn't quite work in all situations (zebedee login for example)...
			in, err := rd.ReadString('\n')
			if err != nil {
				c.glas.errCh <- errors.Wrap(err, tag)
				return
			}

			in = strings.TrimFunc(in, func(c rune) bool {
				return c == '\r' || c == '\n'
			})

			if err := c.glas.observe(in); err != nil {
				c.glas.errCh <- errors.Wrap(err, tag)
				return
			}
		}
	}
}
