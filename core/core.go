package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ziutek/telnet"
)

type (
	entropy struct {
		*telnet.Conn

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

			/*
				if _, err := os.Stdout.WriteString(fmt.Sprintf("%s", data)); err != nil {
					fmt.Println(err.Error())
					return
				}
			*/

			if err := e.observe(data); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}
}

// Start starts the core client and bot services.
func Start(file string) {
	e := &entropy{
		_aliases:     make(map[string]*alias),
		aliasesMutex: &sync.Mutex{},

		_wander:      false,
		roomMap:      make(map[int64]room),
		roomMapMutex: &sync.Mutex{},
	}

	var err error
	e.Conn, err = telnet.Dial("tcp", "216.69.243.18:443")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	quit := make(chan struct{})
	go e.handleInput(quit)
	go e.handleConnection(quit)

	if file != "" {
		char, err := e.loadCharacter(file)
		if err != nil {
			fmt.Printf("error loading character file %s, loading blank character", file)
			goto end
		}

		if char.AutoLogin != nil {
			if err := e.handleAutoLogin(char); err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		func(c *character) {
			e.aliasesMutex.Lock()
			defer e.aliasesMutex.Unlock()
			e._aliases = char.Aliases
		}(char)
	}

end:
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
