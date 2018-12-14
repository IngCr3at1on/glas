package glas

import (
	"io"
)

const (
	defaultCommandPrefix = `/`
)

type (
	// Config houses non-character configuration options.
	Config struct {
		// Input is an input io.Reader (os.Stdin for example).
		Input io.Reader
		// Output is an output io.Writer (os.Stdout for example).
		Output io.Writer
		// CmdPrefix is the prefix used for client commands, by default this
		// is `/`.
		CmdPrefix string
	}
)

// Validate a configuration, error if a required value is missing and set
// defaults (if a value is not provided) when not required.
func (c *Config) Validate() error {
	if c == nil {
		return ErrNilConfig
	}

	if c.Input == nil {
		return ErrNilInput
	}

	if c.Output == nil {
		return ErrNilOutput
	}

	if c.CmdPrefix == "" {
		c.CmdPrefix = defaultCommandPrefix
	}

	return nil
}
