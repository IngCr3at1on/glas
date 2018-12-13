package glas

import (
	"bufio"
	"context"
	"io"
	"strings"
)

func (g *Glas) routineHandleInput(ctx context.Context, cancel context.CancelFunc) error {
	// fmt.Println("routineHandleInput")
	scanner := bufio.NewScanner(g.config.Input)

	errCh := make(chan error, 1)
	go func() {
		for scanner.Scan() {
			b, err := g.handleInput(cancel, scanner.Text())
			if err != nil {
				errCh <- err
				return
			}

			if !b && g.connected {
				nw, err := g.pipeW.Write(scanner.Bytes())
				if err != nil {
					errCh <- err
					return
				}

				if nw != len(scanner.Bytes()) {
					errCh <- io.ErrShortWrite
					return
				}
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			errCh <- err
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

		switch input {
		case "exit":
			cancel()
		case "connect":
			go g.startTelnet("216.69.243.18:7000")
		}

		return true, nil
	}

	return false, nil
}
