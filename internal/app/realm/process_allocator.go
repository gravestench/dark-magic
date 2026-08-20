package realm

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const maximumWorkerLogBytes = 256 << 10

type ProcessAllocatorConfig struct {
	Executable           string
	Arguments            []string
	WorkingDirectory     string
	Environment          []string
	StateDirectory       string
	ControlListenAddress string
	GameListenAddress    string
	GameAdvertiseHost    string
	StartupTimeout       time.Duration
	ShutdownTimeout      time.Duration
	LogWriter            io.Writer
	ExpectedAssetSetID   string
}

type ProcessGameSpec = GameSpec
type ProcessAllocation = WorkerAllocation

// ProcessAllocator supervises one ordinary cmd/server child per Realm game.
// It never invokes a shell and gives every child isolated owner-only keys,
// readiness state, and a pinned TLS identity.
type ProcessAllocator struct {
	mu           sync.RWMutex
	operations   sync.WaitGroup
	config       ProcessAllocatorConfig
	executable   string
	workers      map[string]*processWorker
	reservations map[string]struct{}
	closed       bool
}

type processWorker struct {
	client    *WorkerHTTPClient
	command   *exec.Cmd
	endpoint  GameEndpoint
	directory string
	logs      *boundedProcessLog
	done      chan struct{}
	mu        sync.Mutex
	waitErr   error
}

// NewProcessAllocator constructs the process allocator boundary and validates dependencies before callers can publish
// or mutate shared state.
func NewProcessAllocator(config ProcessAllocatorConfig) (*ProcessAllocator, error) {
	executable, err := exec.LookPath(strings.TrimSpace(config.Executable))
	if err != nil {
		return nil, fmt.Errorf("%w: resolve worker executable: %v", ErrWorker, err)
	}

	executable, err = filepath.Abs(executable)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve worker executable: %v", ErrWorker, err)
	}

	if strings.TrimSpace(config.StateDirectory) == "" {
		return nil, fmt.Errorf("%w: worker state directory is required", ErrWorker)
	}

	config.StateDirectory, err = filepath.Abs(config.StateDirectory)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve worker state directory: %v", ErrWorker, err)
	}

	if config.ControlListenAddress == "" {
		config.ControlListenAddress = "127.0.0.1:0"
	}

	if !validLoopbackAddress(config.ControlListenAddress) {
		return nil, fmt.Errorf("%w: worker control must listen on an explicit loopback host and port", ErrWorker)
	}

	if config.GameListenAddress == "" {
		config.GameListenAddress = "127.0.0.1:0"
	}

	if _, _, err := net.SplitHostPort(config.GameListenAddress); err != nil {
		return nil, fmt.Errorf("%w: invalid game listen address", ErrWorker)
	}

	if strings.ContainsAny(config.GameAdvertiseHost, " /\\\t\r\n") {
		return nil, fmt.Errorf("%w: invalid game advertise host", ErrWorker)
	}

	if config.StartupTimeout <= 0 {
		config.StartupTimeout = 30 * time.Second
	}

	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = 10 * time.Second
	}

	if config.StartupTimeout > 2*time.Minute || config.ShutdownTimeout > time.Minute {
		return nil, fmt.Errorf("%w: worker lifecycle timeout is unbounded", ErrWorker)
	}

	if strings.TrimSpace(config.ExpectedAssetSetID) != "" {
		if err := simulation.ValidateAssetSetID(config.ExpectedAssetSetID); err != nil {
			return nil, fmt.Errorf("%w: invalid expected asset-set identity", ErrWorker)
		}
	}

	if err := os.MkdirAll(config.StateDirectory, 0o700); err != nil {
		return nil, fmt.Errorf("%w: create worker state directory: %v", ErrWorker, err)
	}

	return &ProcessAllocator{config: config, executable: executable, workers: make(map[string]*processWorker),
		reservations: make(map[string]struct{})}, nil
}

// Allocate contains allocate within the process allocator boundary so callers do not duplicate its domain-specific
// policy.
func (allocator *ProcessAllocator) Allocate(ctx context.Context, spec GameSpec) (WorkerAllocation, error) {
	return allocator.allocate(ctx, spec, nil)
}

