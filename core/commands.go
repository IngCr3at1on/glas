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

func (e *entropy) handleCommand(input string, quit chan struct{}) error {
	if strings.HasPrefix(input, "/") {

		input = strings.TrimFunc(input, func(c rune) bool {
			// Trim off any unexpected input
			return unicode.IsSpace(c) || !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsSymbol(c) && c != '*'
		})

		switch {
		case strings.HasPrefix(input, "add"), strings.HasPrefix(input, "set"):
			input = strings.TrimPrefix(input, "add ")
			input = strings.TrimPrefix(input, "set ")

			if strings.HasPrefix(input, "alias ") {
				e.newAlias(strings.TrimPrefix(input, "alias "))
			}
		case strings.HasPrefix(input, "alias"):
			e.newAlias(strings.TrimPrefix(input, "alias "))
		case strings.Compare(input, "connect") == 0:
			if err := e.connect(quit); err != nil {
				return errors.Wrap(err, "connect")
			}
		case strings.Compare(input, "quit") == 0:
			// TODO delayed shutdown to make sure all go routines stop?
			close(quit)
		default:
			fmt.Println("Unkown command")
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
