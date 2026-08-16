// Package realm owns game-worker allocation and lifetime, not gameplay state.
package realm

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

var (
	ErrGameExists   = errors.New("realm: game already exists")
	ErrGameNotFound = errors.New("realm: game not found")
)

type workerFactory func(context.Context, gameserver.Config) (WorkerClient, error)

// Manager is the in-process allocator and stable-ID worker registry. Admissions
// sees only WorkerClient; child-process and cluster allocators can implement the
// same registry without exposing gameserver.Host.
type Manager struct {
	mu      sync.RWMutex
	start   workerFactory
	workers map[string]WorkerClient
}

func NewManager(source fs.FS, records d2legacy.Records) (*Manager, error) {
	if source == nil || records == nil {
		return nil, errors.New("realm: content and records are required")
	}
	return &Manager{start: func(ctx context.Context, config gameserver.Config) (WorkerClient, error) {
		host, err := gameserver.Start(ctx, source, records, config)
		if err != nil {
			return nil, err
		}
		worker, err := newInProcessWorker(host)
		if err != nil {
			return nil, errors.Join(err, host.Close(context.Background()))
		}
		return worker, nil
	}, workers: make(map[string]WorkerClient)}, nil
}

func (manager *Manager) Allocate(ctx context.Context, config gameserver.Config) (WorkerClient, error) {
	if manager == nil {
		return nil, errors.New("realm: nil manager")
	}
	sessionID := strings.TrimSpace(config.SessionID)
	if sessionID == "" {
		return nil, errors.New("realm: session ID is required")
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if _, exists := manager.workers[sessionID]; exists {
		return nil, fmt.Errorf("%w: %s", ErrGameExists, sessionID)
	}
	config.Mode = gameserver.ModeRealm
	worker, err := manager.start(ctx, config)
	if err != nil {
		return nil, err
	}
	if worker == nil {
		return nil, ErrWorker
	}
	manager.workers[sessionID] = worker
	return worker, nil
}

func (manager *Manager) Game(sessionID string) (WorkerClient, bool) {
	if manager == nil {
		return nil, false
	}
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	worker, found := manager.workers[strings.TrimSpace(sessionID)]
	return worker, found
}

func (manager *Manager) Release(ctx context.Context, sessionID string) error {
	if manager == nil {
		return errors.New("realm: nil manager")
	}
	manager.mu.Lock()
	worker, found := manager.workers[strings.TrimSpace(sessionID)]
	if found {
		delete(manager.workers, strings.TrimSpace(sessionID))
	}
	manager.mu.Unlock()
	if !found {
		return fmt.Errorf("%w: %s", ErrGameNotFound, sessionID)
	}
	return worker.Close(ctx)
}

func (manager *Manager) Close(ctx context.Context) error {
	if manager == nil {
		return nil
	}
	manager.mu.Lock()
	ids := make([]string, 0, len(manager.workers))
	for id := range manager.workers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	workers := make([]WorkerClient, 0, len(ids))
	for _, id := range ids {
		workers = append(workers, manager.workers[id])
		delete(manager.workers, id)
	}
	manager.mu.Unlock()
	var result error
	for _, worker := range workers {
		result = errors.Join(result, worker.Close(ctx))
	}
	return result
}