// Restore emits the canonical process allocator representation so persisted and transported values retain one stable
// shape.
func (allocator *ProcessAllocator) Restore(
	ctx context.Context,
	spec GameSpec,
	recovery GameRecovery,
) (WorkerAllocation, error) {
	if err := ValidateGameRecovery(recovery); err != nil {
		return WorkerAllocation{}, fmt.Errorf("%w: invalid game recovery", ErrWorker)
	}

	payload, err := json.Marshal(recovery)
	if err != nil || len(payload) == 0 || len(payload) > maximumGameCheckpointBytes {
		return WorkerAllocation{}, fmt.Errorf("%w: game recovery exceeds its bounded format", ErrWorker)
	}

	return allocator.allocate(ctx, spec, &recovery)
}

// allocate coordinates allocate through the owning process allocator synchronization boundary so shared state is
// published only after a complete transition.
func (allocator *ProcessAllocator) allocate(
	ctx context.Context,
	spec GameSpec,
	recovery *GameRecovery,
) (WorkerAllocation, error) {
	if allocator == nil || ctx == nil {
		return WorkerAllocation{}, ErrWorker
	}

	gameID := strings.TrimSpace(spec.GameID)

	allocationID := strings.TrimSpace(spec.AllocationID)
	if spec.Difficulty == "" {
		spec.Difficulty = DifficultyNormal
	}

	if spec.MaximumPlayers == 0 {
		spec.MaximumPlayers = maximumGamePlayers
	}

	if gameID == "" || len(gameID) > 255 || allocationID == "" || len(allocationID) > 255 {
		return WorkerAllocation{}, fmt.Errorf("%w: invalid game ID", ErrWorker)
	}

	allocator.mu.Lock()
	if allocator.closed {
		allocator.mu.Unlock()
		return WorkerAllocation{}, fmt.Errorf("%w: process allocator is closed", ErrWorker)
	}

	if _, found := allocator.workers[gameID]; found {
		allocator.mu.Unlock()
		return WorkerAllocation{}, fmt.Errorf("%w: %s", ErrGameExists, gameID)
	}

	if _, found := allocator.reservations[gameID]; found {
		allocator.mu.Unlock()
		return WorkerAllocation{}, fmt.Errorf("%w: %s", ErrGameExists, gameID)
	}

	allocator.reservations[gameID] = struct{}{}
	allocator.operations.Add(1)

	allocator.mu.Unlock()
	defer allocator.operations.Done()
	defer func() {
		allocator.mu.Lock()
		delete(allocator.reservations, gameID)
		allocator.mu.Unlock()
	}()

	worker, err := allocator.start(ctx, spec, recovery)
	if err != nil {
		return WorkerAllocation{}, err
	}

	allocator.mu.Lock()
	if allocator.closed {
		allocator.mu.Unlock()
		_ = allocator.stop(context.Background(), worker)

		return WorkerAllocation{}, fmt.Errorf("%w: process allocator is closed", ErrWorker)
	}

	if _, exists := allocator.workers[gameID]; exists {
		allocator.mu.Unlock()
		_ = allocator.stop(context.Background(), worker)

		return WorkerAllocation{}, fmt.Errorf("%w: %s", ErrGameExists, gameID)
	}

	allocator.workers[gameID] = worker
	allocator.mu.Unlock()

	return WorkerAllocation{GameID: gameID, AllocationID: allocationID, Worker: worker.client,
		Tickets: worker.client, Endpoint: worker.endpoint}, nil
}

// Game coordinates game through the owning process allocator synchronization boundary so shared state is published
// only after a complete transition.
func (allocator *ProcessAllocator) Game(gameID string) (WorkerClient, bool) {
	if allocator == nil {
		return nil, false
	}

	allocator.mu.RLock()
	worker := allocator.workers[strings.TrimSpace(gameID)]
	allocator.mu.RUnlock()

	if worker == nil {
		return nil, false
	}

	select {
	case <-worker.done:
		return nil, false
	default:
		return worker.client, true
	}
}

