package ansi

import (
	"testing"

	"github.com/ingcr3at1on/glas/internal/test"
)

func TestAnsi(t *testing.T) {
	testCases := []struct {
		d string
		e []Code
	}{
		{`no ansi codes are set`, nil},
		{`we pass a reset code\033[0m`, []Code{Reset}},
		{`\033[40m\033[37mblack background with white text`, []Code{BlackBg, White}},
		{`\033[40mblack background\033[32mgreen text\033[37mwhite text`, []Code{BlackBg, Green, White}},
	}

	for _, tc := range testCases {
		// fmt.Println("test case", n+1)
		compareSlice(t, tc.e, findCodes(tc.d))
	}
}

func compareSlice(tb testing.TB, expected, actual []Code) {
	test.Equals(tb, len(expected), len(actual))
	// Cool thing here is these should always be in the same order!
	for n := range expected {
		test.Equals(tb, expected[n], actual[n])
	}
}
