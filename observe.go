package glas

import (
	"fmt"

	"github.com/IngCr3at1on/glas/ansi"
	"github.com/sirupsen/logrus"
)

func (g *Glas) observe(data string) error {
	g.log.WithFields(logrus.Fields{
		"command": "observe",
		"data":    data,
	}).Debug("Called")

	// Strip out background color for printing.
	// TODO: control this with a setting.
	data = ansi.Strip(data, ansi.Bg)

	// Strip out all ansi codes for matching (used in triggers)
	//data = ansi.Strip(data, ansi.Codes)

	fmt.Fprint(g.out, data)

	return nil
}