// Release coordinates release through the owning process allocator synchronization boundary so shared state is
// published only after a complete transition.
func (allocator *ProcessAllocator) Release(ctx context.Context, gameID string) error {
	if allocator == nil || ctx == nil {
		return ErrWorker
	}

	gameID = strings.TrimSpace(gameID)

	allocator.mu.Lock()

	worker := allocator.workers[gameID]
	if worker != nil {
		delete(allocator.workers, gameID)
	}
	allocator.mu.Unlock()

	if worker == nil {
		return fmt.Errorf("%w: %s", ErrGameNotFound, gameID)
	}

	return allocator.stop(ctx, worker)
}

// Close coordinates close through the owning process allocator synchronization boundary so shared state is published
// only after a complete transition.
func (allocator *ProcessAllocator) Close(ctx context.Context) error {
	if allocator == nil {
		return nil
	}

	if ctx == nil {
		return ErrWorker
	}

	allocator.mu.Lock()
	allocator.closed = true
	allocator.mu.Unlock()
	allocator.operations.Wait()
	allocator.mu.RLock()

	ids := make([]string, 0, len(allocator.workers))
	for id := range allocator.workers {
		ids = append(ids, id)
	}

	allocator.mu.RUnlock()
	sort.Strings(ids)

	var result error
	for _, id := range ids {
		result = errors.Join(result, allocator.Release(ctx, id))
	}

	return result
}

