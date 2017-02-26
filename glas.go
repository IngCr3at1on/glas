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
	entropy struct {
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

func (e *entropy) send(i interface{}) error {
	var err error
	switch i.(type) {
	case []byte:
		byt := i.([]byte)
		byt = append(byt, '\n')
		_, err = e.Conn.Write(byt)
	case string:
		_, err = e.Conn.Write([]byte(fmt.Sprintf("%s\n", i.(string))))
	default:
		err = errors.New("Invalid data type")
	}

	return err
}

func (e *entropy) handleConnection() {
	for {
		select {
		case <-e._quit:
			return
		default:
			data, err := e.Conn.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintln(e.ioerr, err.Error())
				}

				fmt.Fprintln(e.ioout, "Disconnected.")
				return
			}

			data = strings.TrimFunc(data, func(c rune) bool { return c == '\r' || c == '\n' })

			if err := e.observe(data); err != nil {
				fmt.Fprintln(e.ioerr, err.Error())
				return
			}
		}
	}
}

func (e *entropy) connect() error {
	var err error
	e.Conn, err = telnet.Dial("tcp", e.address)
	if err != nil {
		return errors.Wrap(err, "telnet.Dial")
	}

	if e._conf != nil {
		if e._conf.Connect.AutoLogin != nil {
			if err := e.handleAutoLogin(e._conf); err != nil {
				return errors.Wrap(err, "handleAutoLogin")
			}
		}

		func(c *conf) {
			e.aliasesMutex.Lock()
			defer e.aliasesMutex.Unlock()
			e._aliases = c.Aliases
		}(e._conf)
	}

	// Ensure that we only start our handleConnection thread once
	// this way if we disconnect/reconnect we don't have multiple
	// go routines attempting to read the connection.
	_connect := func() {
		go e.handleConnection()
	}
	var once sync.Once
	once.Do(_connect)

	return nil
}

// Start starts the core services: ioerr and ioout returns all errors
// so that they can be handled from a terminal or gui application.
// While iochan handles input from the client.
// TODO combines these into a single channel?
func Start(iochan chan string, ioout, ioerr io.Writer, file, address string) {
	e := &entropy{
		iochan:       iochan,
		ioout:        ioout,
		ioerr:        ioerr,
		_aliases:     make(map[string]*alias),
		aliasesMutex: &sync.Mutex{},
	}

	var err error
	if file != "" {
		e._conf, err = e.loadConf(file)
		if err != nil {
			fmt.Fprintf(e.ioerr, "%s\nloading character file %s, loading blank character file\n", err.Error(), file)
		}
	}

	if e._conf != nil && e._conf.Connect.Address != "" {
		e.address = e._conf.Connect.Address
	}

	// If address was passed, prefer it!
	if address != "" {
		e.address = address
	}

	if e.address == "" {
		fmt.Fprintln(e.ioerr, errors.New("Address required"))
		return
	}

	e._quit = make(chan struct{})
	if err = e.connect(); err != nil {
		fmt.Fprintln(e.ioerr, err.Error())
		return
	}

	for {
		select {
		case <-e._quit:
			return
		case in := <-e.iochan:
			if err := e.handleCommand(in); err != nil {
				fmt.Fprintln(e.ioerr, errors.Wrap(err, "handleCommand"))
				return
			}
		}
	}
}
