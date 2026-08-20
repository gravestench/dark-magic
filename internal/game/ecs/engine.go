// Package ecs owns Dark Magic's deterministic entity simulation schedule.
package ecs

import (
	"errors"
	"fmt"
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

// native lazily allocates the underlying Akara command buffer. Systems that only mutate component values therefore do
// not pay for structural-mutation bookkeeping.
func (commands *StructuralCommands) native() *akara.CommandBuffer {
	if commands.buffer == nil {
		commands.buffer = akara.NewCommandBuffer()
	}

	return commands.buffer
}

// CreateDynamic queues an entity creation for the current system barrier, so the producer's entity query remains
// stable for the duration of its update.
func (commands *StructuralCommands) CreateDynamic(
	world *akara.World,
	components map[*akara.DynamicStore]map[string]any,
) {
	commands.native().CreateDynamic(world, components)
}

// AddDynamic queues a dynamic component addition for the current system barrier. Later systems observe the addition,
// while the producer cannot invalidate the entity slice it is iterating.
func (commands *StructuralCommands) AddDynamic(store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	commands.native().AddDynamic(store, entity, values)
}

// Remove queues component removal until the current system returns, preserving stable query membership during update.
func (commands *StructuralCommands) Remove(component akara.ComponentType, entity akara.Entity) {
	commands.native().Remove(component, entity)
}

// Destroy queues entity destruction until the current system returns, preventing iterator invalidation in the producer.
func (commands *StructuralCommands) Destroy(world *akara.World, entity akara.Entity) {
	commands.native().Destroy(world, entity)
}

// apply flushes one system's structural mutations before the next system runs. Clearing the buffer even when Apply
// fails prevents a later tick from replaying a partially attempted command batch.
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

// New creates an engine that owns an Akara world and uses the production fixed-step policy. Callers must Close the
// engine to release world subscriptions and storage.
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
		world:    akara.NewWorld(),
		systems:  make(map[string]*registeredSystem),
		step:     step,
		maxSteps: maxCatchUp,
	}
	engine.schedule.Store(&compiledSchedule{})

	return engine
}

// World exposes component storage to trusted game systems and capability
// adapters. It does not transfer engine lifetime or scheduling ownership.
func (engine *Engine) World() *akara.World { return engine.world }

// Tick returns the number of deterministic updates that have begun. A system failure still consumes its tick because
// callbacks may already have observed that value before returning the error.
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

	// Compile against a copy so an invalid dependency graph cannot disturb the live schedule.
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
	// Readers use the atomic schedule without holding the registration mutex throughout callbacks.
	engine.schedule.Store(&compiledSchedule{systems: order})

	return nil
}

// filterOptions translates a system's declarative membership constraints into one Akara subscription. Read and Write
// declare access intent for callers but do not change which entities belong to the query.
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

// Unregister atomically removes a system and closes its query subscription. A dependency failure leaves the live
// schedule unchanged and reports false because removing the system would invalidate remaining registrations.
func (engine *Engine) Unregister(id string) bool {
	engine.mu.Lock()

	entry, found := engine.systems[id]
	if !found {
		engine.mu.Unlock()

		return false
	}

	// Compile a replacement schedule before publishing removal, matching Register's transactional behavior.
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

	// Closing after publication prevents new updates from selecting the removed system while keeping the critical
	// section independent of subscription cleanup.
	_ = entry.subscription.Close()

	return true
}

// Update serializes one caller-supplied simulation delta and runs every system against stable query snapshots. Unlike
// Advance, it neither quantizes nor caps the supplied duration.
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

	// Discard lag beyond the catch-up budget so a stalled host cannot trap the simulation in an update spiral.
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

// update executes one already-serialized simulation step. The tick advances before callbacks and remains advanced on
// failure, preserving the tick value observed by the failing system and matching the engine's historical error model.
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
			// A failed system must not leak its queued structural mutations into a later tick.
			commands.buffer = nil

			return fmt.Errorf("game ecs: update %q: %w", system.definition.ID, err)
		}

		// The per-system barrier makes structural changes visible to the next scheduled system.
		if err := commands.apply(); err != nil {
			return fmt.Errorf("game ecs: apply %q structural changes: %w", system.definition.ID, err)
		}
	}

	return nil
}

// Systems returns a defensive copy of the current deterministic schedule, allowing diagnostics to inspect order
// without retaining mutable engine state.
func (engine *Engine) Systems() []string {
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	result := make([]string, len(engine.order))
	for index, system := range engine.order {
		result[index] = system.definition.ID
	}

	return result
}

// Close waits for any active update, detaches all subscriptions, and closes the owned world. Acquiring updateMu before
// mu preserves the same lock order used by Restore and prevents teardown from racing a system callback.
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

// cloneDefinition defensively copies caller-owned slices so later caller mutations cannot silently change filters or
// ordering constraints after a system has been registered.
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
