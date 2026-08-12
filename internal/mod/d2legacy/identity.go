// Package d2legacy composes the canonical first-party gameplay mod on top of
// generic engine services. It contains host wiring, never Diablo game rules;
// those live in the bundled Lua modules under lua/d2legacy.
package d2legacy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Identity pins the exact authoritative Lua bytes used by one session. Sorted
// file order makes the result independent of filesystem traversal order.
func Identity(source fs.FS, configuration ...map[string]any) (simulation.RuntimeIdentity, error) {
	packageDigest, err := hashSource(source, ".", func(string) bool { return true })
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}
	authoritativeDigest, err := hashSource(source, "lua/d2legacy", func(name string) bool {
		return strings.HasSuffix(name, ".lua")
	})
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	configured := map[string]any{}
	if len(configuration) > 0 && configuration[0] != nil {
		configured = configuration[0]
	}
	encodedConfiguration, err := json.Marshal(configured)
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}
	configurationDigest := sha256.Sum256(encodedConfiguration)
	dependency := func(value string) string { sum := sha256.Sum256([]byte(value)); return hex.EncodeToString(sum[:]) }
	return simulation.RuntimeIdentity{
		ModID: "d2legacy", ContractVersion: "v1", PackageHash: packageDigest,
		AuthoritativeHash: authoritativeDigest,
		Dependencies: map[string]string{"engine/deterministic": "sha256:" + dependency("v1"),
			"engine/worldgen": "sha256:" + dependency("v1")},
		ConfigurationHash: hex.EncodeToString(configurationDigest[:]),
		CapabilityVersions: map[string]string{
			"engine.authority_command": "v1", "engine.authority_random": "v1",
			"engine.authority_state": "v1", "engine.ecs": "v1", "engine.records": "v1",
		},
	}, nil
}

func hashSource(source fs.FS, root string, include func(string) bool) (string, error) {
	var names []string
	err := fs.WalkDir(source, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && include(name) {
			names = append(names, name)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(name))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// NewRandomStreams declares every purpose currently consumed by authoritative
// Lua. Names are part of the deterministic contract and checkpoint identity.
func NewRandomStreams(seed uint64) (*simulation.RandomStreams, error) {
	streams := simulation.NewRandomStreams(seed)
	for _, name := range []string{
		"d2legacy.combat.fire_bolt.damage",
		"d2legacy.combat.basic_melee.hit",
		"d2legacy.combat.basic_melee.damage",
		"d2legacy.monster.spawn.life",
		"d2legacy.population.density", "d2legacy.population.family",
		"d2legacy.population.group", "d2legacy.population.seed",
		"d2legacy.loot.treasure_class", "d2legacy.loot.quality",
		"d2legacy.loot.affix_slot", "d2legacy.loot.affix_choice", "d2legacy.loot.affix_value",
	} {
		if err := streams.Register(name); err != nil {
			return nil, err
		}
	}
	return streams, nil
}
