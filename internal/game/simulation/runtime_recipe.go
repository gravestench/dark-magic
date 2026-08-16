package simulation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	RuntimeRecipeSchema      = "dark-magic.runtime-recipe/v2"
	RuntimeNetworkProtocol   = "dark-magic.game-session/v2"
	GameDataGenerationSchema = "dark-magic.game-data-generation/v1"
	RecordParserSchema       = "dark-magic.records-tsv/v1"
	// EmptyAssetSetID is the canonical identity for runtimes with no external
	// mounted game data, such as hermetic unit tests and synthetic fixtures.
	EmptyAssetSetID = "sha256:4b5f42b9b0f48dc738578940d8ca3db3eaac90364a5106411f83012d47998ef6"
)

type RuntimePackage struct {
	ID              string `json:"id"`
	Version         string `json:"version"`
	Digest          string `json:"digest"`
	Size            int64  `json:"size"`
	Redistributable bool   `json:"redistributable"`
}

// RuntimePackageSet is the storage-neutral package input shared by product
// composition. Base is distribution-owned; Extensions are cache-owned and
// ordered dependencies before dependents.
type RuntimePackageSet struct {
	Base       RuntimePackage   `json:"base"`
	Extensions []RuntimePackage `json:"extensions,omitempty"`
}

// RuntimeRecipe is the complete deterministic implementation recipe pinned by
// admission, reconnect, checkpoints, and replays.
type RuntimeRecipe struct {
	Schema               string            `json:"schema"`
	EngineAPI            string            `json:"engine_api"`
	NetworkProtocol      string            `json:"network_protocol"`
	AssetSetID           string            `json:"asset_set_id"`
	GameDataGenerationID string            `json:"game_data_generation_id"`
	Packages             RuntimePackageSet `json:"packages"`
	AuthoritativeHash    string            `json:"authoritative_hash"`
	ConfigurationHash    string            `json:"configuration_hash"`
	CapabilityVersions   map[string]string `json:"capability_versions,omitempty"`
}

func (recipe RuntimeRecipe) Validate() error {
	if recipe.Schema != RuntimeRecipeSchema || strings.TrimSpace(recipe.EngineAPI) == "" ||
		strings.TrimSpace(recipe.NetworkProtocol) == "" || strings.TrimSpace(recipe.AuthoritativeHash) == "" ||
		strings.TrimSpace(recipe.ConfigurationHash) == "" || ValidateAssetSetID(recipe.AssetSetID) != nil ||
		ValidateGameDataGenerationID(recipe.GameDataGenerationID) != nil {
		return errors.New("simulation: invalid runtime recipe contract")
	}
	if err := validateRuntimePackage(recipe.Packages.Base); err != nil {
		return fmt.Errorf("simulation: invalid base package: %w", err)
	}
	seen := map[string]struct{}{recipe.Packages.Base.ID: {}}
	packageIDs := []string{recipe.Packages.Base.ID}
	for _, pkg := range recipe.Packages.Extensions {
		if err := validateRuntimePackage(pkg); err != nil {
			return fmt.Errorf("simulation: invalid extension package: %w", err)
		}
		if _, duplicate := seen[pkg.ID]; duplicate {
			return fmt.Errorf("simulation: duplicate runtime package %q", pkg.ID)
		}
		seen[pkg.ID] = struct{}{}
		packageIDs = append(packageIDs, pkg.ID)
	}
	for index, id := range packageIDs {
		for _, other := range packageIDs[index+1:] {
			if strings.HasPrefix(id, other+".") || strings.HasPrefix(other, id+".") {
				return fmt.Errorf("simulation: runtime package namespaces %q and %q overlap", id, other)
			}
		}
	}
	for name, version := range recipe.CapabilityVersions {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(version) == "" {
			return errors.New("simulation: invalid runtime capability version")
		}
	}
	return nil
}

// GameDataGenerationIDForAssetSet binds mounted source bytes to the parser and
// schema contract that turns them into authoritative records. It is distinct
// from AssetSetID so a parser/schema change invalidates deterministic sessions
// even when the legally supplied files are unchanged.
func GameDataGenerationIDForAssetSet(assetSetID string) string {
	payload := GameDataGenerationSchema + "\x00" + assetSetID + "\x00" + RecordParserSchema
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func ValidateGameDataGenerationID(value string) error {
	if !validSHA256Digest(value) {
		return errors.New("simulation: invalid game-data generation identity")
	}
	return nil
}

// ValidateAssetSetID validates the storage-neutral digest shared by product
// composition, Realm allocation, workers, clients, and durable session state.
func ValidateAssetSetID(value string) error {
	if !validSHA256Digest(value) {
		return errors.New("simulation: invalid asset-set identity")
	}
	return nil
}

func validateRuntimePackage(pkg RuntimePackage) error {
	if strings.TrimSpace(pkg.ID) == "" || strings.TrimSpace(pkg.Version) == "" || !validSHA256Digest(pkg.Digest) || pkg.Size <= 0 {
		return errors.New("package ID, version, digest, and positive size are required")
	}
	return nil
}

func validSHA256Digest(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	for _, character := range value[len(prefix):] {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
