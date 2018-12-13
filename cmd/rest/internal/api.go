package internal

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ingcr3at1on/glas"
	"github.com/ingcr3at1on/glas/cmd/rest/config"
	"github.com/justanotherorganization/l5424"
	"github.com/justanotherorganization/l5424/x5424"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var upgrader websocket.Upgrader

// SetupRoutes sets up the API routes.
func SetupRoutes(config *config.Config, g *echo.Group) error {
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "config.Validate")
	}

	g.GET("/ready", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	g.GET("/test", makeTestHandler(config))
	g.GET("/connect/:address", makeConnectHandler(config))

	return nil
}

func makeTestHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			errMsg := "error upgrading request to websocket"
			cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}
		defer ws.Close()

		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello echo and gorilla websockets")); err != nil {
			errMsg := "error writing to websocket"
			cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		return c.NoContent(http.StatusOK)
	}
}

func makeConnectHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			errMsg := "error upgrading request to websocket"
			cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}
		defer ws.Close()

		inR, inW := io.Pipe()
		outR, outW := io.Pipe()

		g, err := glas.New(&glas.Config{
			Input:  inR,
			Output: outW,
		})
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		errCh := make(chan error, 1)
		ctx, cancel := context.WithCancel(c.Request().Context())

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.Start(ctx, cancel); err != nil {
				errCh <- err
			}
		}()

		// Read from Glas.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				w, err := ws.NextWriter(websocket.TextMessage)
				if err != nil {
					errCh <- errors.Wrap(err, "ws.NextWriter")
					return
				}

				if _, err := io.Copy(w, outR); err != nil {
					errCh <- err
				}
			}
		}()

		// Write to Glas.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, r, err := ws.NextReader()
				if err != nil {
					errCh <- errors.Wrap(err, "ws.NextReader")
					return
				}

				if _, err := io.Copy(inW, r); err != nil {
					errCh <- err
					return
				}
			}
		}()

		select {
		case <-ctx.Done():
			break
		case err := <-errCh:
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte("closing connection")); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "error writing to websocket"))
		}

		wg.Wait()
		return c.NoContent(http.StatusOK)
	}
}
