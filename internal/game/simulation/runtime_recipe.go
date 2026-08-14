package simulation

import (
	"errors"
	"fmt"
	"strings"
)

const (
	RuntimeRecipeSchema    = "dark-magic.runtime-recipe/v1"
	RuntimeNetworkProtocol = "dark-magic.game-session/v2"
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
	Schema             string            `json:"schema"`
	EngineAPI          string            `json:"engine_api"`
	NetworkProtocol    string            `json:"network_protocol"`
	Packages           RuntimePackageSet `json:"packages"`
	AuthoritativeHash  string            `json:"authoritative_hash"`
	ConfigurationHash  string            `json:"configuration_hash"`
	CapabilityVersions map[string]string `json:"capability_versions,omitempty"`
}

func (recipe RuntimeRecipe) Validate() error {
	if recipe.Schema != RuntimeRecipeSchema || strings.TrimSpace(recipe.EngineAPI) == "" ||
		strings.TrimSpace(recipe.NetworkProtocol) == "" || strings.TrimSpace(recipe.AuthoritativeHash) == "" ||
		strings.TrimSpace(recipe.ConfigurationHash) == "" {
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
