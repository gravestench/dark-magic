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
func RunRealmWorker(
	ctx context.Context,
	host *gameserver.Host,
	quic *sessionquic.Server,
	control *WorkerControlServer,
	drain <-chan struct{},
) error {
	if ctx == nil || host == nil || host.Session == nil || quic == nil || control == nil || drain == nil {
		return errors.New("server: realm worker requires context, host, QUIC, control, and drain")
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errorsChannel := make(chan error, 3)

	var group sync.WaitGroup
	group.Add(3)
	startRealmWorkerService(&group, errorsChannel, "session", func() error { return host.Session.Run(runCtx) })
	startRealmWorkerService(&group, errorsChannel, "QUIC", func() error { return quic.Serve(runCtx) })
	startRealmWorkerService(&group, errorsChannel, "control", func() error { return control.Serve(runCtx) })

	controlledShutdown, first := waitForRealmWorkerStop(ctx, drain, errorsChannel)

	cancel()
	stopRealmWorkerTransports(control, quic)
	group.Wait()
	close(errorsChannel)

	return realmWorkerResult(controlledShutdown, first)
}

// startRealmWorkerService always reports one terminal result; the caller adds
// every service to the wait group before any launch so ownership is complete
// even when the first service exits immediately.
func startRealmWorkerService(
	group *sync.WaitGroup,
	errorsChannel chan<- error,
	name string,
	run func() error,
) {
	go func() {
		defer group.Done()

		errorsChannel <- wrapWorkerRunError(name, run())
	}()
}

// waitForRealmWorkerStop distinguishes an external cancellation or drain from
// a service exit because only the latter should surface as a worker failure.
func waitForRealmWorkerStop(
	ctx context.Context,
	drain <-chan struct{},
	errorsChannel <-chan error,
) (bool, error) {
	select {
	case <-ctx.Done():
		return true, nil
	case <-drain:
		return true, nil
	case first := <-errorsChannel:
		return false, first
	}
}

// stopRealmWorkerTransports preserves the shutdown order: the private control
// plane drains first, then QUIC closes to wake its Serve loop before joining.
func stopRealmWorkerTransports(control *WorkerControlServer, quic *sessionquic.Server) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = control.Close(shutdownCtx)

	shutdownCancel()

	_ = quic.Close()
}

// realmWorkerResult suppresses listener-close fallout only for a deliberate
// shutdown; an unexplained nil service result is still an operational failure.
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

// wrapWorkerRunError treats context cancellation as the expected common
// shutdown signal while retaining service identity and the original cause.
func wrapWorkerRunError(service string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	return fmt.Errorf("server: realm worker %s: %w", service, err)
}
