package gameserver

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type fixtureRecords struct{}

func (fixtureRecords) Invalidate(string)  {}
func (fixtureRecords) Loaded(string) bool { return true }
func (fixtureRecords) Load(path string) ([]map[string]string, error) {
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{"class": "Amazon", "StartSkill": "Fire Bolt"}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "skilldesc": "firebolt", "leftskill": "1", "general": "0", "passive": "0", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
	default:
		return nil, nil
	}
}

func TestHeadlessHostPinsRunningAuthorityToAdmissionAndReconnect(t *testing.T) {
	host, err := Start(t.Context(), content.D2Legacy(), fixtureRecords{}, Config{
		SessionID: "game-7", Seed: 42, Prediction: gamesession.PredictionLimited,
		InitialData: map[string]any{"difficulty": "normal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close(context.Background())

	if !reflect.DeepEqual(host.Allocation.Identity, host.Authority.Identity) {
		t.Fatal("allocation identity differs from running d2legacy authority")
	}
	replay, err := host.Session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := simulation.RuntimeIdentityFromParticipants(replay.InitialParticipants)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pinned, host.Allocation.Identity) {
		t.Fatal("session participant identity differs from allocation")
	}

	token, err := host.Admit("character:alice", host.Authority.Identity)
	if err != nil {
		t.Fatal(err)
	}
	if token.SessionID != "game-7" || token.Prediction != gamesession.PredictionLimited {
		t.Fatalf("admission token = %#v", token)
	}
	if err := host.ValidateReconnect(token, host.Authority.Identity); err != nil {
		t.Fatal(err)
	}
	canonical, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	predictedEngine := gameecs.New()
	defer predictedEngine.Close()
	predicted, err := predictedEngine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	predicted.Tick = canonical.Tick + 100
	correction, err := gamesession.ReconcilePrediction(token.Prediction, &predicted, canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !correction.Corrected {
		t.Fatal("divergent d2legacy client prediction was accepted as canonical")
	}
	if err := correction.Apply(predictedEngine); err != nil {
		t.Fatal(err)
	}
	corrected, err := predictedEngine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if difference := gameecs.FirstDifference(*canonical.Snapshot, corrected); difference != "" {
		t.Fatalf("corrected d2legacy client differs: %s", difference)
	}

	mismatch := host.Authority.Identity
	mismatch.PackageHash = "different-package"
	if _, err := host.Admit("character:bob", mismatch); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("mismatched admission error = %v", err)
	}
	if err := host.ValidateReconnect(token, mismatch); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("mismatched reconnect error = %v", err)
	}
	stale := token
	stale.SessionID = "old-game"
	if err := host.ValidateReconnect(stale, host.Authority.Identity); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("stale reconnect error = %v", err)
	}
}

func TestHeadlessHostRejectsUnknownPredictionContract(t *testing.T) {
	host, err := Start(t.Context(), content.D2Legacy(), fixtureRecords{}, Config{
		SessionID: "game-invalid", Prediction: gamesession.PredictionTier("client_decides"),
	})
	if host != nil {
		_ = host.Close(context.Background())
		t.Fatal("invalid prediction returned a host")
	}
	if !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("invalid prediction error = %v", err)
	}
}

func TestSharedHostRunsStandaloneListenAndRealmModes(t *testing.T) {
	for _, mode := range []Mode{ModeStandalone, ModeListen, ModeRealm} {
		t.Run(string(mode), func(t *testing.T) {
			host, err := Start(t.Context(), content.D2Legacy(), fixtureRecords{}, Config{
				Mode: mode, SessionID: "mode-" + string(mode), Seed: 42,
				Prediction: gamesession.PredictionLimited,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer host.Close(context.Background())
			if host.Mode != mode || host.Session == nil || host.Authority == nil {
				t.Fatalf("host composition = mode %q session %v authority %v",
					host.Mode, host.Session != nil, host.Authority != nil)
			}
			if err := host.Session.Step(); err != nil {
				t.Fatal(err)
			}
			replay, err := host.Session.Replay()
			if err != nil {
				t.Fatal(err)
			}
			identity, err := simulation.RuntimeIdentityFromParticipants(replay.InitialParticipants)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(identity, host.Allocation.Identity) {
				t.Fatal("hosting mode changed the authoritative runtime identity")
			}
		})
	}
}
