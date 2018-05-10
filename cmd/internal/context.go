package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ContextWithSignal returns a context which is cancelled by a SIGINT or SIGTERM.
func ContextWithSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	// FIXME: not sure this works on windows; https://golang.org/pkg/os/#Signal
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	return ctx
}
