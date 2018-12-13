package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ingcr3at1on/glas"
)

// Wrap our functionality to allow defer to work with exit.
func _main() error {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sc
		cancel()
	}()

	g, err := glas.New(&glas.Config{
		Input:  os.Stdin,
		Output: os.Stdout,
	})
	if err != nil {
		return err
	}

	if err := g.Start(ctx, cancel); err != nil {
		return err
	}

	fmt.Println("exiting")
	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
