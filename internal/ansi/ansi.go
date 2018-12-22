package ansi

import (
	"regexp"
	"strings"
)

// var regex = regexp.MustCompile(`(\\(033|x1B)|)\[(.{1,3})m?`)
var regex = regexp.MustCompile(`(\\(033|x1B)|)(\d|\[((\w{1,3}(;\w{1,3})?))m?)`)

type (
	// Code represents all ansi codes.
	Code           string
	specialCode    = Code
	foregroundCode = Code
	backgroundCode = Code
)

const (
	Reset            specialCode = `0`
	Bold                         = `1`
	Italic                       = `3`
	Blink                        = `5`
	Underline                    = `4`
	UnderlineOff                 = `24`
	Inverse                      = `7`
	InverseOff                   = `27`
	Strikethrough                = `9`
	StrikethroughOff             = `29`
	EraseScreen                  = `2J`
	CursorHome                   = `H`
)

const (
	Default foregroundCode = `39`
	White                  = `37`
	Black                  = `30`
	Grey                   = `90`
	Red                    = `31`
	Green                  = `32`
	Blue                   = `34`
	Yellow                 = `33`
	Magenta                = `35`
	Cyan                   = `36`
)

const (
	DefaultBg backgroundCode = `49`
	WhiteBg                  = `47`
	BlackBg                  = `40`
	RedBg                    = `41`
	GreenBg                  = `42`
	BlueBg                   = `44`
	YellowBg                 = `43`
	MagentaBg                = `45`
	CyanBg                   = `46`
)

func findCodes(s string) []Code {
	slice := regex.FindAllStringSubmatch(s, -1)
	codes := make([]Code, 0, len(slice))
	for _, f := range slice {
		// Yeah so my regex sucks so for now we're just going to hack the hell
		// out of this until it works the way I want it to!!!!
		if len(f) != 7 {
			continue
		}

		codes = append(codes, Code(strings.TrimSuffix(f[5], "m")))
	}

	return codes
}
