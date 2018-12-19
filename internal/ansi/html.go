package ansi

import (
	"strconv"
)

var colors map[Code][]byte

func init() {
	colors = make(map[Code][]byte)
	colors[White] = []byte(`#FFFFFF`)
	colors[Black] = []byte(`#000000`)
	colors[Grey] = []byte(`#808080`)
	colors[Red] = []byte(`#FF0000`)
	colors[Green] = []byte(`#008000`)
	colors[Blue] = []byte(`#0000FF`)
	colors[Yellow] = []byte(`#FFFF00`)
	colors[Magenta] = []byte(`#FF00FF`)
	colors[Cyan] = []byte(`#00FFFF`)
}

func ReplaceCodes(b []byte) []byte {
	return regex.ReplaceAllFunc(b, replacer)
}

func replacer(b []byte) []byte {
	fields := regex.FindStringSubmatch(string(b))
	n, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return b
	}

	color, ok := colors[Code(n)]
	if !ok {
		// Strip non-convertable codes
		return nil
	}

	return color
}
