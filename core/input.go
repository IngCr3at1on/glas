package core

import (
	"fmt"

	"github.com/chzyer/readline"
)

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
