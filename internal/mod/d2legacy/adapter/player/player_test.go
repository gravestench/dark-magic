package player

import (
	"encoding/json"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestAdmissionCommandCarriesDurableFactsWithoutInterpretingD2Policy(t *testing.T) {
	character := d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 3, Expansion: true,
		Stats: &d2save.Stats{Dexterity: 20, Defense: 7, Experience: 100, Health: 25, MaxHealth: 30, Mana: 12, MaxMana: 15}}
	destination, err := NewDestination(5, 7, 100, 80, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	command, err := AdmissionCommand(character, "player-1", destination, "server", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	var entry Entry
	if err := json.Unmarshal(command.Payload, &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Dexterity != 20 || entry.Defense != 7 || entry.Class != "Amazon" {
		t.Fatalf("durable facts = %#v", entry)
	}
	encoded := string(command.Payload)
	for _, forbidden := range []string{"attack_rating", "physical_min_raw", "melee_range", "weapon_class", "\"token\""} {
		if contains(encoded, forbidden) {
			t.Fatalf("Go admission interpreted D2 policy into %q", forbidden)
		}
	}
}

func TestEntrySourceEmitsTrustedCommandForSelectedCharacter(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	saves := d2save.New(d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	if err := saves.Select("hero"); err != nil {
		t.Fatal(err)
	}
	source, err := NewEntrySourceAtLocation(engine, saves, "player", 12, 13, 100, 80, 5, 109)
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(1)
	if len(commands) != 1 || commands[0].Kind != EnterCommand || commands[0].Authority != simulation.AuthoritySystem {
		t.Fatalf("commands = %#v", commands)
	}
}

func TestRemoteAdmissionRejectsPlayerAuthority(t *testing.T) {
	destination, _ := NewDestination(23, 17, 100, 80, 1, 1)
	character := d2save.Character{ID: "realm-amazon", Name: "RemoteHero", Class: "Amazon", Level: 1}
	if _, err := AdmissionCommand(character, "account:42", destination, "client", 1, 1, simulation.AuthorityPlayer); err == nil {
		t.Fatal("client minted trusted admission")
	}
}

func contains(value, part string) bool {
	for index := 0; index+len(part) <= len(value); index++ {
		if value[index:index+len(part)] == part {
			return true
		}
	}
	return false
}
