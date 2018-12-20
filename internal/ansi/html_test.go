package ansi

import (
	"testing"

	"github.com/ingcr3at1on/glas/internal/test"
)

func TestHTML(t *testing.T) {
	testCases := []struct {
		d string
		e string
	}{
		{`no ansi codes are set`, `no ansi codes are set`},
		{`we pass a reset code\033[0m`, `we pass a reset code`},
		{`\033[40m\033[37mblack background with white text`, `black background with white text`},
		{`\033[40mblack background\033[32mgreen text \033[37mwhite text`, `black background<font color=#008000>green text </font>white text`},
	}

	for _, tc := range testCases {
		a := ReplaceCodes([]byte(tc.d))
		test.Equals(t, tc.e, string(a))
	}
}
