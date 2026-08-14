package sessionquic

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestTransformFrameRoundTripsWithinConservativeDatagramBudget(t *testing.T) {
	entities := make([]playeradapter.WorldEntity, 100)
	for index := range entities {
		entities[index] = playeradapter.WorldEntity{
			ID: "entity-" + string(rune(index+33)), Position: playeradapter.HUDPosition{X: float64(index) + .25, Y: float64(index) + .5},
			Mode: "RN", Direction: int64(index % 16),
		}
	}
	view := playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion, Tick: 42,
		HUD: playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 42, Position: playeradapter.HUDPosition{X: 12.5, Y: 20.25},
			Movement:  playeradapter.HUDMovement{Velocity: playeradapter.HUDPosition{X: 10, Y: -5}},
			Animation: playeradapter.HUDAnimation{Mode: "WL", Direction: 7, StartTick: 40}},
		World:   playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 42, Entities: entities},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 42},
	}
	payload, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	credential := gameserver.SessionCredential("credential")
	encoded, err := encodeTransformFrame(credential, gameserver.Snapshot{Tick: 42, Payload: payload})
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
	if decoded.Entities[0].IDHash != PublicIDHash(entities[0].ID) || decoded.Entities[0].Mode != [2]byte{'R', 'N'} {
		t.Fatalf("decoded first entity = %#v", decoded.Entities[0])
	}
}

func TestTransformFrameRejectsWrongCredentialAndMalformedLength(t *testing.T) {
	view := playeradapter.ClientView{
		Version: playeradapter.ClientViewVersion, Tick: 1,
		HUD:     playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 1},
		World:   playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 1},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 1},
	}
	payload, _ := json.Marshal(view)
	encoded, err := encodeTransformFrame("one", gameserver.Snapshot{Tick: 1, Payload: payload})
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

func TestTransformAnimationAgePreservesUnknownStartAcrossLongSessions(t *testing.T) {
	tick := uint64(1) << 33
	if got := startTick(tick, tickAge(tick, 0)); got != 0 {
		t.Fatalf("unknown animation start decoded as %d", got)
	}
	if got := startTick(tick, tickAge(tick, tick-12)); got != tick-12 {
		t.Fatalf("recent animation start decoded as %d, want %d", got, tick-12)
	}
}
