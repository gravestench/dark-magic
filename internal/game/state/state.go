// Package state owns source-tagged authoritative timed state instances.
package state

import (
	"fmt"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const (
	SystemID          = "state.timed_instances"
	RequestComponent  = "d2.state.request"
	InstanceComponent = "d2.state.instance"
	EventComponent    = "d2.state.event"

	OperationApply    = "apply"
	OperationRemove   = "remove"
	RefreshSameSource = "refresh_same_source"

	EventApplied   = "state_applied"
	EventRefreshed = "state_refreshed"
	EventRemoved   = "state_removed"
)

func requestSchema() akara.Schema {
	return akara.Schema{Name: RequestComponent, Version: 1, Fields: []akara.Field{
		{Name: "operation", Kind: akara.FieldString}, {Name: "target", Kind: akara.FieldEntity},
		{Name: "state_id", Kind: akara.FieldString}, {Name: "source_id", Kind: akara.FieldString},
		{Name: "duration", Kind: akara.FieldInt64}, {Name: "policy", Kind: akara.FieldString},
	}}
}

func instanceSchema() akara.Schema {
	return akara.Schema{Name: InstanceComponent, Version: 1, Fields: []akara.Field{
		{Name: "target", Kind: akara.FieldEntity}, {Name: "state_id", Kind: akara.FieldString},
		{Name: "source_id", Kind: akara.FieldString}, {Name: "applied_tick", Kind: akara.FieldInt64},
		{Name: "expires_tick", Kind: akara.FieldInt64}, {Name: "policy", Kind: akara.FieldString},
	}}
}

func eventSchema() akara.Schema {
	return akara.Schema{Name: EventComponent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64},
		{Name: "target", Kind: akara.FieldEntity}, {Name: "state_id", Kind: akara.FieldString},
		{Name: "source_id", Kind: akara.FieldString}, {Name: "expires_tick", Kind: akara.FieldInt64},
		{Name: "reason", Kind: akara.FieldString},
	}}
}

// Register installs deterministic apply/refresh/remove/expiration processing.
// Every future-affecting fact lives in dynamic ECS state and is checkpointed.
func Register(engine *gameecs.Engine) error {
	if engine == nil {
		return fmt.Errorf("state: engine is required")
	}
	requests, instances, events, err := registerStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: SystemID, Phase: gameecs.PhaseEffects,
		Any: []akara.ComponentType{requests, instances}, Read: []akara.ComponentType{requests, instances}, Write: []akara.ComponentType{requests, instances, events},
		Update: func(context gameecs.Context, _ []akara.Entity, commands *akara.CommandBuffer) error {
			touched, err := applyRequests(context, commands, engine.World(), requests, instances, events)
			if err != nil {
				return err
			}
			expireInstances(context, commands, engine.World(), instances, events, touched)
			return nil
		},
	})
}

func registerStores(engine *gameecs.Engine) (requests, instances, events *akara.DynamicStore, err error) {
	schemas := []akara.Schema{requestSchema(), instanceSchema(), eventSchema()}
	stores := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		stores[index], err = akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return stores[0], stores[1], stores[2], nil
}

type requestKey struct {
	target   akara.Entity
	stateID  string
	sourceID string
}

type requestPlan struct {
	operation string
	duration  int64
	policy    string
}

// Apply queues a timed state. Duration is measured in authoritative ticks.
func Apply(engine *gameecs.Engine, target akara.Entity, stateID, sourceID string, duration int64) (akara.Entity, error) {
	return enqueue(engine, target, OperationApply, stateID, sourceID, duration)
}

// Remove queues removal of the state contributed by one exact source.
func Remove(engine *gameecs.Engine, target akara.Entity, stateID, sourceID string) (akara.Entity, error) {
	return enqueue(engine, target, OperationRemove, stateID, sourceID, 0)
}

func enqueue(engine *gameecs.Engine, target akara.Entity, operation, stateID, sourceID string, duration int64) (akara.Entity, error) {
	if engine == nil {
		return 0, fmt.Errorf("state: engine is required")
	}
	stateID, sourceID = strings.TrimSpace(stateID), strings.TrimSpace(sourceID)
	if stateID == "" || sourceID == "" {
		return 0, fmt.Errorf("state: state and source are required")
	}
	if operation == OperationApply && duration <= 0 {
		return 0, fmt.Errorf("state: apply duration must be positive")
	}
	requests, err := akara.RegisterSchema(engine.World(), requestSchema())
	if err != nil {
		return 0, err
	}
	entity := engine.World().MustCreateEntity()
	_, err = requests.Set(entity, map[string]any{
		"operation": operation, "target": target, "state_id": stateID,
		"source_id": sourceID, "duration": duration, "policy": RefreshSameSource,
	})
	return entity, err
}

