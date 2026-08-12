// Package d2legacy composes the canonical first-party gameplay mod on top of
// generic engine services. It contains host wiring, never Diablo game rules;
// those live in the bundled Lua modules under lua/d2legacy.
package d2legacy

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Identity pins the exact authoritative Lua bytes used by one session. Sorted
// file order makes the result independent of filesystem traversal order.
func Identity(source fs.FS) (simulation.RuntimeIdentity, error) {
	var names []string
	err := fs.WalkDir(source, "lua/d2legacy", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(name, ".lua") {
			names = append(names, name)
		}
		return nil
	})
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return simulation.RuntimeIdentity{}, err
		}
		_, _ = hash.Write([]byte(name))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(data)
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	return simulation.RuntimeIdentity{
		ModID: "d2legacy", ContractVersion: "v1",
		PackageHash: digest, AuthoritativeHash: digest,
		ConfigurationHash: "d2legacy/default/v1",
		CapabilityVersions: map[string]string{
			"engine.authority_command": "v1", "engine.authority_random": "v1",
			"engine.authority_state": "v1", "engine.ecs": "v1", "engine.records": "v1",
		},
	}, nil
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
	} {
		if err := streams.Register(name); err != nil {
			return nil, err
		}
	}
	return streams, nil
}
