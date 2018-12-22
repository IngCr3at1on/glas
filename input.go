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
		for {
			in := <-g.config.Input
			if in != nil {
				b, err := g.handleInput(cancel, in.Data)
				if err != nil {
					errCh <- err
					return
				}

				if !b && g.connected {
					nw, err := g.pipeW.Write([]byte(in.Data))
					if err != nil {
						errCh <- err
						return
					}

					if nw != len(in.Data) {
						errCh <- io.ErrShortWrite
						return
					}
				}
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
