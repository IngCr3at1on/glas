package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/IngCr3at1on/glas"
	"github.com/IngCr3at1on/glas/cmd/rest/glas/config"
	"github.com/IngCr3at1on/glas/cmd/rest/glas/internal"
	"github.com/IngCr3at1on/glas/internal/safe"
	"github.com/labstack/echo"
)

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

	_glas, err := glas.New(&glas.Config{})
	if err != nil {
		log.Log(err.Error())
		exit()
	}

	e := echo.New()

	var wg sync.WaitGroup
	addr := ":4242"
	if err := internal.SetupRoutes(ctx, cancel, &config.Config{
		Address: addr,
		Glas:    _glas,
		// TODO: test if we're waiting properly on these handlers and if not try to figure out how to do so.
	}, e.Group("/api"), &wg); err != nil {
		log.Log(err.Error())
		exit()
	}

	wg.Add(1)
	go func() {
		defer func() {
			log.Log("FOO!")
			wg.Done()
		}()

		if err := e.Start(addr); err != nil {
			log.Log(err.Error())
			exit()
		}
	}()

	if err := safe.SetupShutdown(cancel, &wg, log); err != nil {
		log.Log(err.Error())
		exit()
	}

	<-ctx.Done()
}

type logger struct{}

// Log writes to Println
func (l logger) Log(v ...interface{}) error {
	_, err := fmt.Println(v...)
	return err
}

var _ safe.Logger = &logger{}
