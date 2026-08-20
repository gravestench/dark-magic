package sessionquic

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestTransformFrameRoundTripsWithinConservativeDatagramBudget verifies truncation and compact owner/entity fields.
func TestTransformFrameRoundTripsWithinConservativeDatagramBudget(t *testing.T) {
	entities := make([]playeradapter.WorldEntity, 100)
	for index := range entities {
		entities[index] = playeradapter.WorldEntity{
			ID: "entity-" + string(rune(index+33)),
			Position: playeradapter.HUDPosition{
				X: float64(index) + .25,
				Y: float64(index) + .5,
			},
			Mode:      "RN",
			Direction: int64(index % 16),
		}
	}

	view := emptyTransformView(42)
	view.HUD = playeradapter.HUD{
		Version:   playeradapter.HUDVersion,
		Tick:      42,
		Position:  playeradapter.HUDPosition{X: 12.5, Y: 20.25},
		Movement:  playeradapter.HUDMovement{Velocity: playeradapter.HUDPosition{X: 10, Y: -5}},
		Animation: playeradapter.HUDAnimation{Mode: "WL", Direction: 7, StartTick: 40},
	}
	view.World.Entities = entities

	credential := gameserver.SessionCredential("credential")

	encoded, err := encodeTransformFrame(credential, transformSnapshot(t, view))
	if err != nil {
		t.Fatal(err)
	}

	if len(encoded) > MaxDatagramPayloadBytes {
		t.Fatalf("transform frame bytes = %d, budget = %d", len(encoded), MaxDatagramPayloadBytes)
	}

	decoded, err := decodeTransformFrame(credential, encoded)
	if err != nil {
		t.Fatal(err)
	}

	wantCount := (MaxDatagramPayloadBytes - transformHeaderBytes) / transformEntityBytes
	if decoded.Tick != 42 || !decoded.Truncated || len(decoded.Entities) != wantCount {
		t.Fatalf("decoded transform = %#v", decoded)
	}

	if decoded.OwnerX != 12.5 || decoded.OwnerY != 20.25 || decoded.VelocityX != 10 || decoded.VelocityY != -5 {
		t.Fatalf("decoded owner transform = %#v", decoded)
	}

	if decoded.OwnerMode != [2]byte{'W', 'L'} || decoded.OwnerDirection != 7 || decoded.OwnerAnimationStartTick != 40 {
		t.Fatalf("decoded owner animation = %#v", decoded)
	}

	if decoded.Entities[0].IDHash != PublicIDHash(entities[0].ID) ||
		decoded.Entities[0].Mode != [2]byte{'R', 'N'} {
		t.Fatalf("decoded first entity = %#v", decoded.Entities[0])
	}
}

// TestTransformFrameRejectsWrongCredentialAndMalformedLength protects membership isolation and exact framing.
func TestTransformFrameRejectsWrongCredentialAndMalformedLength(t *testing.T) {
	encoded, err := encodeTransformFrame("one", transformSnapshot(t, emptyTransformView(1)))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := decodeTransformFrame("two", encoded); err != ErrWire {
		t.Fatalf("wrong credential error = %v", err)
	}

	if _, err := decodeTransformFrame("one", append(encoded, 0)); err != ErrWire {
		t.Fatalf("malformed length error = %v", err)
	}
}

// TestTransformFrameCarriesReliableMissileIdentityAndPosition verifies missile samples omit actor-only state.
func TestTransformFrameCarriesReliableMissileIdentityAndPosition(t *testing.T) {
	missile := playeradapter.WorldMissile{
		ID:       "missile:42",
		Position: playeradapter.HUDPosition{X: 12.25, Y: -3.5},
	}
	view := emptyTransformView(7)
	view.World.Missiles = []playeradapter.WorldMissile{missile}

	encoded, err := encodeTransformFrame("credential", transformSnapshot(t, view))
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := decodeTransformFrame("credential", encoded)
	if err != nil {
		t.Fatal(err)
	}

	if len(decoded.Entities) != 1 || decoded.Entities[0].IDHash != PublicIDHash(missile.ID) ||
		decoded.Entities[0].X != 12.25 || decoded.Entities[0].Y != -3.5 ||
		decoded.Entities[0].Mode != [2]byte{} || decoded.Entities[0].AnimationStartTick != 0 {
		t.Fatalf("decoded missile transform = %+v", decoded.Entities)
	}
}

