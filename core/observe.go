package core

import (
	"fmt"
	"strings"
	"unicode"
)

func (e *entropy) observe(data string) error {
	data = strings.TrimFunc(data, func(c rune) bool {
		return unicode.IsSymbol(c)
		//return unicode.Is(unicode.ASCII_Hex_Digit, c)
		//return c == '\u001b' || c == '\u003e' || unicode.IsSpace(c)
	})
	//data = strings.TrimSpace(strings.TrimPrefix("\u001b[37m\u001b[40m\u003e", data))

	//os.Stdout.WriteString(fmt.Sprintf("%+v\n", data))

	fmt.Print(data)
	return nil
}
