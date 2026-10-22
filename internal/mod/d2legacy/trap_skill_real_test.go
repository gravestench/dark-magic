package d2legacy

import (
	"os"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

// TestOwnedTargetAssassinTrapFamilyRecords pins the ten exact Expansion 1.14d
// rows admitted by the generic six-shape decoder. In particular, Inferno
// Sentry's expression assigns one rate to two hard-level sources and a second
// rate to Wake of Fire; flattening it into one shared percentage is incorrect.
func TestOwnedTargetAssassinTrapFamilyRecords(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = assets.Close() }()

	store := recordstore.New(assets)
	store.SetLogger(nil)

	skills, err := store.Load("data/global/excel/skills.txt")
	if err != nil {
		t.Fatal(err)
	}

	wants := map[string]map[string]string{
		"251": {"skill": "Fire Trauma", "lob": "1", "srvmissile": "bomb in air", "Param1": "5", "Param8": "9"},
		"256": {"skill": "Shock Field", "srvdofunc": "43", "srvmissilea": "shock field in air", "Param1": "6", "Param2": "4"},
		"257": {
			"skill": "Blade Sentinel", "srvdofunc": "44", "summon": "bladecreeper",
			"pettype": "assassintrap", "petmax": "5",
		},
		"261": {
			"skill": "Charged Bolt Sentry", "srvdofunc": "45", "summon": "chargeboltsentry",
			"sumskill1": "BoltSentry", "Param1": "5",
		},
		"262": {
			"skill": "Wake of Fire Sentry", "srvdofunc": "45", "summon": "wakeofdestruction",
			"sumskill1": "Wake Of Destruction Sentry", "Param1": "5",
		},
		"266": {
			"skill": "Blade Fury", "srvstfunc": "26", "srvdofunc": "48", "repeat": "1",
			"usemanaondo": "1", "startmana": "6", "Param3": "3", "Param4": "5",
			"mana": "8", "manashift": "5",
			"srvmissilea": "bladefragment1", "cltmissilea": "bladefragment2", "prgcalc1": "par4",
		},
		"271": {
			"skill": "Lightning Sentry", "srvdofunc": "45", "summon": "lightningsentry",
			"sumskill1": "sentry lightning", "Param1": "10",
		},
		"272": {
			"skill": "Inferno Sentry", "srvdofunc": "45", "summon": "infernosentry",
			"sumskill1": "mon inferno sentry", "Param7": "10", "Param8": "7",
		},
		"276": {
			"skill": "Death Sentry", "srvdofunc": "45", "summon": "deathsentry",
			"sumskill1": "mon death sentry", "sumskill2": "death sentry ltng", "Param1": "5",
		},
		"277": {
			"skill": "Blade Shield", "srvstfunc": "28", "srvdofunc": "54", "periodic": "1",
			"aurastate": "bladeshield", "Param3": "25", "Param4": "6",
		},
	}
	for id, fields := range wants {
		row := rowBy(skills, "Id", id)
		if row == nil {
			t.Fatalf("owned Expansion 1.14d Assassin trap skill %s is missing", id)
		}

		for field, want := range fields {
			if row[field] != want {
				t.Fatalf("owned Expansion 1.14d skill %s %s = %q, want %q", id, field, row[field], want)
			}
		}
	}

	boltSentry := rowBy(skills, "skill", "BoltSentry")
	if boltSentry == nil || boltSentry["srvdofunc"] != "17" || boltSentry["srvmissilea"] != "sentrychargedbolt" {
		t.Fatalf("owned Expansion 1.14d BoltSentry initializer fields = %#v", boltSentry)
	}

	missiles, err := store.Load("data/global/excel/Missiles.txt")
	if err != nil {
		t.Fatal(err)
	}
	for name, fields := range map[string]map[string]string{
		"blade creeper":       {"Id": "392", "NextHit": "1", "NextDelay": "25"},
		"sentrylightningbolt": {"Id": "526", "Vel": "30", "MaxVel": "30", "Range": "10", "pSrvDoFunc": "1"},
		"sentrychargedbolt":   {"Id": "495", "Vel": "12", "MaxVel": "12", "Range": "98", "pSrvDoFunc": "1"},
		"bladefragment1":      {"Id": "506", "Vel": "22", "MaxVel": "22", "Range": "40", "pSrvDoFunc": "1"},
		"bladefragment2":      {"Id": "508", "Vel": "22", "MaxVel": "22", "Range": "33", "pCltDoFunc": "1"},
	} {
		row := rowBy(missiles, "Missile", name)
		if row == nil {
			t.Fatalf("owned Expansion 1.14d sentry missile %s is missing", name)
		}
		for field, want := range fields {
			if row[field] != want {
				t.Fatalf("owned Expansion 1.14d missile %s %s = %q, want %q", name, field, row[field], want)
			}
		}
	}

	inferno := rowBy(skills, "Id", "272")
	formula := inferno["EDmgSymPerCalc"]

	terms := []string{
		"skill('Fire Trauma'.blvl)",
		"skill('Death Sentry'.blvl)",
		"* par7",
		"skill('Wake of Fire Sentry'.blvl)*par8",
	}
	for _, term := range terms {
		if !strings.Contains(formula, term) {
			t.Fatalf("owned Expansion 1.14d Inferno Sentry synergy formula %q is missing %q", formula, term)
		}
	}

	engine := gameecs.New()
	defer func() { _ = engine.Close() }()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), store, engine, session, Config{Seed: 408})
	if err != nil {
		t.Fatalf("compose owned Expansion 1.14d Assassin trap definitions: %v", err)
	}

	if err := authority.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
}
