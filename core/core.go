package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/ziutek/telnet"

	"github.com/IngCr3at1on/glas/ansi"
)

type (
	entropy struct {
		address string
		*telnet.Conn

		conf *conf

		aliasesMutex *sync.Mutex
		_aliases     aliases

		_wander      bool
		roomMapMutex *sync.Mutex
		roomMap      map[int64]room
		here         int64
	}

	exit struct {
		direction   string
		destination int64
	}

	room struct {
		id    int64
		exits []exit
		// TODO handle hidden exits
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

func (e *entropy) handleConnection(quit chan struct{}) {
	for {
		select {
		case <-quit:
			return
		default:
			data, err := e.Conn.ReadString('\n')
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			data = strings.TrimFunc(data, func(c rune) bool {
				return c == '\r' || c == '\n'
			})

			// Strip out the background color for printing.
			// TODO possibly control this by a setting?
			fmt.Println(ansi.Strip(data, ansi.Bg))

			if err := e.observe(data); err != nil {
				fmt.Println(err.Error())
				return
			}

			// TODO reset ansi color back to default
			// (should fix input line from having the ansi color for the last output line)
		}
	}
}

func (e *entropy) connect(quit chan struct{}) error {
	var err error
	e.Conn, err = telnet.Dial("tcp", e.address)
	if err != nil {
		return errors.Wrap(err, "telnet.Dial")
	}

	if e.conf != nil {
		if e.conf.Connect.AutoLogin != nil {
			if err := e.handleAutoLogin(e.conf); err != nil {
				return errors.Wrap(err, "handleAutoLogin")
			}
		}

		func(c *conf) {
			e.aliasesMutex.Lock()
			defer e.aliasesMutex.Unlock()
			e._aliases = c.Aliases
		}(e.conf)
	}

	// Ensure that we only start our handleConnection thread once
	// this way if we disconnect/reconnect we don't have multiple
	// go routines attempting to read the connection.
	_connect := func() {
		go e.handleConnection(quit)
	}
	var once sync.Once
	once.Do(_connect)

	return nil
}

// Start starts the core client and bot services.
func Start(file, address string) {
	e := &entropy{
		_aliases:     make(map[string]*alias),
		aliasesMutex: &sync.Mutex{},

		_wander:      false,
		roomMap:      make(map[int64]room),
		roomMapMutex: &sync.Mutex{},
	}

	var (
		err error
	)

	if file != "" {
		e.conf, err = e.loadCharacter(file)
		if err != nil {
			fmt.Printf("%s\nloading character file %s, loading blank character file\n", err.Error(), file)
		}
	}

	if e.conf != nil && e.conf.Connect.Address != "" {
		e.address = e.conf.Connect.Address
	}

	// If address was passed, prefer it!
	if address != "" {
		e.address = address
	}

	if e.address == "" {
		fmt.Println("Address required")
		return
	}

	// TODO make this part of our struct instead of passing it everywhere.
	quit := make(chan struct{})
	if err = e.connect(quit); err != nil {
		fmt.Println(err.Error())
		return
	}
	go e.handleInput(quit)

	e.roomMapMutex.Lock()
	// Add the church to our map
	// TODO save the map to db (and load here)
	id := int64(len(e.roomMap)) + 1
	e.roomMap[id] = room{
		id:    id,
		exits: []exit{exit{"s", id + 1}},
	}
	e.roomMapMutex.Unlock()
	e.here = id

	go e.wander(quit)

	// Block until quit is called
	<-quit
}
