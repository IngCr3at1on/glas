package internal

import (
	"bytes"
	"io"
)

// Copy is a variant of io.Copy designed for our purposes.
func Copy(dst io.Writer, src io.Reader, bufferSize uint, term bool) (written int64, err error) {
	crlfBuffer := [2]byte{'\r', '\n'}

	rbuf := make([]byte, bufferSize)
	for {
		nr, er := src.Read(rbuf)
		if nr > 0 {
			var buf bytes.Buffer
			buf.Write(rbuf[:nr])
			if term {
				buf.Write(crlfBuffer[:])
			}

			nw, ew := dst.Write(buf.Bytes())
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if len(buf.Bytes()) != nw {
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
