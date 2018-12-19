package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ingcr3at1on/glas/cmd/rest/config"
	"github.com/ingcr3at1on/glas/cmd/rest/internal"
	"github.com/labstack/echo"
)

// Wrap our functionality to allow defer to work with exit.
func _main() error {
	e := echo.New()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go func() {
		<-sc
		fmt.Println("shutting down")
		e.Shutdown(context.Background())
	}()

	if err := internal.SetupRoutes(new(config.Config), e.Group("/api")); err != nil {
		return err
	}

	// FIXME: make this path more friendly...
	e.Static("/", "./static/terminal.html")

	addr := ":8080"
	return e.Start(addr)
}

func main() {
	if err := _main(); err != nil {
		if err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}
