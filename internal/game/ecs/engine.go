// Package ecs owns Dark Magic's deterministic entity simulation schedule.
package ecs

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gravestench/akara"
)

var (
	// Registration errors are explicit so script adapters can distinguish bad
	// definitions from runtime system failures.
	ErrSystemID       = errors.New("game ecs: system ID is required")
	ErrSystemExists   = errors.New("game ecs: system already exists")
	ErrSystemNotFound = errors.New("game ecs: system dependency does not exist")
	ErrSystemCycle    = errors.New("game ecs: system ordering cycle")
	ErrSystemPhase    = errors.New("game ecs: invalid system phase")
	ErrNilUpdate      = errors.New("game ecs: system update is nil")
)

const (
	// DefaultStep is the host's default fixed-tick cadence. Mods may select a
	// different explicit cadence when constructing their session.
	DefaultStep = 40 * time.Millisecond
	// DefaultMaxCatchUp bounds recovery after a slow host frame.
	DefaultMaxCatchUp = 5
)

// Phase is one deterministic simulation barrier.
type Phase string

const (
	// Phases make cross-system ordering semantic and stable across registration
	// order. Structural command buffers flush at each phase barrier.
	PhaseInput        Phase = "input"
	PhaseIntent       Phase = "intent"
	PhasePreSimulate  Phase = "pre_simulation"
	PhaseMovement     Phase = "movement"
	PhaseCollision    Phase = "collision"
	PhaseCombat       Phase = "combat"
	PhaseEffects      Phase = "effects"
	PhaseInventory    Phase = "inventory"
	PhasePresentation Phase = "presentation"
	PhaseCleanup      Phase = "cleanup"
)

var phases = []Phase{
	PhaseInput, PhaseIntent, PhasePreSimulate, PhaseMovement, PhaseCollision,
	PhaseCombat, PhaseEffects, PhaseInventory, PhasePresentation, PhaseCleanup,
}

var phaseIndex = func() map[Phase]int {
	result := make(map[Phase]int, len(phases))
	for index, phase := range phases {
		result[phase] = index
	}
	return result
}()

// Context is the deterministic input supplied to one system update.
type Context struct {
	Tick  uint64
	Delta time.Duration
}

// UpdateFunc executes against a stable entity snapshot. Structural mutations
// must be submitted to commands and are applied after the callback returns.
type UpdateFunc func(Context, []akara.Entity, *StructuralCommands) error

// StructuralCommands allocates Akara's synchronized command buffer only when a
// system actually submits a structural mutation. Most steady-state systems only
// read or update component values and therefore remain allocation-free here.
type StructuralCommands struct{ buffer *akara.CommandBuffer }

func (commands *StructuralCommands) native() *akara.CommandBuffer {
	if commands.buffer == nil {
		commands.buffer = akara.NewCommandBuffer()
	}
	return commands.buffer
}

func (commands *StructuralCommands) CreateDynamic(world *akara.World, components map[*akara.DynamicStore]map[string]any) {
	commands.native().CreateDynamic(world, components)
}

func (commands *StructuralCommands) AddDynamic(store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	commands.native().AddDynamic(store, entity, values)
}

func (commands *StructuralCommands) Remove(component akara.ComponentType, entity akara.Entity) {
	commands.native().Remove(component, entity)
}

func (commands *StructuralCommands) Destroy(world *akara.World, entity akara.Entity) {
	commands.native().Destroy(world, entity)
}

func (commands *StructuralCommands) apply() error {
	if commands.buffer == nil {
		return nil
	}
	err := commands.buffer.Apply()
	commands.buffer = nil
	return err
}

// Definition declares one ordered ECS system and its component access contract.
type Definition struct {
	ID     string
	Phase  Phase
	After  []string
	Before []string
	All    []akara.ComponentType
	Any    []akara.ComponentType
	None   []akara.ComponentType
	Read   []akara.ComponentType
	Write  []akara.ComponentType
	Update UpdateFunc
}

type registeredSystem struct {
	definition   Definition
	subscription *akara.Subscription
	commands     StructuralCommands
}

type compiledSchedule struct{ systems []*registeredSystem }

