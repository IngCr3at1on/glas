package glas

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	// Glas is our mud client.
	Glas struct {
		log *logrus.Entry

		out    io.Writer
		config *Config
		conn   *conn

		errCh  chan error
		stopCh chan error

		characters *characters
	}
)

// New returns a new instance of Glas.
func New(config *Config, out io.Writer, errCh, stopCh chan error, log *logrus.Entry) (*Glas, error) {
	if out == nil {
		return nil, errors.New("out cannot be nil")
	}

	if errCh == nil {
		return nil, errors.New("errCh cannot be nil")
	}

	if stopCh == nil {
		return nil, errors.New("stopCh cannot be nil")
	}

	if config == nil {
		config = &Config{}
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config.Validate")
	}

	if log == nil {
		log = logrus.NewEntry(logrus.New())
	}

	g := &Glas{
		log:    log,
		out:    out,
		config: config,
		characters: &characters{
			m: make(map[string]*CharacterConfig),
		},
		errCh:  errCh,
		stopCh: stopCh,
	}

	g.conn = &conn{
		glas: g,
	}

	if err := g.loadCharacterConfigs(); err != nil {
		return nil, errors.Wrap(err, "loadCharacterConfigs")
	}

	return g, nil
}

// Start starts our mud client.
func (g *Glas) Start(connectArg string) {
	fmt.Fprint(g.out, welcome)
	fmt.Fprint(g.out, fmt.Sprintf("\nThe current command prefix is '%s', you may get help at any time using %shelp", g.config.CmdPrefix, g.config.CmdPrefix))

	if connectArg != "" {
		if err := g.connect(connectArg); err != nil {
			g.errCh <- err
			return
		}
	}
}

// Send data to the mud.
func (g *Glas) Send(data ...interface{}) error {
	g.log.WithFields(logrus.Fields{
		"command": "Send",
		"data":    data,
	}).Debug("Called")

	for i, d := range data {
		var str string

		switch d.(type) {
		case []byte:
			str = string(d.([]byte))
		case string:
			str = d.(string)
		default:
			return errors.New("Invalid data type")
		}

		ok, err := g.handleCommand(str)
		if err != nil {
			return err
		}

		// FIXME:
		// if !ok && g.characterConfig != nil {
		// 	ok, err = g.characterConfig.aliases.maybeHandleAlias(g, str)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		if !ok && g.conn != nil && g.conn.connected {
			if _, err := g.conn.Write([]byte(fmt.Sprintf("%s\n", str))); err != nil {
				return err
			}
		}

		if i+1 < len(data) {
			// TODO: make this configurable.
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}
