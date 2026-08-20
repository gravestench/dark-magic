package modcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Lock struct {
	Schema   string          `json:"schema"`
	Digest   string          `json:"digest"`
	Packages []LockedPackage `json:"packages"`
}

// EmptyLock returns the canonical digest-bearing lock for a session with no
// extensions, allowing empty and non-empty recipes to use one identity format.
func EmptyLock() Lock {
	lock := Lock{Schema: LockSchema}
	lock.Digest, _ = lockDigest(lock)

	return lock
}

type LockedPackage struct {
	Descriptor Descriptor `json:"descriptor"`
	Manifest   Manifest   `json:"manifest"`
}

// ResolvedSet is the complete product package set: one immutable built-in base
// supplied by the distribution and an ordered lock of cached extensions.
type ResolvedSet struct {
	Base       LockedPackage `json:"base"`
	Extensions Lock          `json:"extensions"`
}

// ResolveExact reconstructs an authenticated session lock from exact package
// descriptors. It does not consult the profile's one-version-per-ID selection,
// so simultaneous sessions may safely pin different content-addressed versions.
func (store *Store) ResolveExact(descriptors []Descriptor, base LockedPackage) (ResolvedSet, error) {
	if err := validateLockedPackage(base); err != nil || base.Manifest.Kind != "game" {
		return ResolvedSet{}, errors.New("modcache: invalid built-in base package")
	}

	packages := make([]LockedPackage, 0, len(descriptors))
	for _, descriptor := range descriptors {
		manifest, err := store.verifyDescriptor(descriptor)
		if err != nil {
			return ResolvedSet{}, err
		}

		pkg := LockedPackage{Descriptor: descriptor, Manifest: manifest}
		if err := validateLockedPackage(pkg); err != nil || manifest.Kind != "extension" {
			return ResolvedSet{}, fmt.Errorf("modcache: exact package %q differs from its descriptor", descriptor.ID)
		}

		packages = append(packages, pkg)
	}

	return resolvedSet(base, packages)
}

// Packages returns a new activation-order slice containing the base followed by
// extensions. Callers may reorder or replace those copied package structs, but
// nested manifest slices remain shared and must be treated as read-only.
func (set ResolvedSet) Packages() []LockedPackage {
	result := make([]LockedPackage, 0, 1+len(set.Extensions.Packages))
	result = append(result, set.Base)
	result = append(result, set.Extensions.Packages...)

	return result
}

// ClientComponents returns base and extension entrypoints in activation order so
// dependency components start before dependent components.
func (set ResolvedSet) ClientComponents() []string {
	result := append([]string(nil), set.Base.Manifest.Entrypoints.ClientComponents...)
	return append(result, set.Extensions.ClientComponents()...)
}

// AuthorityComponents returns base and extension entrypoints in activation order
// for deterministic authority composition.
func (set ResolvedSet) AuthorityComponents() []string {
	result := append([]string(nil), set.Base.Manifest.Entrypoints.AuthorityComponents...)
	return append(result, set.Extensions.AuthorityComponents()...)
}

// Resolve expands a profile through verified dependency manifests and emits a
// deterministic dependency-first lock. Disabled corrupt packages are never read.
func (store *Store) Resolve(profile Profile, base LockedPackage) (ResolvedSet, error) {
	if err := ValidateProfile(profile); err != nil {
		return ResolvedSet{}, err
	}

	if err := validateLockedPackage(base); err != nil || base.Manifest.Kind != "game" {
		return ResolvedSet{}, fmt.Errorf("modcache: invalid built-in base package")
	}

	catalog, err := store.readIndex()
	if err != nil {
		return ResolvedSet{}, err
	}

	resolver := profileResolver{
		store:     store,
		base:      base,
		catalog:   catalog,
		manifests: make(map[string]Manifest, len(catalog.Packages)),
		state:     make(map[string]uint8),
	}

	for _, id := range profile.Enabled {
		if id == base.Manifest.ID {
			return ResolvedSet{}, fmt.Errorf(
				"modcache: built-in base %q is always enabled and cannot appear in the extension profile",
				id,
			)
		}

		if err := resolver.visit(id); err != nil {
			return ResolvedSet{}, err
		}
	}

	return resolvedSet(base, resolver.packages)
}

