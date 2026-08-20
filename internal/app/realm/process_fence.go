package realm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
)

// Fence authenticates to every surviving local worker matching the exact
// durable allocation generation and asks it to close public QUIC before the
// private drain returns. Recovery may start a replacement only after all such
// control endpoints are unreachable.
func (allocator *ProcessAllocator) Fence(ctx context.Context, spec GameSpec) error {
	if allocator == nil || ctx == nil {
		return ErrWorker
	}

	gameID, allocationID := strings.TrimSpace(spec.GameID), strings.TrimSpace(spec.AllocationID)
	if gameID == "" || allocationID == "" {
		return ErrWorker
	}

	entries, err := os.ReadDir(allocator.config.StateDirectory)
	if err != nil {
		return fmt.Errorf("%w: scan worker state: %v", ErrWorker, err)
	}

	fenced := 0

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".worker-") {
			continue
		}

		directory := filepath.Join(allocator.config.StateDirectory, entry.Name())

		ready, readErr := ReadWorkerProcessReady(filepath.Join(directory, "ready.json"))
		if readErr != nil || ready.GameID != gameID || ready.AllocationID != allocationID {
			continue
		}

		if err := fenceProcessWorker(ctx, directory, ready); err != nil {
			return err
		}

		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("%w: remove fenced worker state: %v", ErrWorker, err)
		}

		fenced++
	}

	if fenced == 0 {
		return fmt.Errorf("%w: no surviving authority could be positively fenced", ErrWorker)
	}

	return nil
}

// fenceProcessWorker contains fence process worker within the process fence boundary so callers do not duplicate its
// domain-specific policy.
func fenceProcessWorker(ctx context.Context, directory string, ready WorkerProcessReady) error {
	token, err := os.ReadFile(filepath.Join(directory, "control-token"))
	if err != nil || len(token) < 32 || len(token) > 4096 {
		return fmt.Errorf("%w: read surviving worker credential", ErrWorker)
	}

	trust, err := networktrust.New(directory)
	if err != nil {
		return err
	}

	_, clientTLS, fingerprint, err := trust.HostTLS()
	if err != nil || !strings.EqualFold(fingerprint, ready.GameEndpoint.TLSFingerprint) {
		return fmt.Errorf("%w: surviving worker TLS identity differs", ErrWorker)
	}

	transport := &http.Transport{TLSClientConfig: clientTLS, ForceAttemptHTTP2: true,
		DialContext:     (&net.Dialer{Timeout: time.Second, KeepAlive: 10 * time.Second}).DialContext,
		IdleConnTimeout: 5 * time.Second, MaxIdleConnsPerHost: 1}

	client, err := NewWorkerHTTPClient("https://"+ready.ControlAddress, strings.TrimSpace(string(token)),
		&http.Client{Transport: transport, Timeout: 2 * time.Second})
	if err != nil {
		return err
	}

	description, err := client.Describe(ctx)
	if err != nil || description.GameID != ready.GameID {
		return errors.Join(ErrWorker, err)
	}

	if err := client.Close(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		probeCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		_, probeErr := client.Status(probeCtx)

		cancel()

		if probeErr != nil {
			// The drain callback closes public QUIC before it returns, so loss of
			// the authenticated control endpoint is the fencing barrier.
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
