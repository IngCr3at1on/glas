package glas

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/ziutek/telnet"
)

func (g *glas) handleAutoLogin(c *conf) error {
	for _, str := range c.Connect.AutoLogin {
		if str == _user {
			if c.Character.Name == "" {
				return errors.New("autologin not possible: c.Character.Name is not set")
			}

			if err := g.send(c.Character.Name); err != nil {
				return errors.Wrap(err, "g.send")
			}
		} else if str == _pass {
			if c.Character.Password == "" {
				return errors.New("autologin not possible: c.Character.Password is not set")
			}

			if err := g.send(c.Character.Password); err != nil {
				return errors.Wrap(err, "g.send")
			}
		} else {
			if err := g.send(str); err != nil {
				return errors.Wrap(err, "g.send")
			}
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (g *glas) connect() error {
	var err error
	g.Conn, err = telnet.Dial("tcp", g.address)
	if err != nil {
		return errors.Wrap(err, "telnet.Dial")
	}

	if g._conf != nil {
		if g._conf.Connect.AutoLogin != nil {
			if err := g.handleAutoLogin(g._conf); err != nil {
				return errors.Wrap(err, "handleAutoLogin")
			}
		}

		func(c *conf) {
			g.aliasesMutex.Lock()
			defer g.aliasesMutex.Unlock()
			g._aliases = c.Aliases
		}(g._conf)
	}

	// Ensure that we only start our handleConnection thread once
	// this way if we disconnect/reconnect we don't have multiple
	// go routines attempting to read the connection.
	_connect := func() {
		go g.handleConnection()
	}
	var once sync.Once
	once.Do(_connect)

	return nil
}