type profileResolver struct {
	store     *Store
	base      LockedPackage
	catalog   index
	manifests map[string]Manifest
	state     map[string]uint8
	packages  []LockedPackage
}

// load verifies and caches one selected manifest. The resolver calls it only for
// enabled or required IDs, so disabled corrupt packages cannot block startup.
func (resolver *profileResolver) load(id string) (Manifest, error) {
	if manifest, found := resolver.manifests[id]; found {
		return manifest, nil
	}

	descriptor, found := resolver.catalog.Packages[id]
	if !found {
		return Manifest{}, fmt.Errorf("modcache: enabled or required mod %q is not installed", id)
	}

	manifest, err := resolver.store.verifyDescriptor(descriptor)
	if err != nil {
		return Manifest{}, err
	}

	resolver.manifests[id] = manifest

	return manifest, nil
}

// visit performs depth-first dependency resolution. The visiting/complete state
// distinguishes cycles from shared dependencies while append-after-visit yields
// dependency-first activation order.
func (resolver *profileResolver) visit(id string) error {
	if id == resolver.base.Manifest.ID {
		return nil
	}

	switch resolver.state[id] {
	case 1:
		return fmt.Errorf("modcache: dependency cycle includes %q", id)
	case 2:
		return nil
	}

	descriptor := resolver.catalog.Packages[id]

	manifest, err := resolver.load(id)
	if err != nil {
		return err
	}

	if manifest.Kind != "extension" {
		return fmt.Errorf("modcache: cached package %q is not an extension", id)
	}

	resolver.state[id] = 1
	for _, dependency := range manifest.Dependencies {
		if err := resolver.visitDependency(id, dependency); err != nil {
			return err
		}
	}

	resolver.state[id] = 2

	resolver.packages = append(resolver.packages, LockedPackage{
		Descriptor: descriptor,
		Manifest:   manifest,
	})

	return nil
}

// visitDependency validates the base version inline and recursively resolves an
// extension before comparing its installed version to the requirement.
func (resolver *profileResolver) visitDependency(owner string, dependency Dependency) error {
	if dependency.ID == resolver.base.Manifest.ID {
		if dependency.Version != "" && dependency.Version != resolver.base.Manifest.Version {
			return fmt.Errorf(
				"modcache: %s requires %s %s, built-in %s",
				owner,
				dependency.ID,
				dependency.Version,
				resolver.base.Manifest.Version,
			)
		}

		return nil
	}

	if err := resolver.visit(dependency.ID); err != nil {
		return err
	}

	resolvedDependency := resolver.manifests[dependency.ID]
	if dependency.Version != "" && resolvedDependency.Version != dependency.Version {
		return fmt.Errorf(
			"modcache: %s requires %s %s, installed %s",
			owner,
			dependency.ID,
			dependency.Version,
			resolvedDependency.Version,
		)
	}

	return nil
}

// resolvedSet hashes and validates a dependency-ordered package list before
// exposing it, keeping exact and profile-based resolution on one lock contract.
func resolvedSet(base LockedPackage, packages []LockedPackage) (ResolvedSet, error) {
	lock := Lock{Schema: LockSchema, Packages: packages}

	digest, err := lockDigest(lock)
	if err != nil {
		return ResolvedSet{}, err
	}

	lock.Digest = digest
	if err := ValidateLock(lock, base); err != nil {
		return ResolvedSet{}, err
	}

	return ResolvedSet{Base: base, Extensions: lock}, nil
}

