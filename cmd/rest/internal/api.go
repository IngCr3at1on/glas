package internal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ingcr3at1on/glas"
	"github.com/ingcr3at1on/glas/cmd/rest/config"
	"github.com/ingcr3at1on/glas/internal/ansi"
	"github.com/justanotherorganization/l5424"
	"github.com/justanotherorganization/l5424/x5424"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var (
	upgrader websocket.Upgrader
	regex    = regexp.MustCompile(`(\\(033|x1b)|)`)
)

func init() {
	upgrader.CheckOrigin = func(req *http.Request) bool {
		// if req.Host == "127.0.0.1" ||
		// req.Host == "localhost" ||
		// req.Host == "ingcr3at1on.online" ||
		// req.Host == "eyeofmidas.net" {
		// return true
		// }
		// FIXME: this is not good
		return true
		// return false
	}
}

// SetupRoutes sets up the API routes.
func SetupRoutes(config *config.Config, g *echo.Group) error {
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "config.Validate")
	}

	g.GET("/ready", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	g.GET("/connect", makeConnectHandler(config))

	return nil
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

		// Read from Glas.
		wg.Add(1)
		go func() {
			defer wg.Done()
			rbuf := make([]byte, 1)
			// TODO: consider moving this functionality into glas itself
			// (instead of part of the API).
			var wbuf bytes.Buffer
			readByLine := false

			for {
				nr, er := outR.Read(rbuf)
				if nr > 0 {
					_, err := wbuf.Write(rbuf[:nr])
					if err != nil {
						errCh <- err
						return
					}

					if regex.MatchString(wbuf.String()) {
						readByLine = true
					}

					if readByLine && strings.Contains(wbuf.String(), "\r\n") || !readByLine {
						err = ws.WriteMessage(websocket.TextMessage, ansi.ReplaceCodes(wbuf.Bytes()))
						if err != nil {
							errCh <- err
							return
						}

						wbuf.Reset()
					}

					continue
				}
				if er != nil {
					if er != io.EOF {
						errCh <- er
					}
					break
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.Start(ctx, cancel); err != nil {
				errCh <- err
			}
		}()

		// Write to Glas.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, byt, err := ws.ReadMessage()
				if err != nil {
					errCh <- err
					return
				}

				w, err := inW.Write(byt)
				if err != nil {
					errCh <- err
					return
				}

				if w != len(byt) {
					errCh <- io.ErrShortWrite
					return
				}
			}
		}()

		select {
		case <-c.Request().Context().Done():
			break
		case <-ctx.Done():
			cancel()
			break
		case err := <-errCh:
			if err != nil {
				cancel()
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
				) {
					return echo.NewHTTPError(http.StatusInternalServerError, err)
				}
			}
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte("closing connection")); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "error writing to websocket"))
		}

		wg.Wait()
		return c.NoContent(http.StatusOK)
	}
}
