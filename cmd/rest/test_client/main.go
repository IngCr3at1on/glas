package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/IngCr3at1on/glas/internal/safe"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var usage = "Usage: %s <rest address> <mud address>\n"

func exit() {
	if safe.Ready() {
		safe.Shutdown(1)
	} else {
		os.Exit(1)
	}
}

func main() {
	log := &logger{}
	ctx, cancel := context.WithCancel(context.Background())

	var args []string
	if args = os.Args[1:]; len(args) != 2 {
		log.Log(fmt.Sprintf(usage, os.Args[0]))
		exit()
	}

	_url := url.URL{Scheme: "ws", Host: args[0], Path: fmt.Sprintf("/api/connect/%s", args[1])}
	conn, _, err := websocket.DefaultDialer.Dial(_url.String(), nil)
	if err != nil {
		log.Log(err.Error())
		exit()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	defer func() {
		conn.Close()
		wg.Done()
	}()

	if err := safe.SetupShutdown(cancel, &wg, log); err != nil {
		log.Log(err.Error())
		exit()
	}

	// time.Sleep(100 * time.Millisecond)

out:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Log(errors.Wrap(err, "conn.ReadMessage").Error())
				go exit()
				break out
			}

			fmt.Print(string(msg))
		}
	}
}

type logger struct{}

// Log writes to Println
func (l logger) Log(v ...interface{}) error {
	_, err := fmt.Println(v...)
	return err
}

var _ safe.Logger = &logger{}
