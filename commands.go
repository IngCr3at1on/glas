package glas

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

func (e *entropy) handleCommand(input string) error {
	// TODO allow setting this to a different color then normal text
	// also allow this to be disabled
	fmt.Fprintln(e.ioout, input)

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
			if err := e.connect(); err != nil {
				return errors.Wrap(err, "connect")
			}
		case strings.Compare(input, "quit") == 0:
			// TODO delayed shutdown to make sure all go routines stop?
			close(e._quit)
		default:
			fmt.Fprintln(e.ioout, "Unknown command")
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
