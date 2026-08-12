package d2legacy

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

type fixtureRecords struct{}

func (fixtureRecords) Invalidate(string)  {}
func (fixtureRecords) Loaded(string) bool { return true }
func (fixtureRecords) Load(path string) ([]map[string]string, error) {
	if path == "data/global/excel/skills.txt" {
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	}
	return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
}

func TestAuthorityBootsWithoutClientOrRenderer(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(context.Background(), content.Shim(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
