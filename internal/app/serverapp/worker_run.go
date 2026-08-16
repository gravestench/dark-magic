package serverapp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
)

// RunRealmWorker owns the long-lived headless services for an allocated worker.
// A drain request follows the same cancellation and join path as SIGTERM.
func RunRealmWorker(ctx context.Context, host *gameserver.Host, quic *sessionquic.Server, control *WorkerControlServer, drain <-chan struct{}) error {
	if ctx == nil || host == nil || host.Session == nil || quic == nil || control == nil || drain == nil {
		return errors.New("server: realm worker requires context, host, QUIC, control, and drain")
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errorsChannel := make(chan error, 3)
	var group sync.WaitGroup
	group.Add(3)
	go func() {
		defer group.Done()
		errorsChannel <- wrapWorkerRunError("session", host.Session.Run(runCtx))
	}()
	go func() {
		defer group.Done()
		errorsChannel <- wrapWorkerRunError("QUIC", quic.Serve(runCtx))
	}()
	go func() {
		defer group.Done()
		errorsChannel <- wrapWorkerRunError("control", control.Serve(runCtx))
	}()
	var first error
	controlledShutdown := false
	select {
	case <-ctx.Done():
		controlledShutdown = true
	case <-drain:
		controlledShutdown = true
	case first = <-errorsChannel:
	}
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = control.Close(shutdownCtx)
	shutdownCancel()
	_ = quic.Close()
	group.Wait()
	close(errorsChannel)
	return realmWorkerResult(controlledShutdown, first)
}

func realmWorkerResult(controlledShutdown bool, first error) error {
	if controlledShutdown {
		// Closing listeners is how a deliberate drain wakes their Serve loops;
		// listener-closed errors are consequences of shutdown, not worker
		// failures. The initiating control request has already fenced QUIC.
		return nil
	}
	if first == nil {
		return errors.New("server: Realm worker service stopped unexpectedly")
	}
	return first
}

func wrapWorkerRunError(service string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	return fmt.Errorf("server: realm worker %s: %w", service, err)
}
