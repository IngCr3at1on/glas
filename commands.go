package glas

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	helpMsg = `Glas known commands:
connect     connect to an address or use a character config to connect
characters  lists any loaded characters you have by name
help        shows this message
quit        quits Glas`

	connectUsageMsg = `connect <address|character name>`
)

func (g *Glas) handleCommand(input string) (bool, error) {
	if strings.HasPrefix(input, g.config.CmdPrefix) {
		input = strings.TrimFunc(input, func(c rune) bool {
			// Trim off any unexpected input
			return unicode.IsSpace(c) || !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsSymbol(c) && c != '*'
		})

		if prefix := "help"; strings.HasPrefix(input, prefix) {
			input = strings.TrimSpace(strings.TrimPrefix(input, prefix))

			if input == "" {
				fmt.Fprint(g.out, helpMsg)
			}

			if strings.Compare(input, "connect") == 0 {
				fmt.Fprint(g.out, "Connect to a specified address or using a selected character configuration")
			}

			if strings.Compare(input, "characters") == 0 {
				fmt.Fprint(g.out, "List characters that are loaded into glas their name.")
			}

			return true, nil
		}

		if strings.Compare(input, "quit") == 0 {
			g.stopCh <- errors.New("quit called")
			return true, nil
		}

		if prefix := "connect"; strings.HasPrefix(input, prefix) {
			input = strings.TrimSpace(strings.TrimPrefix(input, prefix))
			if input == "" {
				// FIXME: if a character has already been selected or an address already
				// passed in previously, use it now instead of erroring.
				fmt.Fprintf(g.out, connectUsageMsg)
				return true, nil
			}

			if err := g.connect(input); err != nil {
				return true, err
			}

			return true, nil
		}

		if strings.Compare(input, "characters") == 0 {
			fmt.Fprint(g.out, "Characters:")
			for _, c := range g.characters.getCharacters() {
				fmt.Fprint(g.out, c.Name)
			}

			return true, nil
		}
	}

	return false, nil
}
