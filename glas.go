package glas // import "github.com/IngCr3at1on/glas"

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/IngCr3at1on/glas/internal"
	"github.com/pkg/errors"
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
		config *Config
	}
)

// New returns a new instance of Glas.
func New(config *Config) (*Glas, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config.Validate")
	}

	return &Glas{
		config: config,
	}, nil
}

// Connect to a mud passing the data through to the provided io.ReadWriter.
func (g *Glas) Connect(ctx context.Context, addr string, rw io.ReadWriter) error {
	if addr = strings.TrimSpace(addr); addr == "" {
		return errors.New("addr cannot be empty")
	}

	if rw == nil {
		return errors.New("rw cannot be nil")
	}

	fmt.Fprintln(rw, welcome)
	fmt.Fprintf(rw, "The current command prefix is '%s', you may get help at any time using %[1]shelp\n", g.config.CmdPrefix)

	time.Sleep(100 * time.Millisecond)

	// FIXME: This ctx should be a dial context and not the one used to control shutdown
	conn, err := internal.Dial(ctx, addr)
	if err != nil {
		return errors.Wrap(err, "internal.Dial")
	}

	if err := conn.Listen(ctx, rw); err != nil {
		return errors.Wrap(err, "conn.Listen")
	}

	return nil
}
