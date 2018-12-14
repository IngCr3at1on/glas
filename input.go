package glas

import (
	"context"
	"io"
	"strings"
)

func (g *Glas) routineHandleInput(ctx context.Context, cancel context.CancelFunc) error {
	// fmt.Println("routineHandleInput")

	errCh := make(chan error, 1)
	go func() {
		rbuf := make([]byte, 1024)
		for {
			nr, er := g.config.Input.Read(rbuf)
			if nr > 0 {
				byt := rbuf[:nr]
				b, err := g.handleInput(cancel, string(byt))
				if err != nil {
					errCh <- err
					return
				}

				if !b && g.connected {
					nw, err := g.pipeW.Write(byt)
					if err != nil {
						errCh <- err
						return
					}

					if nw != len(byt) {
						errCh <- io.ErrShortWrite
						return
					}
				}
			}
			if er != nil {
				if er != io.EOF {
					errCh <- er
				}
				break
			}
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		break
	}

	return nil
}

func (g *Glas) handleInput(cancel context.CancelFunc, input string) (bool, error) {
	if strings.HasPrefix(input, g.config.CmdPrefix) {
		input = strings.TrimPrefix(input, g.config.CmdPrefix)

		switch {
		case strings.Compare(input, "exit") == 0:
			cancel()
		case strings.HasPrefix(input, "connect"):
			go g.startTelnet(
				strings.TrimSpace(
					strings.TrimPrefix(input, "connect")))
		}

		return true, nil
	}

	return false, nil
}
