package ansi

import "strings"

var (
	// Codes contains all the known ansi codes.
	Codes map[string]string
	// Special contains special ansi codes.
	Special map[string]string
	// Fg contains foreground ansi codes.
	Fg map[string]string
	// Bg contains background ansi codes.
	Bg map[string]string
)

// Ansi color code identifiers.
const (
	Reset            = "reset"
	Bold             = "bold"
	Italic           = "italic"
	Blink            = "blink"
	Underline        = "underline"
	UnderlineOff     = "underline_off"
	Inverse          = "inverse"
	InverseOff       = "inverse_off"
	Strikethrough    = "strikethrough"
	StrikethroughOff = "strikethrough_off"

	Default = "default"
	White   = "white"
	Black   = "black"
	Grey    = "grey"
	Red     = "red"
	Green   = "green"
	Blue    = "blue"
	Yellow  = "yellow"
	Magenta = "magenta"
	Cyan    = "cyan"

	DefaultBg = "default_bg"
	WhiteBg   = "white_bg"
	BlackBg   = "black_bg"
	RedBg     = "red_bg"
	GreenBg   = "green_bg"
	BlueBg    = "blue_bg"
	YellowBg  = "yellow_bg"
	MagentaBg = "magenta_bg"
	CyanBg    = "cyan_bg"
)

func init() {
	Special = make(map[string]string)
	Special[Reset] = "\033[0m"
	Special[Bold] = "\033[1m"
	Special[Italic] = "\033[3m"
	Special[Blink] = "\033[5m"
	Special[Underline] = "\033[4m"
	Special[UnderlineOff] = "\033[24m"
	Special[Inverse] = "\033[7m"
	Special[InverseOff] = "\033[27m"
	Special[Strikethrough] = "\033[9m"
	Special[StrikethroughOff] = "\033[29m"

	Fg = make(map[string]string)
	Fg[Default] = "\033[39m"
	Fg[White] = "\033[37m"
	Fg[Black] = "\033[30m"
	Fg[Grey] = "\x1B[90m"
	Fg[Red] = "\033[31m"
	Fg[Green] = "\033[32m"
	Fg[Blue] = "\033[34m"
	Fg[Yellow] = "\033[33m"
	Fg[Magenta] = "\033[35m"
	Fg[Cyan] = "\033[36m"

	Bg = make(map[string]string)
	Bg[DefaultBg] = "\033[49m"
	Bg[WhiteBg] = "\033[47m"
	Bg[BlackBg] = "\033[40m"
	Bg[RedBg] = "\033[41m"
	Bg[GreenBg] = "\033[42m"
	Bg[BlueBg] = "\033[44m"
	Bg[YellowBg] = "\033[43m"
	Bg[MagentaBg] = "\033[45m"
	Bg[CyanBg] = "\033[46m"

	Codes = make(map[string]string)
	appendMap := func(m map[string]string) {
		for k, v := range m {
			Codes[k] = v
		}
	}
	appendMap(Special)
	appendMap(Fg)
	appendMap(Bg)
}

// Strip accepts a string and an ansi code set and returns the stripped string.
func Strip(data string, set map[string]string) string {
	for _, ac := range set {
		if strings.Contains(data, ac) {
			data = strings.Replace(data, ac, "", -1)
		}
	}

	return data
}