// start prepares per-worker trust material, launches the child, and verifies both readiness ownership and runtime
// identity before returning it. The linear sequence keeps cleanup ownership visible: every failure stops the child or
// removes its private directory, while success transfers both responsibilities to processWorker.
func (allocator *ProcessAllocator) start(
	ctx context.Context,
	spec GameSpec,
	recovery *GameRecovery,
) (*processWorker, error) {
	gameID, allocationID := strings.TrimSpace(spec.GameID), strings.TrimSpace(spec.AllocationID)

	directory, err := os.MkdirTemp(allocator.config.StateDirectory, ".worker-")
	if err != nil {
		return nil, fmt.Errorf("%w: create worker directory: %v", ErrWorker, err)
	}

	cleanup := true
	// Until the process is fully verified, this function owns the private state directory on every return path.
	defer func() {
		if cleanup {
			_ = os.RemoveAll(directory)
		}
	}()

	trust, err := networktrust.New(directory)
	if err != nil {
		return nil, err
	}

	_, clientTLS, fingerprint, err := trust.HostTLS()
	if err != nil {
		return nil, err
	}

	admissionKey, err := writeProcessSecret(directory, "admission-key")
	if err != nil {
		return nil, err
	}

	controlTokenPath, controlToken, err := writeProcessToken(directory, "control-token")
	if err != nil {
		return nil, err
	}

	readyPath := filepath.Join(directory, "ready.json")

	// Copy configured arguments before adding worker-specific flags so repeated allocations never mutate shared config.
	arguments := append([]string(nil), allocator.config.Arguments...)

	if recovery != nil {
		recoveryPath := filepath.Join(directory, "recovery.json")
		if err := writeRealmJSON(recoveryPath, recovery); err != nil {
			return nil, fmt.Errorf("%w: write worker game recovery: %v", ErrWorker, err)
		}

		arguments = append(arguments, "--restore-checkpoint", recoveryPath)
	}

	arguments = append(arguments,
		"--realm-worker",
		"--session-id", gameID,
		"--allocation-id", allocationID,
		"--game-difficulty", string(spec.Difficulty),
		"--game-maximum-players", strconv.Itoa(spec.MaximumPlayers),
		"--quic-listen", allocator.config.GameListenAddress,
		"--tls-cert", filepath.Join(directory, "host-certificate.pem"),
		"--tls-key", filepath.Join(directory, "host-identity.pem"),
		"--admission-key", admissionKey,
		"--worker-control-listen", allocator.config.ControlListenAddress,
		"--worker-control-token", controlTokenPath,
		"--worker-ready-file", readyPath,
	)
	if spec.Hardcore {
		arguments = append(arguments, "--game-hardcore")
	}

	if spec.Ladder {
		arguments = append(arguments, "--game-ladder")
	}

	command := exec.Command(allocator.executable, arguments...)
	command.Dir = allocator.config.WorkingDirectory
	command.Env = append(os.Environ(), allocator.config.Environment...)
	logs := newBoundedProcessLog(maximumWorkerLogBytes, allocator.config.LogWriter)

	command.Stdout, command.Stderr = logs, logs
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("%w: start worker: %v", ErrWorker, err)
	}

	worker := &processWorker{command: command, directory: directory, logs: logs, done: make(chan struct{})}
	// One goroutine owns Wait, publishes its result under the worker lock, and closes done only after publication.
	go func() {
		err := command.Wait()

		worker.mu.Lock()
		worker.waitErr = err
		worker.mu.Unlock()
		close(worker.done)
	}()

	ready, err := allocator.waitReady(ctx, worker, readyPath)
	if err != nil {
		_ = allocator.stop(context.Background(), worker)
		return nil, err
	}

	if !strings.EqualFold(ready.GameEndpoint.TLSFingerprint, fingerprint) {
		_ = allocator.stop(context.Background(), worker)
		return nil, fmt.Errorf("%w: worker readiness TLS identity differs", ErrWorker)
	}

	if ready.GameID != gameID || ready.AllocationID != allocationID || ready.ProcessID != command.Process.Pid {
		_ = allocator.stop(context.Background(), worker)
		return nil, fmt.Errorf("%w: worker readiness ownership differs", ErrWorker)
	}

	// Readiness may advertise a wildcard bind address. Clients need a concrete loopback address unless deployment
	// configuration explicitly supplies the externally reachable host.
	if allocator.config.GameAdvertiseHost != "" {
		_, port, splitErr := net.SplitHostPort(ready.GameEndpoint.Address)
		if splitErr != nil {
			_ = allocator.stop(context.Background(), worker)
			return nil, fmt.Errorf("%w: invalid worker game address", ErrWorker)
		}

		ready.GameEndpoint.Address = net.JoinHostPort(allocator.config.GameAdvertiseHost, port)
	} else if host, port, splitErr := net.SplitHostPort(ready.GameEndpoint.Address); splitErr == nil {
		if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
			if ip.To4() != nil {
				ready.GameEndpoint.Address = net.JoinHostPort("127.0.0.1", port)
			} else {
				ready.GameEndpoint.Address = net.JoinHostPort("::1", port)
			}
		}
	}

	transport := &http.Transport{TLSClientConfig: clientTLS, ForceAttemptHTTP2: true,
		DialContext:     (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		IdleConnTimeout: 30 * time.Second, MaxIdleConnsPerHost: 2}

	client, err := NewWorkerHTTPClient("https://"+ready.ControlAddress, controlToken,
		&http.Client{Transport: transport, Timeout: 10 * time.Second})
	if err != nil {
		_ = allocator.stop(context.Background(), worker)
		return nil, err
	}

	worker.client, worker.endpoint = client, ready.GameEndpoint

	// A ready socket is insufficient: the worker must also prove the exact game and runtime identity requested.
	description, err := client.Describe(ctx)
	if err != nil {
		_ = allocator.stop(context.Background(), worker)
		return nil, fmt.Errorf("%w: verify worker readiness: %v; logs: %s", ErrWorker, err, logs.String())
	}

	digest, err := description.Runtime.Digest()

	expectedAssetSetID := strings.TrimSpace(allocator.config.ExpectedAssetSetID)
	if err != nil || digest != description.IdentityHash || description.GameID != gameID ||
		(expectedAssetSetID != "" && description.Runtime.Recipe.AssetSetID != expectedAssetSetID) {
		_ = allocator.stop(context.Background(), worker)
		return nil, fmt.Errorf("%w: invalid worker runtime identity", ErrWorker)
	}

	cleanup = false

	return worker, nil
}