func applyRequests(context gameecs.Context, commands *akara.CommandBuffer, world *akara.World, requests, instances, events *akara.DynamicStore) (map[akara.Entity]bool, error) {
	// Fold repeated requests by their authoritative identity. Requests are visited
	// in stable entity order, so the final request for a key deterministically wins.
	plans := make(map[requestKey]requestPlan)
	order := make([]requestKey, 0, requests.Len())
	for _, requestEntity := range requests.Entities() {
		request, _ := requests.Get(requestEntity)
		operation, _ := request.Get("operation")
		target, _ := request.Get("target")
		stateID, _ := request.Get("state_id")
		sourceID, _ := request.Get("source_id")
		duration, _ := request.Get("duration")
		policy, _ := request.Get("policy")
		stateName, sourceName := strings.TrimSpace(stateID.(string)), strings.TrimSpace(sourceID.(string))
		if stateName == "" || sourceName == "" || policy != RefreshSameSource {
			return nil, fmt.Errorf("state: request requires state, source, and supported policy")
		}
		key := requestKey{target: target.(akara.Entity), stateID: stateName, sourceID: sourceName}
		if _, exists := plans[key]; !exists {
			order = append(order, key)
		}
		plans[key] = requestPlan{operation: operation.(string), duration: duration.(int64), policy: policy.(string)}
		commands.Remove(requests, requestEntity)
	}

	touched := make(map[akara.Entity]bool)
	for _, key := range order {
		plan := plans[key]
		match, found := findInstance(instances, key.target, key.stateID, key.sourceID)
		switch plan.operation {
		case OperationApply:
			if plan.duration <= 0 {
				return nil, fmt.Errorf("state: apply duration must be positive")
			}
			expires := int64(context.Tick) + plan.duration
			if found {
				component, _ := instances.Get(match)
				if err := component.Set("expires_tick", expires); err != nil {
					return nil, err
				}
				touched[match] = true
				commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventRefreshed, context.Tick, key.target, key.stateID, key.sourceID, expires, "refresh")})
			} else {
				commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{instances: {"target": key.target, "state_id": key.stateID, "source_id": key.sourceID, "applied_tick": int64(context.Tick), "expires_tick": expires, "policy": plan.policy}})
				commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventApplied, context.Tick, key.target, key.stateID, key.sourceID, expires, "apply")})
			}
		case OperationRemove:
			if found {
				component, _ := instances.Get(match)
				expires, _ := component.Get("expires_tick")
				touched[match] = true
				commands.Remove(instances, match)
				commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventRemoved, context.Tick, key.target, key.stateID, key.sourceID, expires.(int64), "explicit")})
			}
		default:
			return nil, fmt.Errorf("state: unsupported operation %q", plan.operation)
		}
	}
	return touched, nil
}

func expireInstances(context gameecs.Context, commands *akara.CommandBuffer, world *akara.World, instances, events *akara.DynamicStore, touched map[akara.Entity]bool) {
	for _, entity := range instances.Entities() {
		if touched[entity] {
			continue
		}
		component, _ := instances.Get(entity)
		expires, _ := component.Get("expires_tick")
		if int64(context.Tick) < expires.(int64) {
			continue
		}
		target, _ := component.Get("target")
		stateID, _ := component.Get("state_id")
		sourceID, _ := component.Get("source_id")
		commands.Remove(instances, entity)
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventRemoved, context.Tick, target.(akara.Entity), stateID.(string), sourceID.(string), expires.(int64), "expired")})
	}
}

func findInstance(instances *akara.DynamicStore, target akara.Entity, stateID, sourceID string) (akara.Entity, bool) {
	for _, entity := range instances.Entities() {
		component, _ := instances.Get(entity)
		candidateTarget, _ := component.Get("target")
		candidateState, _ := component.Get("state_id")
		candidateSource, _ := component.Get("source_id")
		if candidateTarget == target && candidateState == stateID && candidateSource == sourceID {
			return entity, true
		}
	}
	return 0, false
}

func eventValues(kind string, tick uint64, target akara.Entity, stateID, sourceID string, expires int64, reason string) map[string]any {
	return map[string]any{"kind": kind, "tick": int64(tick), "target": target, "state_id": stateID, "source_id": sourceID, "expires_tick": expires, "reason": reason}
}
