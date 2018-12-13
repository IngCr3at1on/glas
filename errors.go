package glas

import "github.com/pkg/errors"

var (
	// ErrNilConfig is returned if the provided config is nil.
	ErrNilConfig = errors.New("config cannot be nil")
	// ErrNilInput is returned if the config.Input value is nil.
	ErrNilInput = errors.New("config.Input cannot be nil")
	// ErrNilOutput is returned if the config.Output value is nil.
	ErrNilOutput = errors.New("config.Output cannot be nil")
	// ErrNilContext is returned if the ctx provided to start is nil.
	ErrNilContext = errors.New("ctx cannot be nil")
	// ErrNilCancelF is returned if the cancel function provided to start is nil.
	ErrNilCancelF = errors.New("cancel cannot be nil")
)
