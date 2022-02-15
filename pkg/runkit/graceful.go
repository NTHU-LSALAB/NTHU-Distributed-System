package runkit

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GracefulConfig struct {
	Timeout time.Duration `long:"timeout" description:"gracefully timeout duration" env:"TIMEOUT" default:"10s"`
}

var (
	ErrGracefullyTimeout = errors.New("gracefully shutdown timeout")
)

type GracefulRunFunc func(context.Context) error

func GracefulRun(fn GracefulRunFunc, conf *GracefulConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-done:
		return err
	case <-shutdownCh:
		// receive termination signal, cancel manually
		cancel()

		select {
		case err := <-done:
			// gracefully shutdown
			return err
		case <-time.After(conf.Timeout):
			// timeout shutdown
			return ErrGracefullyTimeout
		}
	}
}
