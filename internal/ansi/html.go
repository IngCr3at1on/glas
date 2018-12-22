package ansi

import (
	"fmt"
	"strings"
)

var codes map[Code]string

const (
	// Instruction is prepended to instruction codes.
	Instruction = `$instruction$`
	// Separator is used to separating instructions from normal data.
	Separator = `$separator$`

	placeholder = `$placeholder$`
	separator   = `$separator2$`
	blank       = `$blank$`
)

func init() {
	codes = make(map[Code]string)
	// codes[White] = placeholder + `#FFFFFF` + separator
	codes[White] = blank
	codes[Black] = placeholder + `#000000` + separator
	codes[Grey] = placeholder + `#808080` + separator
	codes[Red] = placeholder + `#FF0000` + separator
	codes[Green] = placeholder + `#008000` + separator
	codes[Blue] = placeholder + `#0000FF` + separator
	codes[Yellow] = placeholder + `#FFFF00` + separator
	codes[Magenta] = placeholder + `#FF00FF` + separator
	codes[Cyan] = placeholder + `#00FFFF` + separator
	codes[EraseScreen] = placeholder + Instruction + `ERASESCREEN` + Separator
}

func ReplaceCodes(s string) string {
	_s := regex.ReplaceAllStringFunc(s, replacer)
	_fields := strings.Split(_s, placeholder)
	var fields []string
	for _, f := range _fields {
		fields = append(fields, strings.Split(f, blank)...)
	}

	final := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.Contains(f, separator) {
			f = strings.Replace(f, separator, `;">`, -1)
			var b strings.Builder
			const suffix = "\r\n"
			fmt.Fprint(&b, `<span style="color:`, strings.TrimSuffix(f, suffix), `</span>`)
			if strings.HasSuffix(f, suffix) {
				b.WriteString(suffix)
			}

			f = b.String()
		}

		final = append(final, f)
	}

	return strings.Join(final, "")
}

func replacer(s string) string {
	fields := regex.FindStringSubmatch(s)

	// See comment in ansi.go...
	if len(fields) != 7 {
		return s
	}

	code, ok := codes[Code(strings.TrimSuffix(fields[5], "m"))]
	if !ok {
		// Strip non-convertable codes
		return ""
	}

	return code
}
