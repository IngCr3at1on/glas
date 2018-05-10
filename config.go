package glas

import (
	"github.com/pkg/errors"
)

const (
	defaultCommandPrefix = `/`
)

type (
	// Config houses non-character configuration options.
	Config struct {
		// CmdPrefix is the prefix used for client commands, by default this
		// is `/`.
		CmdPrefix string
	}
)

// Validate a configuration, error if a required value is missing and set
// defaults (if a value is not provided) when not required.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config cannot be nil")
	}
	if c.CmdPrefix == "" {
		c.CmdPrefix = defaultCommandPrefix
	}

	return nil
}
