// Package interaction owns authoritative player-to-world interaction context.
// Presentation may request a target, but only this package resolves which NPC,
// vendor, categories, and services that target actually exposes.
package interaction

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Target struct {
	ID         string   `json:"id"`
	NPC        string   `json:"npc"`
	Vendor     string   `json:"vendor,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Services   []string `json:"services,omitempty"`
}

type Context struct {
	TargetID   string   `json:"target_id,omitempty"`
	NPC        string   `json:"npc,omitempty"`
	Vendor     string   `json:"vendor,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Services   []string `json:"services,omitempty"`
}

type Authority struct {
	mu      sync.RWMutex
	targets map[string]Target
	owners  map[string]Context
}

func NewAuthority(targets ...Target) (*Authority, error) {
	authority := &Authority{targets: make(map[string]Target, len(targets)), owners: make(map[string]Context)}
	for _, target := range targets {
		target = normalizeTarget(target)
		if target.ID == "" || target.NPC == "" {
			return nil, fmt.Errorf("interaction: target ID and NPC are required")
		}
		if _, exists := authority.targets[target.ID]; exists {
			return nil, fmt.Errorf("interaction: duplicate target %q", target.ID)
		}
		authority.targets[target.ID] = target
	}
	return authority, nil
}

// RegisterOwner installs optional initial context before the session captures
// replay state. An empty target leaves the player outside an interaction.
func (authority *Authority) RegisterOwner(owner, targetID string) error {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("interaction: owner is required")
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, exists := authority.owners[owner]; exists {
		return fmt.Errorf("interaction: owner %q is already registered", owner)
	}
	if strings.TrimSpace(targetID) == "" {
		authority.owners[owner] = Context{}
		return nil
	}
	context, err := authority.resolveLocked(targetID)
	if err != nil {
		return err
	}
	authority.owners[owner] = context
	return nil
}

func (authority *Authority) Snapshot(owner string) (Context, error) {
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	context, found := authority.owners[strings.TrimSpace(owner)]
	if !found {
		return Context{}, fmt.Errorf("interaction: unknown owner %q", owner)
	}
	return cloneContext(context), nil
}

func (authority *Authority) open(owner, targetID string) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, found := authority.owners[owner]; !found {
		return fmt.Errorf("interaction: unknown owner %q", owner)
	}
	context, err := authority.resolveLocked(targetID)
	if err != nil {
		return err
	}
	authority.owners[owner] = context
	return nil
}

func (authority *Authority) close(owner string) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, found := authority.owners[owner]; !found {
		return fmt.Errorf("interaction: unknown owner %q", owner)
	}
	authority.owners[owner] = Context{}
	return nil
}

func (authority *Authority) CanTrade(owner, vendor string) bool {
	context, err := authority.Snapshot(owner)
	return err == nil && context.Vendor != "" && strings.EqualFold(context.Vendor, strings.TrimSpace(vendor))
}

func (authority *Authority) CanService(owner, service string) bool {
	context, err := authority.Snapshot(owner)
	if err != nil {
		return false
	}
	service = strings.ToLower(strings.TrimSpace(service))
	index := sort.SearchStrings(context.Services, service)
	return index < len(context.Services) && context.Services[index] == service
}

func (authority *Authority) resolveLocked(targetID string) (Context, error) {
	targetID = strings.ToLower(strings.TrimSpace(targetID))
	target, found := authority.targets[targetID]
	if !found {
		return Context{}, fmt.Errorf("interaction: unknown target %q", targetID)
	}
	return Context{TargetID: target.ID, NPC: target.NPC, Vendor: target.Vendor, Categories: append([]string(nil), target.Categories...), Services: append([]string(nil), target.Services...)}, nil
}

func normalizeTarget(target Target) Target {
	target.ID = strings.ToLower(strings.TrimSpace(target.ID))
	target.NPC = strings.TrimSpace(target.NPC)
	target.Vendor = strings.TrimSpace(target.Vendor)
	target.Categories = normalizeNames(target.Categories)
	target.Services = normalizeNames(target.Services)
	return target
}

func normalizeNames(names []string) []string {
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func cloneContext(context Context) Context {
	context.Categories = append([]string(nil), context.Categories...)
	context.Services = append([]string(nil), context.Services...)
	return context
}
