package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"safe"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var usage = "Usage: %s <rest address>"

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
	if args = os.Args[1:]; len(args) != 1 {
		log.Log(fmt.Sprintf(usage, os.Args[0]))
		exit()
	}

	_url := url.URL{Scheme: "ws", Host: args[0], Path: "/api/test"}
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

	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				conn.Close()
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					errCh <- errors.Wrap(err, "conn.ReadMessage")
					return
				}

				log.Log(string(msg))
			}
		}
	}()

	if err := safe.SetupShutdown(cancel, &wg, log); err != nil {
		log.Log(err.Error())
		exit()
	}

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case err := <-errCh:
			log.Log(err.Error())
			exit()
			//default:
			// TODO: write back
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