// waitReady contains wait ready within the process allocator boundary so callers do not duplicate its domain-specific
// policy.
func (allocator *ProcessAllocator) waitReady(
	ctx context.Context,
	worker *processWorker,
	readyPath string,
) (WorkerProcessReady, error) {
	timer := time.NewTimer(allocator.config.StartupTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return WorkerProcessReady{}, ctx.Err()
		case <-timer.C:
			return WorkerProcessReady{}, fmt.Errorf("%w: worker readiness timeout; logs: %s", ErrWorker, worker.logs.String())
		case <-worker.done:
			return WorkerProcessReady{}, fmt.Errorf(
				"%w: worker exited before readiness: %v; logs: %s",
				ErrWorker,
				worker.result(),
				worker.logs.String(),
			)
		case <-ticker.C:
			ready, err := ReadWorkerProcessReady(readyPath)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			if err != nil {
				return WorkerProcessReady{}, fmt.Errorf("%w: read worker readiness: %v", ErrWorker, err)
			}

			return ready, nil
		}
	}
}

// stop owns the process allocator cleanup transition so resources and durable state are retired in the required order.
func (allocator *ProcessAllocator) stop(ctx context.Context, worker *processWorker) error {
	if worker == nil {
		return nil
	}

	shutdownCtx := ctx

	var cancel context.CancelFunc
	if deadline, found := ctx.Deadline(); !found || time.Until(deadline) > allocator.config.ShutdownTimeout {
		shutdownCtx, cancel = context.WithTimeout(context.WithoutCancel(ctx), allocator.config.ShutdownTimeout)
		defer cancel()
	}

	var result error

	select {
	case <-worker.done:
	default:
		if worker.client != nil {
			result = worker.client.Close(shutdownCtx)
		}
	}

	select {
	case <-worker.done:
	case <-shutdownCtx.Done():
		if worker.command != nil && worker.command.Process != nil {
			result = errors.Join(result, worker.command.Process.Kill())
		}

		<-worker.done
	}

	if waitErr := worker.result(); waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) || exitError.ExitCode() != -1 {
			result = errors.Join(result, fmt.Errorf("worker process: %w", waitErr))
		}
	}

	if err := os.RemoveAll(worker.directory); err != nil {
		result = errors.Join(result, err)
	}

	return result
}

// result coordinates result through the owning process allocator synchronization boundary so shared state is published
// only after a complete transition.
func (worker *processWorker) result() error {
	worker.mu.Lock()
	defer worker.mu.Unlock()

	return worker.waitErr
}

// writeProcessSecret emits the canonical process allocator representation so persisted and transported values retain
// one stable shape.
func writeProcessSecret(directory, name string) (string, error) {
	path, _, err := writeProcessToken(directory, name)
	return path, err
}

// writeProcessToken emits the canonical process allocator representation so persisted and transported values retain
// one stable shape.
func writeProcessToken(directory, name string) (string, string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(data)

	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(token), 0o600); err != nil {
		return "", "", err
	}

	return path, token, nil
}

type boundedProcessLog struct {
	mu       sync.Mutex
	maximum  int
	data     []byte
	external io.Writer
}

// newBoundedProcessLog constructs the process allocator boundary and validates dependencies before callers can publish
// or mutate shared state.
func newBoundedProcessLog(maximum int, external io.Writer) *boundedProcessLog {
	return &boundedProcessLog{maximum: maximum, external: external}
}

// Write emits the canonical process allocator representation so persisted and transported values retain one stable
// shape.
func (log *boundedProcessLog) Write(data []byte) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.external != nil {
		_, _ = log.external.Write(data)
	}

	log.data = append(log.data, data...)
	if len(log.data) > log.maximum {
		log.data = append([]byte(nil), log.data[len(log.data)-log.maximum:]...)
	}

	return len(data), nil
}

// String coordinates string through the owning process allocator synchronization boundary so shared state is published
// only after a complete transition.
func (log *boundedProcessLog) String() string {
	log.mu.Lock()
	defer log.mu.Unlock()

	return strings.TrimSpace(string(log.data))
}
