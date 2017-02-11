package core

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

type (
	chain []string

	alias struct {
		Action chain `json:"action"`
	}

	aliases map[string]*alias
)

func (e *entropy) handleChain(c chain) error {
	for _, str := range c {
		if err := e.send(str); err != nil {
			return errors.Wrap(err, "e.send")
		}

		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (e *entropy) handleAutoLogin(c *character) error {
	if c.Name == "" {
		return errors.New("autologin not possible: c.Name is not set")
	}

	if c.Password == nil {
		return errors.New("autologin not possible: c.Password is not set")
	}

	for _, str := range c.AutoLogin {
		if str == _user {
			if err := e.send(c.Name); err != nil {
				return errors.Wrap(err, "e.send")
			}
		} else if str == _pass {
			if err := e.send(c.Password); err != nil {
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

func (e *entropy) maybeHandleAlias(input string) (bool, error) {
	e.aliasesMutex.Lock()
	defer e.aliasesMutex.Unlock()

	c, ok := e._aliases[input]
	if !ok {
		return false, nil
	}

	if err := e.handleChain(c.Action); err != nil {
		return false, errors.Wrap(err, "handleChain")
	}

	return true, nil
}

func (e *entropy) handleCommand(input string, quit chan struct{}) error {
	if strings.HasPrefix(input, "/") {
		input = strings.TrimFunc(input, func(c rune) bool {
			return !unicode.IsLetter(c)
		})

		/*
			if strings.Compare(input, "here") == 0 {
				os.Stdout.WriteString(fmt.Sprintf("%d\n", e.here))
			}

			if strings.Compare(input, "wander") == 0 {
				os.Stdout.WriteString("wander enabled\n")
				e._wander = true
			}

			if strings.Compare(input, "stop") == 0 {
				os.Stdout.WriteString("wander stopped\n")
				e._wander = false
			}

			if strings.Compare(input, "goto sewers") == 0 {
				if err := e.handleScript(goToSewers()); err != nil {
					return err
				}
			}
		*/
		if strings.Compare(input, "quit") == 0 {
			close(quit)
		}

		return nil
	}

	b, err := e.maybeHandleAlias(input)
	if err != nil {
		return errors.Wrap(err, "maybeHandleAlias")
	}

	if !b {
		if _, err := e.Conn.Write([]byte(fmt.Sprintf("%s\n", input))); err != nil {
			return err
		}
	}

	return nil
}
