package glas // import "github.com/ingcr3at1on/glas"

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ingcr3at1on/glas/internal"
	pb "github.com/ingcr3at1on/glas/proto"
	"github.com/pkg/errors"
	telnet "github.com/reiver/go-telnet"
)

var (
	welcome = strings.Join([]string{
		`\ \      / /__| | ___ ___  _ __ ___   ___  | |_ ___    / ___| | __ _ ___`,
		" \\ \\ /\\ / / _ \\ |/ __/ _ \\| '_ ` _ \\ / _ \\ | __/ _ \\  | |  _| |/ _` / __|",
		`  \ V  V /  __/ | (_| (_) | | | | | |  __/ | || (_) | | |_| | | (_| \__ \_`,
		`   \_/\_/ \___|_|\___\___/|_| |_| |_|\___|  \__\___/   \____|_|\__,_|___( )`,
		`                                                                        |/`,
		`                                         _                      _        _`,
		`  __ _ _ __     _____  ___ __   ___ _ __(_)_ __ ___   ___ _ __ | |_ __ _| |`,
		" / _` | '_ \\   / _ \\ \\/ / '_ \\ / _ \\ '__| | '_ ` _ \\ / _ \\ '_ \\| __/ _` | |",
		`| (_| | | | | |  __/>  <| |_) |  __/ |  | | | | | | |  __/ | | | || (_| | |`,
		` \__,_|_| |_|  \___/_/\_\ .__/ \___|_|  |_|_| |_| |_|\___|_| |_|\__\__,_|_|`,
		`                        |_|`,
		` __  __ _   _ ____         _ _            _`,
		`|  \/  | | | |  _ \    ___| (_) ___ _ __ | |_`,
		`| |\/| | | | | | | |  / __| | |/ _ \ '_ \| __|`,
		`| |  | | |_| | |_| | | (__| | |  __/ | | | |_`,
		`|_|  |_|\___/|____/   \___|_|_|\___|_| |_|\__|`,
	}, "\n")
)

type (
	// Glas is a mud client backend.
	Glas struct {
		config    *Config
		errCh     chan error
		terrCh    chan error
		pipeW     *io.PipeWriter
		pipeR     *io.PipeReader
		connected bool
	}
)

// New returns a new instance of Glas.
func New(config *Config) (*Glas, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config.Validate")
	}

	g := Glas{
		config: config,
		errCh:  make(chan error, 1),
		terrCh: make(chan error),
	}

	g.pipeR, g.pipeW = io.Pipe()

	return &g, nil
}

// Start the Glas client.
func (g *Glas) Start(ctx context.Context, cancel context.CancelFunc) error {
	if ctx == nil {
		return ErrNilContext
	}

	if cancel == nil {
		return ErrNilCancelF
	}

	g.config.Output <- &pb.Output{Data: welcome}
	// Add help and mention it here...
	g.config.Output <- &pb.Output{Data: fmt.Sprintf("The current command prefix is '%s'\n", g.config.CmdPrefix)}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := g.routineHandleInput(ctx, cancel); err != nil {
			g.errCh <- errors.Wrap(err, "routineHandleInput")
		}

		// fmt.Println("routineHandleInput finished")
	}()

	select {
	case <-ctx.Done():
		break
	case err := <-g.errCh:
		if err != nil {
			return err
		}
	case err := <-g.terrCh:
		if err != nil {
			g.config.Output <- &pb.Output{Data: fmt.Sprintf("%s\n", err.Error())}
		}
	}

	wg.Wait()
	return nil
}

func (g *Glas) startTelnet(addr string) {
	// FIXME: support tls
	conn, err := telnet.DialTo(addr)
	if err != nil {
		g.terrCh <- err
		return
	}
	defer conn.Close()

	caller, err := internal.NewCaller(&internal.CallerConfig{
		In:  g.pipeR,
		Out: g.config.Output,
	}, g.terrCh)
	if err != nil {
		g.terrCh <- err
		return
	}

	client := telnet.Client{
		Caller: caller,
	}

	g.connected = true

	err = client.Call(conn)
	if err != nil {
		g.terrCh <- errors.Wrapf(err, "client.Call : %s", conn.RemoteAddr().String())
		return
	}
}
