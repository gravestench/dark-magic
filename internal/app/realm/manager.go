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

type gameFactory func(context.Context, fs.FS, d2legacy.Records, gameserver.Config) (*gameserver.Host, error)

// Manager owns a stable-ID registry of realm-hosted game workers. Each worker
// is the same gameserver.Host used by standalone and future listen modes.
type Manager struct {
	mu      sync.RWMutex
	source  fs.FS
	records d2legacy.Records
	start   gameFactory
	hosts   map[string]*gameserver.Host
}

func NewManager(source fs.FS, records d2legacy.Records) (*Manager, error) {
	if source == nil || records == nil {
		return nil, errors.New("realm: content and records are required")
	}
	return &Manager{source: source, records: records, start: gameserver.Start,
		hosts: make(map[string]*gameserver.Host)}, nil
}

func (manager *Manager) Allocate(ctx context.Context, config gameserver.Config) (*gameserver.Host, error) {
	if manager == nil {
		return nil, errors.New("realm: nil manager")
	}
	sessionID := strings.TrimSpace(config.SessionID)
	if sessionID == "" {
		return nil, errors.New("realm: session ID is required")
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if _, exists := manager.hosts[sessionID]; exists {
		return nil, fmt.Errorf("%w: %s", ErrGameExists, sessionID)
	}
	config.Mode = gameserver.ModeRealm
	host, err := manager.start(ctx, manager.source, manager.records, config)
	if err != nil {
		return nil, err
	}
	manager.hosts[sessionID] = host
	return host, nil
}

func (manager *Manager) Game(sessionID string) (*gameserver.Host, bool) {
	if manager == nil {
		return nil, false
	}
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	host, found := manager.hosts[strings.TrimSpace(sessionID)]
	return host, found
}

func (manager *Manager) Release(ctx context.Context, sessionID string) error {
	if manager == nil {
		return errors.New("realm: nil manager")
	}
	manager.mu.Lock()
	host, found := manager.hosts[strings.TrimSpace(sessionID)]
	if found {
		delete(manager.hosts, strings.TrimSpace(sessionID))
	}
	manager.mu.Unlock()
	if !found {
		return fmt.Errorf("%w: %s", ErrGameNotFound, sessionID)
	}
	return host.Close(ctx)
}

func (manager *Manager) Close(ctx context.Context) error {
	if manager == nil {
		return nil
	}
	manager.mu.Lock()
	ids := make([]string, 0, len(manager.hosts))
	for id := range manager.hosts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	hosts := make([]*gameserver.Host, 0, len(ids))
	for _, id := range ids {
		hosts = append(hosts, manager.hosts[id])
		delete(manager.hosts, id)
	}
	manager.mu.Unlock()
	var result error
	for _, host := range hosts {
		result = errors.Join(result, host.Close(ctx))
	}
	return result
}
