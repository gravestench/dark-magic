// Package interaction owns authoritative player-to-world interaction context.
// Presentation may request a target, but only this package resolves which NPC,
// vendor, categories, and services that target actually exposes.
package interaction

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

type Target struct {
	ID           string   `json:"id"`
	NPC          string   `json:"npc"`
	Vendor       string   `json:"vendor,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	Services     []string `json:"services,omitempty"`
	X            float64  `json:"x"`
	Y            float64  `json:"y"`
	Radius       float64  `json:"radius"`
	SelectRadius float64  `json:"select_radius,omitempty"`
}

type Context struct {
	TargetID   string   `json:"target_id,omitempty"`
	NPC        string   `json:"npc,omitempty"`
	Vendor     string   `json:"vendor,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Services   []string `json:"services,omitempty"`
}

type Authority struct {
	mu       sync.RWMutex
	targets  map[string]Target
	owners   map[string]Context
	selector *gameworld.Selector
	world    *gameworld.Map
}

func NewAuthority(targets ...Target) (*Authority, error) {
	authority := &Authority{targets: make(map[string]Target, len(targets)), owners: make(map[string]Context)}
	if err := authority.AddTargets(targets...); err != nil {
		return nil, err
	}
	return authority, nil
}

// AddTargets materializes server-known selectable entities before simulation
// starts. Pointer commands submit coordinates; this registry chooses the ID.
func (authority *Authority) AddTargets(targets ...Target) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	for _, target := range targets {
		target = normalizeTarget(target)
		if target.SelectRadius == 0 {
			target.SelectRadius = 1.5
		}
		if target.ID == "" || target.NPC == "" || target.Radius <= 0 || target.SelectRadius <= 0 || !finite(target.X) || !finite(target.Y) || !finite(target.Radius) || !finite(target.SelectRadius) {
			return fmt.Errorf("interaction: target ID, NPC, and positive range/select radii are required")
		}
		if _, exists := authority.targets[target.ID]; exists {
			return fmt.Errorf("interaction: duplicate target %q", target.ID)
		}
		authority.targets[target.ID] = target
	}
	return authority.rebuildSelectorLocked()
}

func (authority *Authority) ConfigureWorld(world *gameworld.Map) {
	authority.mu.Lock()
	authority.world = world
	authority.mu.Unlock()
}

func (authority *Authority) rebuildSelectorLocked() error {
	items := make([]gameworld.Selectable, 0, len(authority.targets))
	for _, target := range authority.targets {
		items = append(items, gameworld.Selectable{ID: target.ID, Kind: "interaction", X: target.X, Y: target.Y, Radius: target.SelectRadius})
	}
	selector, err := gameworld.NewSelector(items, 8)
	if err != nil {
		return err
	}
	authority.selector = selector
	return nil
}

func (authority *Authority) targetAt(x, y float64) (Target, error) {
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	selected, found := authority.selector.Hit(x, y)
	if !found {
		return Target{}, fmt.Errorf("interaction: no selectable target at %.2f,%.2f", x, y)
	}
	return authority.targets[selected.ID], nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

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

// CanTradeAt and CanServiceAt are the item authority boundary. They re-check
// current ECS position for every transaction, so walking away invalidates an
// already-open panel without trusting presentation to close it first.
func (authority *Authority) CanTradeAt(engine *gameecs.Engine, owner, vendor string) bool {
	context, err := authority.Snapshot(owner)
	return err == nil && context.Vendor != "" && strings.EqualFold(context.Vendor, strings.TrimSpace(vendor)) && authority.validateRange(engine, owner, context.TargetID) == nil
}

func (authority *Authority) CanServiceAt(engine *gameecs.Engine, owner, service string) bool {
	context, err := authority.Snapshot(owner)
	if err != nil || authority.validateRange(engine, owner, context.TargetID) != nil {
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
