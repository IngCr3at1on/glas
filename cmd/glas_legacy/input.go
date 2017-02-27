package main

import (
	"fmt"

	"github.com/chzyer/readline"
)

func handleInput() {
	rl, err := readline.New("$ ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rl.Close()

	for {
		in, err := rl.Readline()
		errAndExit(err)

		iochan <- in
	}
}
