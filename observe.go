package glas

import (
	"fmt"

	"github.com/IngCr3at1on/glas/ansi"
)

func (g *glas) observe(data string) error {
	// Strip out the background color for printing.
	// TODO possibly control this by a setting?
	fmt.Fprintln(g.ioout, ansi.Strip(data, ansi.Bg))

	// Strip out all ansi codes for matching (used in triggers)
	//data = ansi.Strip(data, ansi.Codes)

	return nil
}
