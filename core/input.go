package core

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

var (
	ansi []string

	special = []string{
		"\033[0m", "\033[1m", "\033[3m", "\033[5m", "\033[4m", "\033[24m",
		"\033[7m", "\033[27m", "\033[9m", "\033[29m",
	}

	fg = []string{
		"\033[39m", "\033[37m", "\033[30m", "\x1B[90m", "\033[31m", "\033[32m",
		"\033[34m", "\033[33m", "\033[35m", "\033[36m",
	}

	bg = []string{
		"\033[49m", "\033[47m", "\033[40m", "\033[41m", "\033[42m", "\033[44m",
		"\033[43m", "\033[45m", "\033[46m",
	}
)

func init() {
	ansi = append(ansi, special...)
	ansi = append(ansi, fg...)
	ansi = append(ansi, bg...)
}

func stripAnsi(data string, set []string) string {
	for _, ac := range set {
		if strings.Contains(data, ac) {
			data = strings.Replace(data, ac, "", -1)
		}
	}

	return data
}

/*
func (e *entropy) handleKeyPress(input string) error {
	switch input {
	case ""
	}
}
*/

func (e *entropy) handleInput(quit chan struct{}) {
	rl, err := readline.New("$ ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rl.Close()

	for {
		select {
		case <-quit:
			return
		default:
			/*
				r := rl.Terminal.ReadRune()
				fmt.Println(r)
			*/
			in, err := rl.Readline()
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			if in == "" {
				fmt.Println("")
				break
			}

			if err := e.handleCommand(in, quit); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}
}
