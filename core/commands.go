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
	for _, str := range c.AutoLogin {
		if str == _user {
			if c.Name == "" {
				return errors.New("autologin not possible: c.Name is not set")
			}

			if err := e.send(c.Name); err != nil {
				return errors.Wrap(err, "e.send")
			}
		} else if str == _pass {
			if c.Password == nil {
				return errors.New("autologin not possible: c.Password is not set")
			}

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

	fields := strings.Fields(input)
	c, ok := e._aliases[fields[0]]
	if !ok {
		return false, nil
	}

	action := chain{}
	for _, str := range c.Action {
		// Ignores extra input fields beyond the number of wildcards in
		// the match string.
		// Should also cut off extra wildcards if an input field is not
		// provided for it...
		if strings.Contains(str, "%s") {
			slice := strings.Fields(str)
			if len(fields) > 1 {
				for i := range fields {
					if i+1 > len(slice) {
						break
					}
					if i == 0 {
						continue
					}
					str = fmt.Sprintf(str, fields[i])
					action = append(action, str)
				}
			} else {
				str = strings.Fields(str)[0]
				action = append(action, str)
			}
		}
	}

	if err := e.handleChain(action); err != nil {
		return false, errors.Wrap(err, "handleChain")
	}

	return true, nil
}

func (e *entropy) handleCommand(input string, quit chan struct{}) error {
	if strings.HasPrefix(input, "/") {
		input = strings.TrimFunc(input, func(c rune) bool {
			return !unicode.IsLetter(c)
		})

		if strings.Compare(input, "connect") == 0 {
			if err := e.connect(quit); err != nil {
				return errors.Wrap(err, "connect")
			}
		}

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