// ValidateLock proves that package metadata and activation order still match
// the exact identity emitted by Resolve. Blob verification remains Store's job
// because a lock intentionally contains no local filesystem paths.
func ValidateLock(lock Lock, base LockedPackage) error {
	if lock.Schema != LockSchema || !validDigest(lock.Digest) {
		return errors.New("modcache: invalid lock")
	}

	want, err := lockDigest(lock)
	if err != nil || want != lock.Digest {
		return errors.New("modcache: lock digest mismatch")
	}

	seen := make(map[string]struct{}, len(lock.Packages))
	positions := make(map[string]int, len(lock.Packages))

	packageIDs := []string{base.Manifest.ID}
	for index, pkg := range lock.Packages {
		if err := validateLockedPackage(pkg); err != nil || pkg.Manifest.Kind != "extension" {
			return fmt.Errorf("modcache: invalid locked extension %q", pkg.Manifest.ID)
		}

		if _, duplicate := seen[pkg.Manifest.ID]; duplicate {
			return fmt.Errorf("modcache: duplicate locked package %q", pkg.Manifest.ID)
		}

		seen[pkg.Manifest.ID] = struct{}{}
		positions[pkg.Manifest.ID] = index
		packageIDs = append(packageIDs, pkg.Manifest.ID)
	}

	for index, id := range packageIDs {
		for _, other := range packageIDs[index+1:] {
			if strings.HasPrefix(id, other+".") || strings.HasPrefix(other, id+".") {
				return fmt.Errorf("modcache: package namespaces %q and %q overlap", id, other)
			}
		}
	}

	for index, pkg := range lock.Packages {
		for _, dependency := range pkg.Manifest.Dependencies {
			if dependency.ID == base.Manifest.ID {
				if dependency.Version != "" && dependency.Version != base.Manifest.Version {
					return fmt.Errorf("modcache: extension %q requires incompatible built-in base", pkg.Manifest.ID)
				}

				continue
			}

			position, found := positions[dependency.ID]
			if !found || position >= index {
				return fmt.Errorf(
					"modcache: extension %q dependency %q is missing or ordered after it",
					pkg.Manifest.ID,
					dependency.ID,
				)
			}

			resolved := lock.Packages[position].Manifest
			if dependency.Version != "" && dependency.Version != resolved.Version {
				return fmt.Errorf("modcache: extension %q dependency %q version differs", pkg.Manifest.ID, dependency.ID)
			}
		}
	}

	return nil
}

// validateLockedPackage proves descriptor and manifest metadata describe the
// same immutable package before either is trusted by a session lock.
func validateLockedPackage(pkg LockedPackage) error {
	if err := ValidateManifest(pkg.Manifest); err != nil {
		return err
	}

	if pkg.Descriptor.ID != pkg.Manifest.ID || pkg.Descriptor.Version != pkg.Manifest.Version ||
		pkg.Descriptor.Redistributable != pkg.Manifest.Redistributable || pkg.Descriptor.Size <= 0 ||
		!validDigest(pkg.Descriptor.Digest) {
		return fmt.Errorf("modcache: invalid locked package %q", pkg.Manifest.ID)
	}

	return nil
}

// lockDigest hashes canonical JSON with the digest field cleared, making the
// lock self-identifying without recursively hashing its prior digest.
func lockDigest(lock Lock) (string, error) {
	lock.Digest = ""

	data, err := json.Marshal(lock)
	if err != nil {
		return "", errors.New("modcache: encode resolved lock")
	}

	digest := sha256.Sum256(data)

	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

// ClientComponents flattens extension entrypoints in locked activation order.
func (lock Lock) ClientComponents() []string {
	var result []string
	for _, pkg := range lock.Packages {
		result = append(result, pkg.Manifest.Entrypoints.ClientComponents...)
	}

	return result
}

// AuthorityComponents flattens extension entrypoints in locked activation order.
func (lock Lock) AuthorityComponents() []string {
	var result []string
	for _, pkg := range lock.Packages {
		result = append(result, pkg.Manifest.Entrypoints.AuthorityComponents...)
	}

	return result
}