// Engine owns one Akara world and a deterministic, dependency-ordered schedule.
type Engine struct {
	world    *akara.World
	mu       sync.RWMutex
	systems  map[string]*registeredSystem
	order    []*registeredSystem
	schedule atomic.Pointer[compiledSchedule]
	tick     uint64
	step     time.Duration
	lag      time.Duration
	maxSteps int
	updateMu sync.Mutex
}

// New creates an owned Akara world with the production fixed-step policy.
func New() *Engine {
	return NewWithClock(DefaultStep, DefaultMaxCatchUp)
}

// NewWithClock constructs an engine with an explicit fixed simulation step and
// catch-up limit. The limit prevents a stalled renderer from causing an
// unbounded update spiral.
func NewWithClock(step time.Duration, maxCatchUp int) *Engine {
	if step <= 0 {
		step = DefaultStep
	}
	if maxCatchUp <= 0 {
		maxCatchUp = DefaultMaxCatchUp
	}
	engine := &Engine{
		world: akara.NewWorld(), systems: make(map[string]*registeredSystem),
		step: step, maxSteps: maxCatchUp,
	}
	engine.schedule.Store(&compiledSchedule{})
	return engine
}

// World exposes component storage to trusted game systems and capability
// adapters. It does not transfer engine lifetime or scheduling ownership.
func (engine *Engine) World() *akara.World { return engine.world }

// Tick returns the number of completed/started deterministic updates.
func (engine *Engine) Tick() uint64 {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	return engine.tick
}

// Register validates and atomically adds a system. Failed ordering leaves the
// existing schedule unchanged.
func (engine *Engine) Register(definition Definition) error {
	if definition.ID == "" {
		return ErrSystemID
	}
	if definition.Update == nil {
		return ErrNilUpdate
	}
	if _, valid := phaseIndex[definition.Phase]; !valid {
		return fmt.Errorf("%w: %q", ErrSystemPhase, definition.Phase)
	}
	options := filterOptions(definition)
	subscription, err := engine.world.Subscribe(options...)
	if err != nil {
		return err
	}
	entry := &registeredSystem{definition: cloneDefinition(definition), subscription: subscription}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if _, exists := engine.systems[definition.ID]; exists {
		_ = subscription.Close()
		return fmt.Errorf("%w: %q", ErrSystemExists, definition.ID)
	}
	candidate := make(map[string]*registeredSystem, len(engine.systems)+1)
	for id, system := range engine.systems {
		candidate[id] = system
	}
	candidate[definition.ID] = entry
	order, err := compileOrder(candidate)
	if err != nil {
		_ = subscription.Close()
		return err
	}
	engine.systems = candidate
	engine.order = order
	engine.schedule.Store(&compiledSchedule{systems: order})
	return nil
}

func filterOptions(definition Definition) []akara.FilterOption {
	options := make([]akara.FilterOption, 0, 3)
	if len(definition.All) > 0 {
		options = append(options, akara.All(definition.All...))
	}
	if len(definition.Any) > 0 {
		options = append(options, akara.Any(definition.Any...))
	}
	if len(definition.None) > 0 {
		options = append(options, akara.None(definition.None...))
	}
	return options
}

// Unregister removes a system and closes its query subscription.
func (engine *Engine) Unregister(id string) bool {
	engine.mu.Lock()
	entry, found := engine.systems[id]
	if !found {
		engine.mu.Unlock()
		return false
	}
	candidate := make(map[string]*registeredSystem, len(engine.systems)-1)
	for existingID, system := range engine.systems {
		if existingID != id {
			candidate[existingID] = system
		}
	}
	order, err := compileOrder(candidate)
	if err != nil {
		engine.mu.Unlock()
		return false
	}
	engine.systems, engine.order = candidate, order
	engine.schedule.Store(&compiledSchedule{systems: order})
	engine.mu.Unlock()
	_ = entry.subscription.Close()
	return true
}

// Update runs each phase and system sequentially against stable query snapshots.
func (engine *Engine) Update(delta time.Duration) error {
	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()
	return engine.update(delta)
}

