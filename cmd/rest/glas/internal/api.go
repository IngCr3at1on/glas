package internal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/IngCr3at1on/glas/cmd/rest/glas/config"
	"github.com/JustAnotherOrganization/l5424"
	"github.com/JustAnotherOrganization/l5424/x5424"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var upgrader websocket.Upgrader

// SetupRoutes sets up the API routes.
func SetupRoutes(ctx context.Context, cancel context.CancelFunc, config *config.Config, g *echo.Group, wg *sync.WaitGroup) error {
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "config.Validate")
	}

	g.GET("/ready", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	g.GET("/test", makeTestHandler(ctx, config, wg))
	g.GET("/connect/:address", makeConnectHandler(ctx, cancel, config, wg))

	return nil
}

func makeTestHandler(extCtx context.Context, cfg *config.Config, wg *sync.WaitGroup) echo.HandlerFunc {
	return func(c echo.Context) error {
		wg.Add(1)
		// FIXME: need a defer function that can be added to farther down...
		defer wg.Done()

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			errMsg := "error upgrading request to websocket"
			cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}
		// FIXME: see above...
		defer ws.Close()

		for {
			select {
			case <-extCtx.Done():
				return echo.NewHTTPError(http.StatusServiceUnavailable, "server shutting down")
			default:
				if err := ws.WriteMessage(websocket.TextMessage, []byte("hello echo and gorilla websockets")); err != nil {
					errMsg := "error writing to websocket"
					cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
					return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
				}
			}
		}
	}
}

func makeConnectHandler(extCtx context.Context, cancel context.CancelFunc, cfg *config.Config, wg *sync.WaitGroup) echo.HandlerFunc {
	return func(c echo.Context) error {
		wg.Add(1)
		// FIXME: need a defer function that can be added to farther down...
		defer wg.Done()

		ctx := c.Request().Context()

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			errMsg := "error upgrading request to websocket"
			cfg.Logger.Log(x5424.Severity, l5424.ErrorLvl, errMsg, err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}
		// FIXME: see above...
		defer ws.Close()

		var (
			buf bytes.Buffer
			_wg sync.WaitGroup

			errCh = make(chan error)
		)

		// Start a connection to a mud.
		_wg.Add(1)
		go func() {
			defer _wg.Done()
			if err := cfg.Glas.Connect(extCtx, c.Param("address"), &buf); err != nil {
				errCh <- errors.Wrap(err, "glas.Connect")
				return
			}
		}()

		// Read from the mud and write back to the websocket.
		_wg.Add(1)
		go func() {
			defer _wg.Done()
			for {
				select {
				case <-extCtx.Done():
					return
				case <-ctx.Done():
					cancel()
					return
				default:
					w, err := ws.NextWriter(websocket.TextMessage)
					if err != nil {
						errCh <- errors.Wrap(err, "ws.NextWriter")
						return
					}

					if _, err := io.Copy(w, &buf); err != nil {
						errCh <- errors.Wrap(err, "io.Copy")
					}
				}
			}
		}()

		// Read from the websocket and pass onto the mud.
		_wg.Add(1)
		go func() {
			defer _wg.Done()
			for {
				select {
				case <-extCtx.Done():
					return
				case <-ctx.Done():
					cancel()
					return
				default:
					_, r, err := ws.NextReader()
					if err != nil {
						errCh <- errors.Wrap(err, "ws.NextReader")
						return
					}

					if _, err := io.Copy(&buf, r); err != nil {
						errCh <- errors.Wrap(err, "io.Copy")
						return
					}
				}
			}
		}()

		select {
		case <-extCtx.Done():
			return echo.NewHTTPError(http.StatusServiceUnavailable, "server shutting down")
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

		_wg.Wait()
		return c.NoContent(http.StatusOK)
	}
}
