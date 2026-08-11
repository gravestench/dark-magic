package state

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestTimedStateRefreshesSameSourceAndExpiresExclusively(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := Register(engine); err != nil {
		t.Fatal(err)
	}
	_, instances, events, _ := registerStores(engine)
	target := engine.World().MustCreateEntity()
	apply(t, engine, target, "stun", "skill:42", 3)
	step(t, engine)
	if instances.Len() != 1 || eventKinds(events)[0] != EventApplied {
		t.Fatalf("instances=%d events=%v", instances.Len(), eventKinds(events))
	}
	instance, _ := instances.Get(instances.Entities()[0])
	expires, _ := instance.Get("expires_tick")
	if expires != int64(4) {
		t.Fatalf("expires=%v", expires)
	}
	step(t, engine)
	apply(t, engine, target, "stun", "skill:42", 3)
	step(t, engine)
	if instances.Len() != 1 {
		t.Fatalf("refresh duplicated instance: %d", instances.Len())
	}
	expires, _ = instance.Get("expires_tick")
	if expires != int64(6) {
		t.Fatalf("refreshed expires=%v", expires)
	}
	if got := eventKinds(events); len(got) != 2 || got[1] != EventRefreshed {
		t.Fatalf("events=%v", got)
	}
	for engine.Tick() < 5 {
		step(t, engine)
	}
	if instances.Len() != 1 {
		t.Fatal("state expired before exclusive deadline")
	}
	step(t, engine)
	if instances.Len() != 0 {
		t.Fatal("state remained at expiration tick")
	}
	removed := events.Entities()[2]
	event, _ := events.Get(removed)
	reason, _ := event.Get("reason")
	if reason != "expired" {
		t.Fatalf("reason=%v", reason)
	}
}

func TestTimedStateKeepsIndependentSourcesAndExplicitRemoval(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := Register(engine); err != nil {
		t.Fatal(err)
	}
	_, instances, events, _ := registerStores(engine)
	target := engine.World().MustCreateEntity()
	apply(t, engine, target, "chill", "skill:a", 10)
	apply(t, engine, target, "chill", "skill:b", 10)
	step(t, engine)
	if instances.Len() != 2 {
		t.Fatalf("instances=%d", instances.Len())
	}
	if _, err := Remove(engine, target, "chill", "skill:a"); err != nil {
		t.Fatal(err)
	}
	step(t, engine)
	if instances.Len() != 1 {
		t.Fatalf("instances=%d", instances.Len())
	}
	remaining, _ := instances.Get(instances.Entities()[0])
	source, _ := remaining.Get("source_id")
	if source != "skill:b" {
		t.Fatalf("source=%v", source)
	}
	last, _ := events.Get(events.Entities()[2])
	reason, _ := last.Get("reason")
	if reason != "explicit" {
		t.Fatalf("reason=%v", reason)
	}
}

func TestTimedStateSnapshotRestoreReplaysExpiration(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := Register(engine); err != nil {
		t.Fatal(err)
	}
	target := engine.World().MustCreateEntity()
	apply(t, engine, target, "stun", "skill:42", 2)
	step(t, engine)
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := gameecs.RestoreSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if err := Register(restored); err != nil {
		t.Fatal(err)
	}
	step(t, engine)
	step(t, restored)
	step(t, engine)
	step(t, restored)
	left, _ := engine.Snapshot()
	right, _ := restored.Snapshot()
	leftSum, _ := left.Checksum()
	rightSum, _ := right.Checksum()
	if leftSum != rightSum {
		t.Fatalf("restored expiration differs: %s != %s", leftSum, rightSum)
	}
}

func TestTimedStateLastSameTickRequestWins(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := Register(engine); err != nil {
		t.Fatal(err)
	}
	_, instances, events, _ := registerStores(engine)
	target := engine.World().MustCreateEntity()
	apply(t, engine, target, "stun", "skill:42", 2)
	apply(t, engine, target, "stun", "skill:42", 7)
	step(t, engine)
	if instances.Len() != 1 || events.Len() != 1 {
		t.Fatalf("instances=%d events=%d", instances.Len(), events.Len())
	}
	instance, _ := instances.Get(instances.Entities()[0])
	expires, _ := instance.Get("expires_tick")
	if expires != int64(8) {
		t.Fatalf("expires=%v", expires)
	}
}

func apply(t *testing.T, engine *gameecs.Engine, target akara.Entity, stateID, sourceID string, duration int64) {
	t.Helper()
	if _, err := Apply(engine, target, stateID, sourceID, duration); err != nil {
		t.Fatal(err)
	}
}

func step(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
}

func eventKinds(events *akara.DynamicStore) []string {
	result := make([]string, 0, events.Len())
	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)
		kind, _ := event.Get("kind")
		result = append(result, kind.(string))
	}
	return result
}
