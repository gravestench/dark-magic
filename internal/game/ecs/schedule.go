package ecs

import (
	"fmt"
	"sort"
)

// Phase identifies one deterministic simulation barrier. Structural command buffers flush after every system,
// allowing later systems and phases to observe mutations without exposing partial updates to the producer.
type Phase string

const (
	// The phase sequence makes cross-system ordering semantic and stable regardless of registration order.
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

var phaseIndex = buildPhaseIndex()

// buildPhaseIndex converts the canonical phase sequence into ranks used by the dependency compiler. Keeping this map
// derived from phases prevents validation and ordering from drifting apart when a phase is added.
func buildPhaseIndex() map[Phase]int {
	result := make(map[Phase]int, len(phases))

	for index, phase := range phases {
		result[phase] = index
	}

	return result
}

// compileOrder performs a stable topological sort over explicit dependencies and implicit phase barriers. It returns
// no partial schedule on failure, allowing registration and removal to remain transactional.
func compileOrder(systems map[string]*registeredSystem) ([]*registeredSystem, error) {
	indegree := make(map[string]int, len(systems))
	edges := make(map[string]map[string]struct{}, len(systems))

	// Duplicate constraints must not inflate indegrees or make an otherwise valid schedule look cyclic.
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

	// Phase edges make every earlier-phase system precede every later-phase system, even when callers did not declare
	// explicit dependencies between them.
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

	// Lexical tie-breaking makes independent systems deterministic across Go's randomized map iteration.
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
