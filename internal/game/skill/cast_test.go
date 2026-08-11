package skill

import (
	"testing"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestCastLifecycleCommitsManaAndEmitsHeadlessPhases(t *testing.T) {
	registry, err := NewRegistry(Definition{SkillID: 42, Behavior: BehaviorPointEvent, TargetPolicy: TargetPoint, ManaCost: 3, EffectDelay: 1, CompleteDelay: 2, Interruptible: true})
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterCastLifecycle(engine, registry); err != nil {
		t.Fatal(err)
	}
	requests, states, events, vitals, _, _ := registerCastStores(engine)
	owner := engine.World().MustCreateEntity()
	mustSet(t, vitals, owner, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(8), "max_mana": int64(8)})
	mustSet(t, requests, owner, castRequestFixture())
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	vital, _ := vitals.Get(owner)
	mana, _ := vital.Get("mana")
	if mana != int64(5) || !states.Has(owner) || requests.Has(owner) {
		t.Fatalf("start mana=%v state=%v request=%v", mana, states.Has(owner), requests.Has(owner))
	}
	state, _ := states.Get(owner)
	phase, _ := state.Get("phase")
	if phase != "started" {
		t.Fatalf("phase=%v", phase)
	}
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	phase, _ = state.Get("phase")
	if phase != "effect" {
		t.Fatalf("phase=%v", phase)
	}
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	if states.Has(owner) {
		t.Fatal("completed cast remains")
	}
	want := []string{EventCastStarted, EventSkillEffect, EventCastCompleted}
	if events.Len() != len(want) {
		t.Fatalf("events=%d", events.Len())
	}
	for index, entity := range events.Entities() {
		event, _ := events.Get(entity)
		kind, _ := event.Get("kind")
		if kind != want[index] {
			t.Fatalf("event %d=%v", index, kind)
		}
	}
}

func TestCastLifecycleInterruptsWithoutEffect(t *testing.T) {
	registry, _ := NewRegistry(Definition{SkillID: 42, Behavior: BehaviorPointEvent, TargetPolicy: TargetPoint, EffectDelay: 2, CompleteDelay: 3, Interruptible: true})
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterCastLifecycle(engine, registry); err != nil {
		t.Fatal(err)
	}
	requests, states, events, vitals, _, _ := registerCastStores(engine)
	owner := engine.World().MustCreateEntity()
	mustSet(t, vitals, owner, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(0), "max_mana": int64(0)})
	mustSet(t, requests, owner, castRequestFixture())
	_ = engine.Update(time.Second / 25)
	state, _ := states.Get(owner)
	_ = state.Set("interruption_requested", true)
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	if states.Has(owner) || events.Len() != 2 {
		t.Fatalf("state=%v events=%d", states.Has(owner), events.Len())
	}
	last, _ := events.Get(events.Entities()[1])
	kind, _ := last.Get("kind")
	if kind != EventCastInterrupted {
		t.Fatalf("last=%v", kind)
	}
}

func TestCastLifecycleRejectsInsufficientManaWithoutCrashing(t *testing.T) {
	registry, _ := NewRegistry(Definition{SkillID: 42, Behavior: BehaviorPointEvent, TargetPolicy: TargetPoint, ManaCost: 4, EffectDelay: 1, CompleteDelay: 1})
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterCastLifecycle(engine, registry); err != nil {
		t.Fatal(err)
	}
	requests, states, events, vitals, _, _ := registerCastStores(engine)
	owner := engine.World().MustCreateEntity()
	mustSet(t, vitals, owner, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(3), "max_mana": int64(3)})
	mustSet(t, requests, owner, castRequestFixture())
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	if states.Has(owner) || requests.Has(owner) || events.Len() != 1 {
		t.Fatalf("state=%v request=%v events=%d", states.Has(owner), requests.Has(owner), events.Len())
	}
	event, _ := events.Get(events.Entities()[0])
	kind, _ := event.Get("kind")
	reason, _ := event.Get("reason")
	if kind != EventCastRejected || reason != "insufficient mana" {
		t.Fatalf("event=%v reason=%v", kind, reason)
	}
}

func castRequestFixture() map[string]any {
	return map[string]any{"player": "alpha", "side": "right", "skill_id": int64(42), "skill_level": int64(3), "target_x": 12.0, "target_y": 9.0, "target_id": "", "request_tick": int64(1)}
}
