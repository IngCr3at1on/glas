package internal

import (
	"io"
)

// Copy is a variant of io.Copy designed for our purposes.
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	// A small buffer means many iterations but also that we don't
	// have to wait for it to fill.
	buf := make([]byte, 1)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
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
