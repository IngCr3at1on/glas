package glas

import (
	"fmt"
	"strings"
	"sync"

	"io"

	"github.com/pkg/errors"
	"github.com/ziutek/telnet"
)

type (
	glas struct {
		address string
		*telnet.Conn
		iochan chan string
		ioout  io.Writer
		ioerr  io.Writer

		_conf *conf
		_quit chan struct{}

		aliasesMutex *sync.Mutex
		_aliases     aliases
	}
)

func (g *glas) send(i interface{}) error {
	var err error
	switch i.(type) {
	case []byte:
		byt := i.([]byte)
		byt = append(byt, '\n')
		_, err = g.Conn.Write(byt)
	case string:
		_, err = g.Conn.Write([]byte(fmt.Sprintf("%s\n", i.(string))))
	default:
		err = errors.New("Invalid data type")
	}

	return err
}

func (g *glas) handleConnection() {
	for {
		select {
		case <-g._quit:
			return
		default:
			data, err := g.Conn.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintln(g.ioerr, err.Error())
				}

				fmt.Fprintln(g.ioout, "Disconnected.")
				return
			}

			data = strings.TrimFunc(data, func(c rune) bool { return c == '\r' || c == '\n' })

			if err := g.observe(data); err != nil {
				fmt.Fprintln(g.ioerr, err.Error())
				return
			}
		}
	}
}

// Start starts the core services: ioerr and ioout returns all errors
// so that they can be handled from a terminal or gui application.
// While iochan handles input from the client.
func Start(iochan chan string, ioout, ioerr io.Writer, file, address string, _quit chan struct{}) {
	g := &glas{
		iochan:       iochan,
		ioout:        ioout,
		ioerr:        ioerr,
		_aliases:     make(map[string]*alias),
		aliasesMutex: &sync.Mutex{},
		_quit:        _quit,
	}

	var err error
	if file != "" {
		g._conf, err = g.loadConf(file)
		if err != nil {
			fmt.Fprintf(g.ioerr, "%s\nloading character file %s, loading blank character file\n", err.Error(), file)
		}
	}

	if g._conf != nil && g._conf.Connect.Address != "" {
		g.address = g._conf.Connect.Address
	}

	// If address was passed, prefer it!
	if address != "" {
		g.address = address
	}

	if g.address == "" {
		fmt.Fprintln(g.ioerr, errors.New("Address required"))
		return
	}

	if err = g.connect(); err != nil {
		fmt.Fprintln(g.ioerr, err.Error())
		return
	}

	for {
		select {
		case <-g._quit:
			return
		case in := <-g.iochan:
			if err := g.handleCommand(in); err != nil {
				fmt.Fprintln(g.ioerr, errors.Wrap(err, "handleCommand"))
				return
			}
		}
	}
}
