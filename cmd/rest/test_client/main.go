package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

var usage = "Usage: %s <websocket address>\n"

func _main() error {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sc
		cancel()
	}()

	var args []string
	if args = os.Args[1:]; len(args) != 1 {
		return fmt.Errorf(usage, os.Args[0])
	}

	_url := url.URL{Scheme: "ws", Host: args[0], Path: "/api/connect"}
	fmt.Println("dialing " + _url.String())
	conn, _, err := websocket.DefaultDialer.Dial(_url.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	errCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			err = conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
			if err != nil {
				errCh <- err
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if err != io.EOF {
				errCh <- err
			}
		}
	}()

	go func() {
		for {
			_, byt, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}

			fmt.Print(string(byt))
		}
	}()

	select {
	case <-ctx.Done():
		break
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
