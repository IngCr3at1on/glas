package internal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"github.com/ingcr3at1on/glas"
	"github.com/ingcr3at1on/glas/cmd/rest/config"
	"github.com/ingcr3at1on/glas/internal/ansi"
	pb "github.com/ingcr3at1on/glas/proto"
	"github.com/justanotherorganization/l5424"
	"github.com/justanotherorganization/l5424/x5424"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var upgrader websocket.Upgrader

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

		var wg sync.WaitGroup
		errCh := make(chan error, 1)
		outCh := make(chan *pb.Output)

		// Read from Glas.
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				out := <-outCh
				if out != nil {
					out.Data = ansi.ReplaceCodes(out.Data)

					m := new(jsonpb.Marshaler)
					var buf bytes.Buffer
					if err := m.Marshal(&buf, out); err != nil {
						errCh <- err
						return
					}

					err = ws.WriteMessage(websocket.TextMessage, buf.Bytes())
					if err != nil {
						errCh <- err
						return
					}
				}
			}
		}()

		g, err := glas.New(&glas.Config{
			Input:  inR,
			Output: outCh,
		})
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(c.Request().Context())

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
