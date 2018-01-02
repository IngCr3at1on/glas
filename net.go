package glas

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	_user = "$user"
	_pass = "$pass"
)

func (g *Glas) connect() (err error) {
	g.conn, err = net.Dial("tcp", g.characterConfig.Address)
	if err != nil {
		return errors.Wrap(err, "net.Dial")
	}

	time.Sleep(time.Millisecond * 100)

	if g.characterConfig.AutoLogin != nil {
		if err := g.handleAutoLogin(); err != nil {
			return errors.Wrap(err, "handleAutoLogin")
		}
	}

	// Ensure that we only start out handleConnection thread once this way if
	// we disconnect/reconnect we don't have multiple go-routines attempting
	// to read the connection.
	_connect := func() {
		go g.handleConnection()
	}
	var once sync.Once
	once.Do(_connect)

	return nil
}

func (g *Glas) handleAutoLogin() error {
	for _, str := range g.characterConfig.AutoLogin {
		switch str {
		case _user:
			if g.characterConfig.Name == "" {
				return errors.New("autologin not possible: character name not set")
			}

			if err := g.Send(g.characterConfig.Name); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		case _pass:
			if g.characterConfig.Password == "" {
				return errors.New("autologin not possible: character password not set")
			}

			// TODO: decode/decrypt this from whatever value it's in (something other
			// than plain text or what's the point)...

			if err := g.Send(g.characterConfig.Password); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		default:
			if err := g.Send(str); err != nil {
				return errors.Wrap(err, "g.Send")
			}
		}

		// TODO: make this configurable.
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (g *Glas) handleConnection() {
	tag := "handleConnection"

	rd := bufio.NewReader(g.conn)

	for {
		select {
		case <-g.stopCh:
			break
		default:
			// FIXME: This doesn't quite work in all situations (zebedee login for example)...
			in, err := rd.ReadString('\n')
			if err != nil {
				g.errCh <- errors.Wrap(err, tag)
				break
			}

			in = strings.TrimFunc(in, func(c rune) bool {
				return c == '\r' || c == '\n'
			})

			if err := g.observe(in); err != nil {
				g.errCh <- errors.Wrap(err, tag)
				break
			}
		}
	}
}