// Advance accumulates wall-clock time and executes zero or more fixed updates.
// Excess lag is discarded at the configured catch-up boundary.
func (engine *Engine) Advance(elapsed time.Duration) (int, error) {
	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()
	if elapsed < 0 {
		elapsed = 0
	}
	maximumLag := engine.step * time.Duration(engine.maxSteps)
	engine.lag = min(engine.lag+elapsed, maximumLag)
	steps := 0
	for engine.lag >= engine.step {
		if err := engine.update(engine.step); err != nil {
			return steps, err
		}
		engine.lag -= engine.step
		steps++
	}
	return steps, nil
}

func (engine *Engine) update(delta time.Duration) error {
	engine.mu.Lock()
	engine.tick++
	context := Context{Tick: engine.tick, Delta: delta}
	engine.mu.Unlock()
	for _, system := range engine.schedule.Load().systems {
		commands := &system.commands
		var entities []akara.Entity
		if system.subscription.Len() > 0 {
			entities = system.subscription.Entities()
		}
		if err := system.definition.Update(context, entities, commands); err != nil {
			commands.buffer = nil
			return fmt.Errorf("game ecs: update %q: %w", system.definition.ID, err)
		}
		if err := commands.apply(); err != nil {
			return fmt.Errorf("game ecs: apply %q structural changes: %w", system.definition.ID, err)
		}
	}
	return nil
}

func (engine *Engine) Systems() []string {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	result := make([]string, len(engine.order))
	for index, system := range engine.order {
		result[index] = system.definition.ID
	}
	return result
}

func (engine *Engine) Close() error {
	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()
	engine.mu.Lock()
	for _, system := range engine.systems {
		_ = system.subscription.Close()
	}
	engine.systems = make(map[string]*registeredSystem)
	engine.order = nil
	engine.schedule.Store(&compiledSchedule{})
	engine.mu.Unlock()
	return engine.world.Close()
}

func cloneDefinition(definition Definition) Definition {
	definition.After = append([]string(nil), definition.After...)
	definition.Before = append([]string(nil), definition.Before...)
	definition.All = append([]akara.ComponentType(nil), definition.All...)
	definition.Any = append([]akara.ComponentType(nil), definition.Any...)
	definition.None = append([]akara.ComponentType(nil), definition.None...)
	definition.Read = append([]akara.ComponentType(nil), definition.Read...)
	definition.Write = append([]akara.ComponentType(nil), definition.Write...)
	return definition
}

func compileOrder(systems map[string]*registeredSystem) ([]*registeredSystem, error) {
	indegree := make(map[string]int, len(systems))
	edges := make(map[string]map[string]struct{}, len(systems))
	addEdge := func(from, to string) {
		if edges[from] == nil {
			edges[from] = make(map[string]struct{})
		}
		if _, exists := edges[from][to]; !exists {
			edges[from][to] = struct{}{}
			indegree[to]++
		}
	}
	for id := range systems {
		indegree[id] = 0
	}
	for id, system := range systems {
		for _, dependency := range system.definition.After {
			if _, exists := systems[dependency]; !exists {
				return nil, fmt.Errorf("%w: %q after %q", ErrSystemNotFound, id, dependency)
			}
			addEdge(dependency, id)
		}
		for _, dependency := range system.definition.Before {
			if _, exists := systems[dependency]; !exists {
				return nil, fmt.Errorf("%w: %q before %q", ErrSystemNotFound, id, dependency)
			}
			addEdge(id, dependency)
		}
	}
	ids := make([]string, 0, len(systems))
	for id := range systems {
		ids = append(ids, id)
	}
	for _, from := range ids {
		for _, to := range ids {
			if from == to {
				continue
			}
			if phaseIndex[systems[from].definition.Phase] < phaseIndex[systems[to].definition.Phase] {
				addEdge(from, to)
			}
		}
	}
	ready := make([]string, 0)
	for id, degree := range indegree {
		if degree == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	order := make([]*registeredSystem, 0, len(systems))
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		order = append(order, systems[id])
		for target := range edges[id] {
			indegree[target]--
			if indegree[target] == 0 {
				ready = append(ready, target)
				sort.Strings(ready)
			}
		}
	}
	if len(order) != len(systems) {
		return nil, ErrSystemCycle
	}
	return order, nil
}
