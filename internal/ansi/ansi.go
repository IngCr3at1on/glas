package ansi

import (
	"regexp"
	"strconv"
)

var regex = regexp.MustCompile(`(\\(033|x1B)|)\[(\d{1,2})m`)

type (
	// Code represents all ansi codes.
	Code           uint32
	specialCode    = Code
	foregroundCode = Code
	backgroundCode = Code
)

const (
	Reset            specialCode = 0
	Bold                         = 1
	Italic                       = 3
	Blink                        = 5
	Underline                    = 4
	UnderlineOff                 = 24
	Inverse                      = 7
	InverseOff                   = 27
	Strikethrough                = 9
	StrikethroughOff             = 29
)

const (
	Default foregroundCode = 39
	White                  = 37
	Black                  = 30
	Grey                   = 90
	Red                    = 31
	Green                  = 32
	Blue                   = 34
	Yellow                 = 33
	Magenta                = 35
	Cyan                   = 36
)

const (
	DefaultBg backgroundCode = 49
	WhiteBg                  = 47
	BlackBg                  = 40
	RedBg                    = 41
	GreenBg                  = 42
	BlueBg                   = 44
	YellowBg                 = 43
	MagentaBg                = 45
	CyanBg                   = 46
)

func findCodes(s string) []Code {
	slice := regex.FindAllStringSubmatch(s, -1)
	codes := make([]Code, 0, len(slice))
	for _, f := range slice {
		n, err := strconv.Atoi(f[len(f)-1])
		if err == nil {
			codes = append(codes, Code(n))
		}
	}

	return codes
}
