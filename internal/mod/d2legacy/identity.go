// Package d2legacy composes the canonical first-party gameplay mod on top of
// generic engine services. It contains host wiring, never Diablo game rules;
// those live in the bundled Lua modules under lua/d2legacy.
package d2legacy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// Identity pins the canonical authoritative Lua source used by one session.
// Sorted file order makes the result independent of filesystem traversal order.
func Identity(source fs.FS, configuration ...map[string]any) (simulation.RuntimeIdentity, error) {
	builtin, err := modcache.DescribeBuiltin(source)
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	packages := simulation.RuntimePackageSet{Base: runtimePackageForBuiltin(builtin)}

	return IdentityForPackages(source, packages, simulation.EmptyAssetSetID, configuration...)
}

// IdentityForPackages builds the one canonical recipe used by every
// production host and client. It also proves that the supplied package set
// names the exact built-in d2legacy bytes used to construct the runtime.
func IdentityForPackages(
	source fs.FS,
	packages simulation.RuntimePackageSet,
	assetSetID string,
	configuration ...map[string]any,
) (simulation.RuntimeIdentity, error) {
	gameDataGenerationID := simulation.GameDataGenerationIDForAssetSet(assetSetID)

	return IdentityForPackagesAndData(
		source,
		packages,
		assetSetID,
		gameDataGenerationID,
		configuration...,
	)
}

// IdentityForPackagesAndData pins the exact immutable authoritative table
// generation used by a live session independently from presentation assets.
func IdentityForPackagesAndData(
	source fs.FS,
	packages simulation.RuntimePackageSet,
	assetSetID string,
	gameDataGenerationID string,
	configuration ...map[string]any,
) (simulation.RuntimeIdentity, error) {
	builtin, err := modcache.DescribeBuiltin(source)
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	expectedBase := runtimePackageForBuiltin(builtin)
	if packages.Base != expectedBase {
		return simulation.RuntimeIdentity{}, fmt.Errorf("d2legacy: runtime package set does not identify the built-in base")
	}

	if err := simulation.ValidateGameDataGenerationID(gameDataGenerationID); err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	authoritativeDigest, err := hashSource(source, "lua/d2legacy", func(name string) bool {
		return strings.HasSuffix(name, ".lua")
	})
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	configurationDigest, err := hashConfiguration(configuration)
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}

	return simulation.RuntimeIdentity{
		Recipe: simulation.RuntimeRecipe{
			Schema:               simulation.RuntimeRecipeSchema,
			EngineAPI:            modcache.EngineAPI,
			NetworkProtocol:      simulation.RuntimeNetworkProtocol,
			AssetSetID:           assetSetID,
			GameDataGenerationID: gameDataGenerationID,
			Packages:             packages,
			AuthoritativeHash:    authoritativeDigest,
			ConfigurationHash:    configurationDigest,
			CapabilityVersions:   authoritativeCapabilityVersions(),
		},
	}, nil
}

// runtimePackageForBuiltin translates the embedded package descriptor into the
// runtime identity form. Keeping this conversion in one place prevents package
// verification from silently omitting metadata that peers compare.
func runtimePackageForBuiltin(builtin modcache.LockedPackage) simulation.RuntimePackage {
	return simulation.RuntimePackage{
		ID:              builtin.Manifest.ID,
		Version:         builtin.Manifest.Version,
		Digest:          builtin.Descriptor.Digest,
		Size:            builtin.Descriptor.Size,
		Redistributable: builtin.Descriptor.Redistributable,
	}
}

// hashConfiguration canonicalizes an omitted configuration as an empty object.
// That convention keeps identity hashes stable for callers that use either no
// variadic argument or an explicit nil map.
func hashConfiguration(configuration []map[string]any) (string, error) {
	configured := map[string]any{}
	if len(configuration) > 0 && configuration[0] != nil {
		configured = configuration[0]
	}

	encoded, err := json.Marshal(configured)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256(encoded)

	return hex.EncodeToString(digest[:]), nil
}

// authoritativeCapabilityVersions lists every module whose implementation can
// affect deterministic state. Changing this map intentionally changes runtime
// identity, preventing mismatched peers from starting a session together.
func authoritativeCapabilityVersions() map[string]string {
	return map[string]string{
		"d2legacy.map_catalog":     "v1",
		"d2legacy.movement_rules":  "v1",
		"d2legacy.quest_catalog":   "v1",
		"engine.authority_command": "v1",
		"engine.authority_random":  "v1",
		"engine.authority_state":   "v1",
		"engine.animdata":          "v1",
		"engine.data":              "v1",
		"engine.deterministic":     "v1",
		"engine.ecs":               "v1",
		"engine.initial_data":      "v1",
		"engine.records":           "v1",
		"engine.worldgen":          "v1",
	}
}

// hashSource hashes selected files by sorted path and canonical contents. The
// explicit path separators ensure that different file boundaries cannot produce
// the same byte stream, while line-ending normalization keeps checkouts portable.
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

	// Filesystem traversal order is not guaranteed, so sort before hashing to
	// make the identity reproducible across archive and operating-system readers.
	sort.Strings(names)

	hash := sha256.New()

	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return "", err
		}

		data = modcache.CanonicalBuiltinSource(name, data)
		_, _ = hash.Write([]byte(name)) // Hash writes cannot fail for sha256.
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
		"d2legacy.combat.damage.roll",
		"d2legacy.combat.basic_melee.hit",
		"d2legacy.combat.basic_melee.damage",
		"d2legacy.skill.aura.corpse_chance",
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
