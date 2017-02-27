package glas

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/ziutek/telnet"
)

func (e *entropy) handleAutoLogin(c *conf) error {
	for _, str := range c.Connect.AutoLogin {
		if str == _user {
			if c.Character.Name == "" {
				return errors.New("autologin not possible: c.Character.Name is not set")
			}

			if err := e.send(c.Character.Name); err != nil {
				return errors.Wrap(err, "e.send")
			}
		} else if str == _pass {
			if c.Character.Password == "" {
				return errors.New("autologin not possible: c.Character.Password is not set")
			}

			if err := e.send(c.Character.Password); err != nil {
				return errors.Wrap(err, "e.send")
			}
		} else {
			if err := e.send(str); err != nil {
				return errors.Wrap(err, "e.send")
			}
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (e *entropy) connect() error {
	var err error
	e.Conn, err = telnet.Dial("tcp", e.address)
	if err != nil {
		return errors.Wrap(err, "telnet.Dial")
	}

	if e._conf != nil {
		if e._conf.Connect.AutoLogin != nil {
			if err := e.handleAutoLogin(e._conf); err != nil {
				return errors.Wrap(err, "handleAutoLogin")
			}
		}

		func(c *conf) {
			e.aliasesMutex.Lock()
			defer e.aliasesMutex.Unlock()
			e._aliases = c.Aliases
		}(e._conf)
	}

	// Ensure that we only start our handleConnection thread once
	// this way if we disconnect/reconnect we don't have multiple
	// go routines attempting to read the connection.
	_connect := func() {
		go e.handleConnection()
	}
	var once sync.Once
	once.Do(_connect)

	return nil
}