// TestTransformAnimationAgePreservesUnknownStartAcrossLongSessions protects the reserved unknown-age sentinel.
func TestTransformAnimationAgePreservesUnknownStartAcrossLongSessions(t *testing.T) {
	tick := uint64(1) << 33
	if got := startTick(tick, tickAge(tick, 0)); got != 0 {
		t.Fatalf("unknown animation start decoded as %d", got)
	}

	if got := startTick(tick, tickAge(tick, tick-12)); got != tick-12 {
		t.Fatalf("recent animation start decoded as %d, want %d", got, tick-12)
	}
}

// FuzzDecodeTransformFrame proves arbitrary datagrams cannot escape size, tick, or entity-count bounds.
func FuzzDecodeTransformFrame(f *testing.F) {
	credential := gameserver.SessionCredential("credential")
	view := emptyTransformView(4)
	view.HUD.Position = playeradapter.HUDPosition{X: 1, Y: 2}

	valid, err := encodeTransformFrame(credential, transformSnapshot(f, view))
	if err != nil {
		f.Fatal(err)
	}

	f.Add(valid)
	f.Add([]byte{0x44, 0x4d, TransformFrameVersion})
	f.Add(make([]byte, MaxDatagramPayloadBytes+1))
	f.Fuzz(func(t *testing.T, payload []byte) {
		frame, err := decodeTransformFrame(credential, payload)
		if err != nil {
			return
		}

		maximum := (MaxDatagramPayloadBytes - transformHeaderBytes) / transformEntityBytes
		if frame.Tick == 0 || len(frame.Entities) > maximum || len(payload) > MaxDatagramPayloadBytes {
			t.Fatalf("decoder accepted out-of-bounds frame: %#v", frame)
		}
	})
}

// BenchmarkTransformFrameEncodeDecode tracks the hot codec path under the maximum reliable entity projection.
func BenchmarkTransformFrameEncodeDecode(b *testing.B) {
	credential, snapshot := transformBenchmarkInput(b)
	b.ReportAllocs()

	for b.Loop() {
		encoded, err := encodeTransformFrame(credential, snapshot)
		if err != nil {
			b.Fatal(err)
		}

		if _, err := decodeTransformFrame(credential, encoded); err != nil {
			b.Fatal(err)
		}
	}
}

// TestTransformCodecAllocationBudget prevents helper extraction from adding per-frame allocation growth.
func TestTransformCodecAllocationBudget(t *testing.T) {
	credential, snapshot := transformBenchmarkInput(t)

	allocations := testing.AllocsPerRun(25, func() {
		encoded, err := encodeTransformFrame(credential, snapshot)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := decodeTransformFrame(credential, encoded); err != nil {
			t.Fatal(err)
		}
	})
	if allocations > 320 {
		t.Fatalf("transform codec allocations = %.0f, budget = 320", allocations)
	}
}

// fataler captures the cleanup-aware failure surface shared by tests, fuzz seeds, and benchmarks.
type fataler interface {
	Helper()
	Fatal(...any)
}

// emptyTransformView builds version-consistent semantic sections so individual tests vary only relevant fields.
func emptyTransformView(tick uint64) playeradapter.ClientView {
	return playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion,
		Tick:    tick,
		HUD:     playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: tick},
		World:   playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: tick},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: tick},
		Events: playeradapter.EventView{
			Version: playeradapter.EventViewVersion,
			Tick:    tick,
			Events:  []playeradapter.SemanticEvent{},
		},
	}
}

// transformSnapshot marshals the reliable semantic view and keeps its envelope tick aligned.
func transformSnapshot(test fataler, view playeradapter.ClientView) gameserver.Snapshot {
	test.Helper()

	payload, err := json.Marshal(view)
	if err != nil {
		test.Fatal(err)
	}

	return gameserver.Snapshot{Tick: view.Tick, Payload: payload}
}

// transformBenchmarkInput fills the reliable entity ceiling to exercise truncation and compact record loops.
func transformBenchmarkInput(test fataler) (gameserver.SessionCredential, gameserver.Snapshot) {
	test.Helper()

	entities := make([]playeradapter.WorldEntity, playeradapter.MaxWorldViewEntities)
	for index := range entities {
		entities[index] = playeradapter.WorldEntity{
			ID: fmt.Sprintf("entity-%d", index),
			Position: playeradapter.HUDPosition{
				X: float64(index),
				Y: float64(index),
			},
		}
	}

	view := emptyTransformView(100)
	view.World.Entities = entities

	return gameserver.SessionCredential("benchmark-credential"), transformSnapshot(test, view)
}
