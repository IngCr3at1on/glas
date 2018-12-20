package ansi

import (
	"fmt"
	"strconv"
	"strings"
)

var colors map[Code]string

const (
	placeholder = `$placeholder$`
	separator   = `$separator$`
	blank       = `$blank$`
)

func init() {
	colors = make(map[Code]string)
	// colors[White] = placeholder + `#FFFFFF` + separator
	colors[White] = blank
	colors[Black] = placeholder + `#000000` + separator
	colors[Grey] = placeholder + `#808080` + separator
	colors[Red] = placeholder + `#FF0000` + separator
	colors[Green] = placeholder + `#008000` + separator
	colors[Blue] = placeholder + `#0000FF` + separator
	colors[Yellow] = placeholder + `#FFFF00` + separator
	colors[Magenta] = placeholder + `#FF00FF` + separator
	colors[Cyan] = placeholder + `#00FFFF` + separator
}

func ReplaceCodes(byt []byte) []byte {
	_s := regex.ReplaceAllStringFunc(string(byt), replacer)
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

	return []byte(strings.Join(final, ""))
}

func replacer(s string) string {
	fields := regex.FindStringSubmatch(s)
	n, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return s
	}

	color, ok := colors[Code(n)]
	if !ok {
		// Strip non-convertable codes
		return ""
	}

	return color
}
