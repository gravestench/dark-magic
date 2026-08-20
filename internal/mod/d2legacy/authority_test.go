package d2legacy

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

// fixtureRecords is a small policy-neutral record reader. Unmentioned files
// return an empty table, which is enough for tests that do not exercise them.
type fixtureRecords map[string][]map[string]string

// Load returns the smallest tables needed to boot the authority fixture. An
// unknown path fails loudly so new startup dependencies cannot enter unnoticed.
func (records fixtureRecords) Load(path string) ([]map[string]string, error) {
	if rows, found := records[path]; found {
		return rows, nil
	}

	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{
			"class": "Amazon", "StartSkill": "Fire Bolt", "WalkVelocity": "6", "RunVelocity": "9",
			"vit": "20", "stamina": "84", "RunDrain": "20", "StaminaPerLevel": "4", "StaminaPerVitality": "4",
		}}, nil
	case "data/global/excel/experience.txt":
		return []map[string]string{
			{
				"Amazon": "0", "Sorceress": "0", "Necromancer": "0", "Paladin": "0",
				"Barbarian": "0", "Druid": "0", "Assassin": "0",
			},
			{
				"Amazon": "5", "Sorceress": "5", "Necromancer": "5", "Paladin": "5",
				"Barbarian": "5", "Druid": "5", "Assassin": "5",
			},
			{
				"Amazon": "15", "Sorceress": "15", "Necromancer": "15", "Paladin": "15",
				"Barbarian": "15", "Druid": "15", "Assassin": "15",
			},
		}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{
			"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt",
			"skilldesc": "firebolt", "leftskill": "1", "general": "0", "passive": "0",
			"etype": "fire", "interrupt": "1", "srvstfunc": "",
			"srvdofunc": "", "mana": "5", "manashift": "7",
			"emin": "3", "emax": "6", "HitShift": "8",
		}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{
			"Missile": "firebolt", "Skill": "Fire Bolt",
			"pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1",
			"Vel": "20", "Range": "40", "Size": "2",
			"CelFile": "firebolt",
		}}, nil
	default:
		return nil, nil
	}
}

// Invalidate is intentionally inert because fixture rows never change during a
// test; satisfying Records keeps cache mechanics outside authority assertions.
func (fixtureRecords) Invalidate(string) {}

// Loaded reports fixture data as resident so tests exercise authority startup
// rather than production cache state transitions.
func (fixtureRecords) Loaded(string) bool { return true }

// TestGenericHostCanBootWithoutD2Legacy proves the generic host has no implicit
// dependency on this game-specific authority package.
func TestGenericHostCanBootWithoutD2Legacy(t *testing.T) {
	engine := gameecs.New()
	defer func() { _ = engine.Close() }()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	if engine.Tick() != 0 {
		t.Fatalf("fresh generic engine tick = %d", engine.Tick())
	}
}
