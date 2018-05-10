package safe

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	// TODO: should this be guarded in a mutex?
	_ready      bool
	_exitCodeCh chan int
	_logger     Logger
)

// SetupShutdown sets up a safe shutdown option using the provided context.CancelFunc and sync.WaitGroup.
// SetupShutdown must be called AFTER all thread counts have been added to the waitgroup.
func SetupShutdown(cancel context.CancelFunc, wg *sync.WaitGroup, log Logger) error {
	if wg == nil {
		return errors.New("wg cannot be nil")
	}

	if log == nil {
		log = &noOpLogger{}
	}
	_logger = log

	_exitCodeCh = make(chan int, 1)

	go func() {
		// Block until we receive an exit code.
		exitCode := <-_exitCodeCh
		cancel()
		wg.Wait()
		os.Exit(exitCode)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		Shutdown(0)
	}()

	_ready = true
	return nil
}

// Shutdown initiates a safe shutdown using the provided code.
func Shutdown(code int) {
	_logger.Log("Shutdown called...")
	_exitCodeCh <- code

	// Stal, waiting for os.Exit.
	stal := make(chan struct{}, 1)
	<-stal
}

// Ready can be used to confirm SetupShutdown has been called.
func Ready() bool {
	return _ready
}

type (
	// Logger is a bare logger interface as described in the current standardized logging proposals.
	Logger interface {
		Log(...interface{}) error
	}

	noOpLogger struct{}
)

var _ Logger = &noOpLogger{}

// Log returns nil without performing any operations.
func (l *noOpLogger) Log(v ...interface{}) error {
	return nil
}
