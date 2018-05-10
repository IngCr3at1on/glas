package internal

import (
	"io"
	"strings"

	"github.com/IngCr3at1on/glas/ansi"
)

// A version of io.Copy that we can strap our own functionality over...
func _copy(dst io.Writer, src io.Reader) (written int64, err error) {
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}

	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			str := string(buf[0:nr])
			str = strings.TrimFunc(str, func(c rune) bool {
				return c == '\r' || c == '\n'
			})
			// Strip out background color for printing.
			// TODO: control this with a setting.
			str = ansi.Strip(str, ansi.Bg)

			// Strip out all ansi codes for matching (used in triggers)
			// data = ansi.Strip(data, ansi.Codes)

			nw, ew := dst.Write([]byte(str))
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			// We stripped some bytes off so we can't use nr for our write check.
			if len(str) != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
